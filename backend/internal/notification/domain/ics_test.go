package domain

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func testCalendarEvent() CalendarEvent {
	paris := PracticeLocation()
	return CalendarEvent{
		UID:         BookingCalendarUID("550e8400-e29b-41d4-a716-446655440000"),
		Start:       time.Date(2026, 11, 3, 14, 30, 0, 0, paris), // winter time, UTC+1
		End:         time.Date(2026, 11, 3, 15, 30, 0, 0, paris),
		Summary:     "Massage – Bob Jones",
		Location:    "12 rue de la Paix, 75002 Paris",
		Description: "Gérer ou annuler votre réservation : https://leviosa.test/bookings?token=abc",
		URL:         "https://leviosa.test/bookings?token=abc",
		Stamp:       time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC),
	}
}

// unfold reverses RFC 5545 line folding.
func unfold(ics string) string {
	return strings.ReplaceAll(ics, "\r\n ", "")
}

func icsProperty(t *testing.T, ics, name string) string {
	t.Helper()
	for _, line := range strings.Split(unfold(ics), "\r\n") {
		if strings.HasPrefix(line, name+":") {
			return strings.TrimPrefix(line, name+":")
		}
	}
	t.Fatalf("property %s not found in:\n%s", name, ics)
	return ""
}

func TestCalendarEventICS_TimesInUTC(t *testing.T) {
	ics := string(testCalendarEvent().ICS())

	// 14:30 in Paris on 3 November (UTC+1) is 13:30 UTC.
	if got := icsProperty(t, ics, "DTSTART"); got != "20261103T133000Z" {
		t.Errorf("DTSTART = %q, want 20261103T133000Z", got)
	}
	if got := icsProperty(t, ics, "DTEND"); got != "20261103T143000Z" {
		t.Errorf("DTEND = %q, want 20261103T143000Z", got)
	}
	if got := icsProperty(t, ics, "DTSTAMP"); got != "20261029T090000Z" {
		t.Errorf("DTSTAMP = %q", got)
	}

	// Summer time (UTC+2): 10:00 in Paris on 1 July is 08:00 UTC.
	summer := testCalendarEvent()
	summer.Start = time.Date(2026, 7, 1, 10, 0, 0, 0, PracticeLocation())
	summer.End = summer.Start.Add(time.Hour)
	if got := icsProperty(t, string(summer.ICS()), "DTSTART"); got != "20260701T080000Z" {
		t.Errorf("summer DTSTART = %q, want 20260701T080000Z", got)
	}
}

func TestCalendarEventICS_Structure(t *testing.T) {
	ics := string(testCalendarEvent().ICS())

	if !strings.HasPrefix(ics, "BEGIN:VCALENDAR\r\nVERSION:2.0\r\n") {
		t.Errorf("unexpected calendar header: %q", ics[:40])
	}
	if !strings.HasSuffix(ics, "END:VEVENT\r\nEND:VCALENDAR\r\n") {
		t.Error("calendar must end with END:VEVENT and END:VCALENDAR, CRLF terminated")
	}
	if strings.Contains(strings.ReplaceAll(ics, "\r\n", ""), "\n") {
		t.Error("every line must end with CRLF, never a bare LF")
	}
	for _, name := range []string{"PRODID", "UID", "SUMMARY", "LOCATION", "DESCRIPTION", "URL"} {
		icsProperty(t, ics, name)
	}
	if got := icsProperty(t, ics, "UID"); got != "booking-550e8400-e29b-41d4-a716-446655440000@leviosa" {
		t.Errorf("UID = %q", got)
	}
}

func TestCalendarEventICS_UIDStablePerBooking(t *testing.T) {
	first := testCalendarEvent()
	second := testCalendarEvent()
	second.Stamp = second.Stamp.Add(48 * time.Hour)
	if icsProperty(t, string(first.ICS()), "UID") != icsProperty(t, string(second.ICS()), "UID") {
		t.Error("the same booking must always get the same UID")
	}
	if BookingCalendarUID("a") == BookingCalendarUID("b") {
		t.Error("two bookings must not share a UID")
	}
}

func TestCalendarEventICS_EscapesText(t *testing.T) {
	event := testCalendarEvent()
	event.Summary = `Soin; visage, \ corps`
	event.Location = "Cabinet\nÉtage 2"
	ics := string(event.ICS())

	if got := icsProperty(t, ics, "SUMMARY"); got != `Soin\; visage\, \\ corps` {
		t.Errorf("SUMMARY = %q", got)
	}
	if got := icsProperty(t, ics, "LOCATION"); got != `Cabinet\nÉtage 2` {
		t.Errorf("LOCATION = %q", got)
	}
	// The URI value is not escaped.
	if got := icsProperty(t, ics, "URL"); got != "https://leviosa.test/bookings?token=abc" {
		t.Errorf("URL = %q", got)
	}
}

func TestCalendarEventICS_FoldsLongLines(t *testing.T) {
	event := testCalendarEvent()
	longURL := "https://leviosa.test/bookings?token=" + strings.Repeat("x", 120)
	event.Description = "Gérer ou annuler votre réservation : " + longURL
	event.URL = longURL
	ics := string(event.ICS())

	for _, line := range strings.Split(strings.TrimSuffix(ics, "\r\n"), "\r\n") {
		if len(line) > 75 {
			t.Errorf("line longer than 75 octets (%d): %q", len(line), line)
		}
		if !utf8.ValidString(line) {
			t.Errorf("folding split a UTF-8 character: %q", line)
		}
	}
	if !strings.Contains(ics, "\r\n ") {
		t.Fatal("expected folded continuation lines")
	}
	if got := icsProperty(t, ics, "URL"); got != longURL {
		t.Errorf("unfolded URL = %q, want %q", got, longURL)
	}
	if got := icsProperty(t, ics, "DESCRIPTION"); got != `Gérer ou annuler votre réservation : `+longURL {
		t.Errorf("unfolded DESCRIPTION = %q", got)
	}
}

func TestFoldICSLine_MultiByteBoundary(t *testing.T) {
	line := "SUMMARY:" + strings.Repeat("é", 80) // 2 octets each
	folded := foldICSLine(line)
	for _, part := range strings.Split(folded, "\r\n") {
		if len(part) > 75 {
			t.Errorf("folded line too long: %d octets", len(part))
		}
		if !utf8.ValidString(part) {
			t.Errorf("folded line splits a character: %q", part)
		}
	}
	if unfold(folded) != line {
		t.Error("unfolding must give back the original line")
	}
	if foldICSLine("SHORT:line") != "SHORT:line" {
		t.Error("short lines are not folded")
	}
}
