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
