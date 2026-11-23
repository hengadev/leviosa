package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Client-file grouping (issue 042): nothing is stored, everything is computed
// from the practitioner's bookings. Bookings with a client account group by
// that account; guest bookings group by normalised phone (E.164), or by
// lower-cased email when there's no phone. A guest whose phone matches a
// registered client's stays a separate group (no merging in v1).

// NormalizePhoneE164 normalises a phone number for grouping: separators
// (spaces, dots, dashes) are stripped, a leading 0 French number becomes
// +33…, and a 00 international prefix becomes +.
func NormalizePhoneE164(phone string) string {
	var b strings.Builder
	for _, r := range phone {
		switch r {
		case ' ', '.', '-', '(', ')':
			continue
		}
		b.WriteRune(r)
	}
	p := b.String()
	switch {
	case strings.HasPrefix(p, "00"):
		return "+" + p[2:]
	case strings.HasPrefix(p, "0"):
		return "+33" + p[1:]
	default:
		return p
	}
}

// ClientGroupKey returns the grouping key of a booking: "client:<uuid>" when
// the client has an account, else "phone:<E.164>", else "email:<lowercased>".
func ClientGroupKey(b PartnerBookingResponse) string {
	if b.ClientID != nil {
		return "client:" + b.ClientID.String()
	}
	if p := NormalizePhoneE164(b.ClientPhone); p != "" {
		return "phone:" + p
	}
	return "email:" + strings.ToLower(b.ClientEmail)
}

// ClientGroup is one client of the practitioner: their contact details and
// all their bookings with this practitioner.
type ClientGroup struct {
	Key       string
	ClientID  *uuid.UUID
	Name      string
	FirstName string
	LastName  string
	Phone     string
	Email     string
	Bookings  []PartnerBookingResponse
}

// GroupClientBookings groups the practitioner's bookings by client. Bookings
// that are not real appointments (awaiting payment or expired) are excluded;
// groups keep first-seen order.
func GroupClientBookings(bookings []PartnerBookingResponse) []ClientGroup {
	var order []string
	byKey := map[string]*ClientGroup{}
	for _, b := range bookings {
		if b.Status == BookingStatusAwaitingPayment || b.Status == BookingStatusExpired {
			continue
		}
		key := ClientGroupKey(b)
		g, ok := byKey[key]
		if !ok {
			g = &ClientGroup{Key: key}
			byKey[key] = g
			order = append(order, key)
		}
		if g.ClientID == nil {
			g.ClientID = b.ClientID
		}
		if g.FirstName == "" {
			g.FirstName = b.GuestFirstName
		}
		if g.LastName == "" {
			g.LastName = b.GuestLastName
		}
		if g.Name == "" {
			g.Name = b.ClientName
		}
		if g.Phone == "" {
			g.Phone = b.ClientPhone
		}
		if g.Email == "" {
			g.Email = b.ClientEmail
		}
		g.Bookings = append(g.Bookings, b)
	}
	groups := make([]ClientGroup, 0, len(order))
	for _, key := range order {
		groups = append(groups, *byKey[key])
	}
	return groups
}

// ClientFileBooking is one session in the client file.
type ClientFileBooking struct {
	ID            uuid.UUID     `json:"id"`
	SlotStartTime time.Time     `json:"slot_start_time"`
	SlotEndTime   time.Time     `json:"slot_end_time"`
	ProductName   string        `json:"product_name"`
	Status        BookingStatus `json:"status"`
	PaymentStatus PaymentStatus `json:"payment_status"`
}

// ClientFileResponse is the client file: contact details plus upcoming
// bookings (soonest first) and past ones (latest first). Cancelled bookings
// are listed and marked by their status. No notes, ever (ADR-0019).
type ClientFileResponse struct {
	Name     string              `json:"name"`
	Phone    string              `json:"phone,omitempty"`
	Email    string              `json:"email,omitempty"`
	Upcoming []ClientFileBooking `json:"upcoming"`
	Past     []ClientFileBooking `json:"past"`
}

// BuildClientFile splits a client group into the client file: sessions still
// running or in the future are upcoming (soonest first), finished ones are
// past (latest first).
func BuildClientFile(g ClientGroup, now time.Time) ClientFileResponse {
	file := ClientFileResponse{
		Name:     g.Name,
		Phone:    g.Phone,
		Email:    g.Email,
		Upcoming: []ClientFileBooking{},
		Past:     []ClientFileBooking{},
	}
	for _, b := range g.Bookings {
		entry := ClientFileBooking{
			ID:            b.ID,
			SlotStartTime: b.SlotStartTime,
			SlotEndTime:   b.SlotEndTime,
			ProductName:   b.ProductName,
			Status:        b.Status,
			PaymentStatus: b.PaymentStatus,
		}
		if b.SlotEndTime.After(now) {
			file.Upcoming = append(file.Upcoming, entry)
		} else {
			file.Past = append(file.Past, entry)
		}
	}
	sortClientBookings(file.Upcoming, true)
	sortClientBookings(file.Past, false)
	return file
}

func sortClientBookings(bs []ClientFileBooking, ascending bool) {
	for i := 1; i < len(bs); i++ {
		for j := i; j > 0; j-- {
			a, b := bs[j-1].SlotStartTime, bs[j].SlotStartTime
			swap := ascending && b.Before(a) || !ascending && b.After(a)
			if !swap {
				break
			}
			bs[j-1], bs[j] = bs[j], bs[j-1]
		}
	}
}

// ClientSummary is one distinct client of the practitioner, for the client
// search (issue 043 wires it into the booking form). ClientID is set when the
// group is a client account (the form then books by id); FirstName/LastName
// come from the bookings' guest fields, or are split from the name for
// accounts, so a guest can be pre-filled.
type ClientSummary struct {
	ClientID  *uuid.UUID `json:"client_id,omitempty"`
	Name      string     `json:"name"`
	FirstName string     `json:"first_name,omitempty"`
	LastName  string     `json:"last_name,omitempty"`
	Phone     string     `json:"phone,omitempty"`
	Email     string     `json:"email,omitempty"`
}

// SearchClientSummaries returns the distinct clients whose name or phone
// matches the query: a name fragment matches case-insensitively, a phone
// fragment matches on separator-stripped digits. An empty query returns
// every client.
func SearchClientSummaries(groups []ClientGroup, query string) []ClientSummary {
	q := strings.ToLower(strings.TrimSpace(query))
	qDigits := NormalizePhoneE164(query)
	clients := make([]ClientSummary, 0, len(groups))
	for _, g := range groups {
		if q != "" {
			nameMatch := strings.Contains(strings.ToLower(g.Name), q)
			phoneMatch := qDigits != "" && strings.Contains(NormalizePhoneE164(g.Phone), qDigits)
			if !nameMatch && !phoneMatch {
				continue
			}
		}
		firstName, lastName := g.FirstName, g.LastName
		if firstName == "" && lastName == "" {
			// An account group has no guest fields: split its display name
			// (first word as first name, the rest as last name).
			firstName, lastName = splitName(g.Name)
		}
		clients = append(clients, ClientSummary{
			ClientID:  g.ClientID,
			Name:      g.Name,
			FirstName: firstName,
			LastName:  lastName,
			Phone:     g.Phone,
			Email:     g.Email,
		})
	}
	return clients
}

func splitName(name string) (string, string) {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}
