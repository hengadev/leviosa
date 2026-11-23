package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// PractitionerBookingInput is a booking the practitioner makes for a client
// who called (ADR-0021): a product, a free slot of one of the practitioner's
// availability occurrences, and the client, either an existing user
// (ClientID) or a guest entered by hand (first name, last name and phone;
// email optional). Exactly one of the two.
type PractitionerBookingInput struct {
	AvailabilityID uuid.UUID  `json:"availability_id"`
	ProductID      uuid.UUID  `json:"product_id"`
	SlotStartTime  time.Time  `json:"slot_start_time"`
	ClientID       *uuid.UUID `json:"client_id,omitempty"`
	GuestFirstName string     `json:"guest_first_name,omitempty"`
	GuestLastName  string     `json:"guest_last_name,omitempty"`
	GuestPhone     string     `json:"guest_phone,omitempty"`
	GuestEmail     string     `json:"guest_email,omitempty"`
}

// Normalize trims the guest fields.
func (in *PractitionerBookingInput) Normalize() {
	in.GuestFirstName = strings.TrimSpace(in.GuestFirstName)
	in.GuestLastName = strings.TrimSpace(in.GuestLastName)
	in.GuestPhone = strings.TrimSpace(in.GuestPhone)
	in.GuestEmail = strings.TrimSpace(in.GuestEmail)
}

// ClientProblem names what is wrong with the input's client, or "" when it
// is valid: an existing client and guest fields are exclusive, and a guest
// needs a first name, a last name and a phone number.
func (in PractitionerBookingInput) ClientProblem() string {
	hasGuest := in.GuestFirstName != "" || in.GuestLastName != "" || in.GuestPhone != "" || in.GuestEmail != ""
	if in.ClientID != nil {
		if *in.ClientID == uuid.Nil {
			return "client_id: must not be empty"
		}
		if hasGuest {
			return "client_id: an existing client takes no guest fields"
		}
		return ""
	}
	switch {
	case in.GuestFirstName == "":
		return "guest_first_name: required"
	case in.GuestLastName == "":
		return "guest_last_name: required"
	case in.GuestPhone == "":
		return "guest_phone: required"
	}
	return ""
}

// PractitionerProduct is a product the practitioner can book for a client
// who called: a published product they offer, with its active price.
type PractitionerProduct struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	DurationMinutes int       `json:"duration_minutes"`
	PriceCents      int       `json:"price_cents"`
	Currency        string    `json:"currency"`
}

// IsOpenForPractitioner reports whether the practitioner can still book in
// the availability at now: open (not deleted or blocked) and not over. Unlike
// IsAvailable, a one-off availability that has started but not ended still
// is: the practitioner books without the minimum notice, so a later slot of
// today's occurrence can be taken (the slot engine checks the slot itself).
func (a *Availability) IsOpenForPractitioner(now time.Time) bool {
	return a.Status == AvailabilityStatusAvailable && !a.SeriesEndedBefore(now)
}
