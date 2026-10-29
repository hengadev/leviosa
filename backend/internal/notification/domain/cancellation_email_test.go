package domain

import (
	"strings"
	"testing"
	"time"
)

func testClientCancellationRequest(locale Locale) BookingClientCancellationRequest {
	start := time.Date(2026, 11, 3, 13, 30, 0, 0, time.UTC) // 14:30 in Paris
	return BookingClientCancellationRequest{
		ToEmail:               "client@example.com",
		ToFirstName:           "Alice",
		Locale:                locale,
		BookingID:             "550e8400-e29b-41d4-a716-446655440000",
		ProductName:           "Ostéopathie",
		StartTime:             start,
		EndTime:               start.Add(time.Hour),
		CancelledBy:           CancelledByClient,
		AmountPaidOnlineCents: 3000,
		RefundedAmountCents:   1500,
		Currency:              "EUR",
		BookAgainURL:          "https://leviosa.test/",
	}
}

func TestComposeBookingClientCancellation_French(t *testing.T) {
	email := ComposeBookingClientCancellation(testClientCancellationRequest(LocaleFR), composeNow)

	if email.Subject != "Votre rendez-vous est annulé" {
		t.Errorf("subject = %q", email.Subject)
	}
	if email.Template != BookingClientCancellationTemplate {
		t.Errorf("template = %q", email.Template)
	}
	body := renderComposed(t, email)
	for _, want := range []string{
		`lang="fr"`, "Rendez-vous annulé", "Bonjour Alice,", "mardi 3 novembre 2026", "14h30 – 15h30",
		"est bien annulé", "15,00 € vous sont remboursés sur les 30,00 € payés en ligne",
		"5 à 10 jours ouvrés", "Réserver un autre créneau", "550e8400-e29b-41d4-a716-446655440000",
	} {
		assertContains(t, body, want)
	}
}

func TestComposeBookingClientCancellation_English(t *testing.T) {
	email := ComposeBookingClientCancellation(testClientCancellationRequest(LocaleEN), composeNow)

	if email.Subject != "Your appointment is cancelled" {
		t.Errorf("subject = %q", email.Subject)
	}
	body := renderComposed(t, email)
	for _, want := range []string{
		`lang="en"`, "Appointment cancelled", "Hello Alice,", "Tuesday 3 November 2026", "14:30 – 15:30",
		"€15.00 of the €30.00 paid online is refunded", "Book another slot",
	} {
		assertContains(t, body, want)
	}
}

func TestComposeBookingClientCancellation_RefundCases(t *testing.T) {
	tests := []struct {
		name      string
		locale    Locale
		paid      int
		refunded  int
		by        string
		want      string
		wantDelay bool
	}{
		{"full refund", LocaleFR, 3000, 3000, CancelledByClient, "Votre paiement de 30,00 € vous est intégralement remboursé.", true},
		{"no refund", LocaleFR, 3000, 0, CancelledByClient, "ne donne pas lieu à un remboursement", false},
		{"no refund, English", LocaleEN, 3000, 0, CancelledByClient, "this cancellation is not refunded", false},
		{"nothing paid", LocaleEN, 0, 0, CancelledByClient, "Nothing was paid online", false},
		{"practitioner", LocaleFR, 3000, 3000, CancelledByPractitioner, "Votre praticien a dû annuler", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testClientCancellationRequest(tt.locale)
			req.AmountPaidOnlineCents, req.RefundedAmountCents, req.CancelledBy = tt.paid, tt.refunded, tt.by
			body := renderComposed(t, ComposeBookingClientCancellation(req, composeNow))
			assertContains(t, body, tt.want)
			delay := strings.Contains(body, "5 à 10 jours ouvrés") || strings.Contains(body, "5 to 10 business days")
			if delay != tt.wantDelay {
				t.Errorf("refund delay note shown = %v, want %v", delay, tt.wantDelay)
			}
		})
	}
}

func TestComposeBookingPartnerCancellation(t *testing.T) {
	start := time.Date(2026, 11, 3, 13, 30, 0, 0, time.UTC)
	email := ComposeBookingPartnerCancellation(BookingPartnerCancellationRequest{
		ToEmail:             "partner@example.com",
		ToFirstName:         "Bob",
		BookingID:           "550e8400-e29b-41d4-a716-446655440000",
		ProductName:         "Ostéopathie",
		ClientName:          "Alice Martin",
		StartTime:           start,
		EndTime:             start.Add(time.Hour),
		CancelledBy:         CancelledByClient,
		RefundedAmountCents: 1500,
		Currency:            "EUR",
	}, composeNow)

	if email.Subject != "Réservation annulée : Ostéopathie, mardi 3 novembre 2026 à 14h30" {
		t.Errorf("subject = %q", email.Subject)
	}
	body := renderComposed(t, email)
	for _, want := range []string{
		`lang="fr"`, "Réservation annulée", "Bonjour Bob,", "annulée par le client", "Alice Martin",
		"14h30 – 15h30", "Remboursé au client", "15,00 €",
	} {
		assertContains(t, body, want)
	}
}
