package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewBookingPush_FirstNameInitialProductAndParisTime(t *testing.T) {
	// 09:00 UTC on Saturday 28 November 2026 is 10:00 in Paris.
	start := time.Date(2026, 11, 28, 9, 0, 0, 0, time.UTC)
	id := uuid.New()

	msg := NewBookingPush(BookingPushDetails{
		BookingID:       id,
		ClientFirstName: "Marie",
		ClientLastName:  "Dupont",
		ProductName:     "Ostéopathie",
		Start:           start,
		Now:             time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC),
	})

	if msg.Title != "Nouvelle réservation" {
		t.Errorf("title = %q", msg.Title)
	}
	if want := "Marie D. · Ostéopathie · sam. 28 nov. 10:00"; msg.Body != want {
		t.Errorf("body = %q, want %q", msg.Body, want)
	}
	if strings.Contains(msg.Title+msg.Body, "Dupont") {
		t.Errorf("the full last name leaked: %+v", msg)
	}
	if msg.URL != PractitionerBookingsURL {
		t.Errorf("url = %q, want Réservations for another day", msg.URL)
	}
	if msg.Tag != "new-booking:"+id.String() {
		t.Errorf("tag = %q", msg.Tag)
	}
}

func TestNewSessionReminderPush_TodayURL(t *testing.T) {
	// 09:00 UTC on Saturday 28 November 2026 is 10:00 in Paris.
	start := time.Date(2026, 11, 28, 9, 0, 0, 0, time.UTC)
	id := uuid.New()

	msg := NewSessionReminderPush(BookingPushDetails{
		BookingID:       id,
		ClientFirstName: "Marie",
		ClientLastName:  "Dupont",
		ProductName:     "Séance ostéopathie",
		Start:           start,
		Now:             time.Date(2026, 11, 28, 8, 45, 0, 0, time.UTC),
	})

	if msg.Title != "Dans 15 min" {
		t.Errorf("title = %q", msg.Title)
	}
	if want := "Marie D. · Séance ostéopathie · sam. 28 nov. 10:00"; msg.Body != want {
		t.Errorf("body = %q, want %q", msg.Body, want)
	}
	// The session is imminent: always Aujourd'hui, even across midnight in Paris.
	if msg.URL != PractitionerTodayURL {
		t.Errorf("url = %q, want %q", msg.URL, PractitionerTodayURL)
	}
	if msg.Tag != "session-reminder:"+id.String() {
		t.Errorf("tag = %q", msg.Tag)
	}
}

func TestPractitionerBookingURL_TodayInParis(t *testing.T) {
	// 23:30 UTC on 27 November is already 28 November in Paris.
	now := time.Date(2026, 11, 27, 23, 30, 0, 0, time.UTC)
	today := time.Date(2026, 11, 28, 16, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 11, 29, 8, 0, 0, 0, time.UTC)

	if got := PractitionerBookingURL(today, now); got != PractitionerTodayURL {
		t.Errorf("today: %q, want %q", got, PractitionerTodayURL)
	}
	if got := PractitionerBookingURL(tomorrow, now); got != PractitionerBookingsURL {
		t.Errorf("tomorrow: %q, want %q", got, PractitionerBookingsURL)
	}
}

func TestClientShortName(t *testing.T) {
	cases := []struct{ first, last, want string }{
		{"Marie", "Dupont", "Marie D."},
		{" Jean ", "  émile-durand", "Jean É."},
		{"Paul", "", "Paul"},
		{"", "Martin", "M."},
		{"", "", ""},
	}
	for _, c := range cases {
		if got := ClientShortName(c.first, c.last); got != c.want {
			t.Errorf("ClientShortName(%q, %q) = %q, want %q", c.first, c.last, got, c.want)
		}
	}
}

func TestFormatPushTime_SummerTime(t *testing.T) {
	// 12:15 UTC on Monday 1 June 2026 is 14:15 in Paris (UTC+2).
	if got := FormatPushTime(time.Date(2026, 6, 1, 12, 15, 0, 0, time.UTC)); got != "lun. 1 juin 14:15" {
		t.Errorf("got %q", got)
	}
}

func TestCancellationPush_BodyAndRéservationsURL(t *testing.T) {
	id := uuid.New()
	msg := CancellationPush(BookingPushDetails{
		BookingID:       id,
		ClientFirstName: "Marie",
		ClientLastName:  "Dupont",
		ProductName:     "Ostéopathie",
		Start:           time.Date(2026, 11, 28, 9, 0, 0, 0, time.UTC),
		Now:             time.Date(2026, 11, 28, 8, 0, 0, 0, time.UTC),
	})
	if msg.Title != "Annulation" {
		t.Errorf("title = %q", msg.Title)
	}
	if want := "Marie D. · Ostéopathie · sam. 28 nov. 10:00"; msg.Body != want {
		t.Errorf("body = %q, want %q", msg.Body, want)
	}
	// Always Réservations, even for a booking today: the today page shows
	// what is coming, not what was cancelled.
	if msg.URL != PractitionerBookingsURL {
		t.Errorf("url = %q, want %q", msg.URL, PractitionerBookingsURL)
	}
	if msg.Tag != "cancellation:"+id.String() {
		t.Errorf("tag = %q", msg.Tag)
	}
}

func TestPaymentProblemPushes_ShortNameOnlyAndComptabilitéURL(t *testing.T) {
	d := BookingPushDetails{
		BookingID:       uuid.New(),
		ClientFirstName: "Marie",
		ClientLastName:  "Dupont",
		ProductName:     "Ostéopathie",
		Start:           time.Date(2026, 11, 28, 9, 0, 0, 0, time.UTC),
	}
	for name, msg := range map[string]PushMessage{
		"dispute":     DisputePush(d),
		"late refund": LatePaymentRefundPush(d),
	} {
		if msg.Body != "Marie D." {
			t.Errorf("%s: body = %q, want the short name only", name, msg.Body)
		}
		if msg.URL != PractitionerPaymentsURL {
			t.Errorf("%s: url = %q, want %q", name, msg.URL, PractitionerPaymentsURL)
		}
		if strings.Contains(msg.Title+msg.Body, "Dupont") || strings.Contains(msg.Title+msg.Body, "€") {
			t.Errorf("%s: payload leaked more than the short name: %+v", name, msg)
		}
	}
	if got := DisputePush(d).Title; got != "Litige de paiement" {
		t.Errorf("dispute title = %q", got)
	}
	if got := LatePaymentRefundPush(d).Title; got != "Paiement tardif remboursé" {
		t.Errorf("late refund title = %q", got)
	}
}
