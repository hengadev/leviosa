package domain

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func testConfirmationRequest(locale Locale) BookingConfirmationRequest {
	start := time.Date(2026, 11, 3, 13, 30, 0, 0, time.UTC) // 14:30 in Paris
	return BookingConfirmationRequest{
		ToEmail:               "client@example.com",
		ToFirstName:           "Alice",
		ToLastName:            "Smith",
		Locale:                locale,
		BookingID:             "550e8400-e29b-41d4-a716-446655440000",
		ProductName:           "Massage",
		PartnerName:           "Bob Jones",
		Address:               "12 rue de la Paix, 75002 Paris",
		StartTime:             start,
		EndTime:               start.Add(time.Hour),
		AmountPaidOnlineCents: 2000,
		BalanceDueCents:       4500,
		Currency:              "EUR",
		ManageURL:             "https://leviosa.test/bookings?token=abc",
	}
}

var composeNow = time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)

func renderComposed(t *testing.T, email ComposedEmail) string {
	t.Helper()
	tmpl := parseTemplateOrFail(t, email.Template)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, email.Data); err != nil {
		t.Fatalf("execute %s: %v", email.Template, err)
	}
	return buf.String()
}

func TestComposeBookingConfirmation_French(t *testing.T) {
	email := ComposeBookingConfirmation(testConfirmationRequest(LocaleFR), composeNow)

	if email.Subject != "Votre rendez-vous du mardi 3 novembre 2026 est confirmé" {
		t.Errorf("subject = %q", email.Subject)
	}
	body := renderComposed(t, email)
	for _, want := range []string{
		`lang="fr"`, "Rendez-vous confirmé", "Bonjour Alice,", "Massage", "mardi 3 novembre 2026",
		"14h30 – 15h30", "12 rue de la Paix, 75002 Paris", "Bob Jones",
		"Payé en ligne", "20,00 €", "Reste à régler sur place", "45,00 €",
		"Gérer ma réservation", "https://leviosa.test/bookings?token=abc",
	} {
		assertContains(t, body, want)
	}

	if len(email.Attachments) != 1 {
		t.Fatalf("expected one .ics attachment, got %d", len(email.Attachments))
	}
	att := email.Attachments[0]
	if att.Filename != "rendez-vous.ics" || !strings.HasPrefix(att.ContentType, "text/calendar") {
		t.Errorf("unexpected attachment %q (%q)", att.Filename, att.ContentType)
	}
	ics := unfold(string(att.Data))
	assertContains(t, ics, "DTSTART:20261103T133000Z")
	assertContains(t, ics, "LOCATION:12 rue de la Paix\\, 75002 Paris")
	assertContains(t, ics, "DESCRIPTION:Gérer ou annuler votre réservation : https://leviosa.test/bookings?token=abc")
}

func TestComposeBookingConfirmation_English(t *testing.T) {
	email := ComposeBookingConfirmation(testConfirmationRequest(LocaleEN), composeNow)

	if email.Subject != "Your appointment on Tuesday 3 November 2026 is confirmed" {
		t.Errorf("subject = %q", email.Subject)
	}
	body := renderComposed(t, email)
	for _, want := range []string{
		`lang="en"`, "Appointment confirmed", "Hello Alice,", "Tuesday 3 November 2026", "14:30 – 15:30",
		"Paid online", "€20.00", "Balance due at the practice", "€45.00", "Manage my booking",
	} {
		assertContains(t, body, want)
	}
	if strings.Contains(body, "Reste à régler") {
		t.Error("an English email must not contain French copy")
	}
	if email.Attachments[0].Filename != "appointment.ics" {
		t.Errorf("attachment filename = %q", email.Attachments[0].Filename)
	}
	assertContains(t, unfold(string(email.Attachments[0].Data)), "DESCRIPTION:Manage or cancel your booking: https://leviosa.test/bookings?token=abc")
}

func TestComposeBookingConfirmation_MissingLocaleIsFrench(t *testing.T) {
	email := ComposeBookingConfirmation(testConfirmationRequest(""), composeNow)
	assertContains(t, renderComposed(t, email), "Rendez-vous confirmé")
}

func TestComposeBookingConfirmation_NoBalanceWhenPaidInFull(t *testing.T) {
	for _, locale := range []Locale{LocaleFR, LocaleEN} {
		req := testConfirmationRequest(locale)
		req.AmountPaidOnlineCents = 6500
		req.BalanceDueCents = 0
		body := renderComposed(t, ComposeBookingConfirmation(req, composeNow))
		for _, label := range []string{"Reste à régler sur place", "Balance due at the practice"} {
			if strings.Contains(body, label) {
				t.Errorf("%s: balance row must be omitted when nothing is due, found %q", locale, label)
			}
		}
	}
}

func TestComposeBookingConfirmation_NoManageLinkWithoutToken(t *testing.T) {
	req := testConfirmationRequest(LocaleFR)
	req.ManageURL = ""
	email := ComposeBookingConfirmation(req, composeNow)
	if strings.Contains(renderComposed(t, email), "Gérer ma réservation") {
		t.Error("the manage button needs a manage URL")
	}
	ics := string(email.Attachments[0].Data)
	if strings.Contains(ics, "DESCRIPTION:") || strings.Contains(ics, "URL:") {
		t.Error("no manage link in the calendar file without a manage URL")
	}
}

func TestComposeBookingPartnerNotice_French(t *testing.T) {
	start := time.Date(2026, 11, 3, 13, 30, 0, 0, time.UTC)
	email := ComposeBookingPartnerNotice(BookingPartnerNoticeRequest{
		ToEmail:     "partner@example.com",
		ToFirstName: "Bob",
		BookingID:   "550e8400-e29b-41d4-a716-446655440000",
		ProductName: "Massage",
		ClientName:  "Alice Smith",
		ClientEmail: "alice@example.com",
		ClientPhone: "+33612345678",
		StartTime:   start,
		EndTime:     start.Add(time.Hour),
	}, composeNow)

	if email.Subject != "Nouvelle réservation : Massage, mardi 3 novembre 2026 à 14h30" {
		t.Errorf("subject = %q", email.Subject)
	}
	if len(email.Attachments) != 0 {
		t.Error("the partner notice has no attachment")
	}
	body := renderComposed(t, email)
	for _, want := range []string{
		"Nouvelle réservation", "Bonjour Bob,", "Alice Smith", "alice@example.com", "33612345678", // html/template writes "+" as &#43;
		"Massage", "mardi 3 novembre 2026", "14h30 – 15h30",
	} {
		assertContains(t, body, want)
	}
}

func TestComposeBookingPartnerNotice_OmitsMissingPhone(t *testing.T) {
	email := ComposeBookingPartnerNotice(BookingPartnerNoticeRequest{
		ToEmail: "partner@example.com", ProductName: "Massage", ClientName: "Alice Smith", ClientEmail: "alice@example.com",
		StartTime: composeNow, EndTime: composeNow.Add(time.Hour),
	}, composeNow)
	if strings.Contains(renderComposed(t, email), "Téléphone") {
		t.Error("the phone row must not render without a phone number")
	}
}

func TestFormatDateAndTime_PracticeTimeZone(t *testing.T) {
	// 23:30 UTC on 31 December is already 1 January in Paris.
	instant := time.Date(2026, 12, 31, 23, 30, 0, 0, time.UTC)
	if got := FormatDate(instant, LocaleFR); got != "vendredi 1 janvier 2027" {
		t.Errorf("FormatDate FR = %q", got)
	}
	if got := FormatDate(instant, LocaleEN); got != "Friday 1 January 2027" {
		t.Errorf("FormatDate EN = %q", got)
	}
	if got := FormatTime(instant, LocaleFR); got != "00h30" {
		t.Errorf("FormatTime FR = %q", got)
	}
	if got := FormatTime(instant, LocaleEN); got != "00:30" {
		t.Errorf("FormatTime EN = %q", got)
	}
}

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		cents    int
		currency string
		locale   Locale
		want     string
	}{
		{4500, "EUR", LocaleFR, "45,00 €"},
		{4500, "EUR", LocaleEN, "€45.00"},
		{5, "eur", LocaleEN, "€0.05"},
		{1999, "CHF", LocaleFR, "19,99 CHF"},
		{1999, "CHF", LocaleEN, "19.99 CHF"},
	}
	for _, tt := range tests {
		if got := FormatMoney(tt.cents, tt.currency, tt.locale); got != tt.want {
			t.Errorf("FormatMoney(%d, %q, %q) = %q, want %q", tt.cents, tt.currency, tt.locale, got, tt.want)
		}
	}
}

func TestParseLocale(t *testing.T) {
	for input, want := range map[string]Locale{"": LocaleFR, "fr": LocaleFR, "en": LocaleEN, "EN": LocaleEN, "de": LocaleFR} {
		if got := ParseLocale(input); got != want {
			t.Errorf("ParseLocale(%q) = %q, want %q", input, got, want)
		}
	}
}
