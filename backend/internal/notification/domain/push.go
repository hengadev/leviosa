package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// PushSubscription is one device on which a user receives Web Push
// notifications (a practitioner notification, ADR-0020). A user may have
// several. The endpoint is a bearer capability (whoever holds it, with the
// VAPID private key, can push to the device), so it is encrypted at rest with
// its keys; its basic hash finds the device again on unsubscribe.
type PushSubscription struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Endpoint  string    `json:"-" encx:"encrypt,hash_basic"`
	P256dh    string    `json:"-" encx:"encrypt"`
	Auth      string    `json:"-" encx:"encrypt"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrPushSubscriptionGone is what a push sender returns when the push
// service answers that the subscription no longer exists (404 or 410): the
// device unsubscribed, or the app was removed. The subscription is deleted.
var ErrPushSubscriptionGone = errors.New("push subscription is gone")

// ErrPushUnavailable refuses a subscription while Web Push is not configured
// (no VAPID keys).
var ErrPushUnavailable = errors.New("push notifications are not available")

// PushMessage is the payload of a practitioner notification, read by the
// service worker (frontend/static/push-sw.js). It never carries the client's
// full name, phone, email, an amount or the practice access (the door code):
// it shows on a lock screen.
type PushMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	// URL is the page the notification opens, relative to the app's origin.
	URL string `json:"url"`
	// Tag lets a later notification about the same booking replace this one.
	Tag string `json:"tag,omitempty"`
}

// Pages a practitioner notification opens.
const (
	PractitionerTodayURL    = "/admin"
	PractitionerBookingsURL = "/staff/agenda/reservations"
)

// BookingPushDetails is what a practitioner notification about a booking
// may show: the client's first name and last-name initial, the product and
// the time. Nothing else is accepted, so nothing else can leak into it.
type BookingPushDetails struct {
	BookingID       uuid.UUID
	ClientFirstName string
	ClientLastName  string
	ProductName     string
	Start           time.Time
	// Now decides the page to open: Aujourd'hui for a booking today
	// (Europe/Paris), Réservations otherwise.
	Now time.Time
}

// NewBookingPush is the "new booking" practitioner notification:
// "Nouvelle réservation" / "Marie D. · Ostéopathie · sam. 28 nov. 10:00".
func NewBookingPush(d BookingPushDetails) PushMessage {
	return bookingPush("Nouvelle réservation", "new-booking", d)
}

// bookingPush builds a practitioner notification about a booking; the next
// events (cancellation, payment problem, reminder) are one call each.
func bookingPush(title, kind string, d BookingPushDetails) PushMessage {
	parts := []string{}
	if name := ClientShortName(d.ClientFirstName, d.ClientLastName); name != "" {
		parts = append(parts, name)
	}
	if product := strings.TrimSpace(d.ProductName); product != "" {
		parts = append(parts, product)
	}
	parts = append(parts, FormatPushTime(d.Start))
	return PushMessage{
		Title: title,
		Body:  strings.Join(parts, " · "),
		URL:   PractitionerBookingURL(d.Start, d.Now),
		Tag:   kind + ":" + d.BookingID.String(),
	}
}

// ClientShortName is the client's first name and last-name initial,
// "Marie D.": all a lock screen may show of who booked.
func ClientShortName(firstName, lastName string) string {
	first := strings.TrimSpace(firstName)
	last := strings.TrimSpace(lastName)
	initial := ""
	for _, r := range last {
		if unicode.IsLetter(r) {
			initial = string(unicode.ToUpper(r)) + "."
			break
		}
	}
	switch {
	case first != "" && initial != "":
		return first + " " + initial
	case first != "":
		return first
	default:
		return initial
	}
}

var (
	frenchShortWeekdays = [...]string{"dim.", "lun.", "mar.", "mer.", "jeu.", "ven.", "sam."}
	frenchShortMonths   = [...]string{"janv.", "févr.", "mars", "avr.", "mai", "juin", "juil.", "août", "sept.", "oct.", "nov.", "déc."}
)

// FormatPushTime renders a time as a notification shows it, in the
// practice's time zone: "sam. 28 nov. 10:00".
func FormatPushTime(t time.Time) string {
	local := t.In(practiceLocation)
	return fmt.Sprintf("%s %d %s %s",
		frenchShortWeekdays[local.Weekday()], local.Day(), frenchShortMonths[local.Month()-1], local.Format("15:04"))
}

// PractitionerBookingURL is the page a notification about a booking opens:
// Aujourd'hui when the booking is today in the practice's time zone,
// Réservations otherwise.
func PractitionerBookingURL(start, now time.Time) string {
	s := start.In(practiceLocation)
	n := now.In(practiceLocation)
	if s.Year() == n.Year() && s.YearDay() == n.YearDay() {
		return PractitionerTodayURL
	}
	return PractitionerBookingsURL
}
