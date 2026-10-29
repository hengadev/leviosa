package domain

import (
	"strings"
	"testing"
	"time"
)

func testLatePaymentRefundRequest(locale Locale) LatePaymentRefundRequest {
	start := time.Date(2026, 11, 3, 13, 30, 0, 0, time.UTC) // 14:30 in Paris
	return LatePaymentRefundRequest{
		ToEmail:             "client@example.com",
		ToFirstName:         "Alice",
		Locale:              locale,
		BookingID:           "550e8400-e29b-41d4-a716-446655440000",
		ProductName:         "Ostéopathie",
		StartTime:           start,
		EndTime:             start.Add(time.Hour),
		RefundedAmountCents: 2000,
		Currency:            "EUR",
		BookAgainURL:        "https://leviosa.test/",
	}
}

func TestComposeLatePaymentRefund_French(t *testing.T) {
	email := ComposeLatePaymentRefund(testLatePaymentRefundRequest(LocaleFR), composeNow)

	if email.Subject != "Votre paiement a été remboursé" {
		t.Errorf("subject = %q", email.Subject)
	}
	if len(email.Attachments) != 0 {
		t.Errorf("expected no attachment, got %d", len(email.Attachments))
	}
	body := renderComposed(t, email)
	for _, want := range []string{
		`lang="fr"`, "Paiement remboursé", "Bonjour Alice,", "mardi 3 novembre 2026", "14h30 – 15h30",
		"n&#39;était plus disponible", "intégralement remboursé", "20,00\u00a0€",
		"Choisir un autre créneau", "https://leviosa.test/", "550e8400-e29b-41d4-a716-446655440000",
	} {
		assertContains(t, body, want)
	}
}

func TestComposeLatePaymentRefund_English(t *testing.T) {
	email := ComposeLatePaymentRefund(testLatePaymentRefundRequest(LocaleEN), composeNow)

	if email.Subject != "Your payment has been refunded" {
		t.Errorf("subject = %q", email.Subject)
	}
	body := renderComposed(t, email)
	for _, want := range []string{
		`lang="en"`, "Payment refunded", "Hello Alice,", "Tuesday 3 November 2026", "14:30 – 15:30",
		"no longer available", "refunded in full", "€20.00", "Pick another slot",
	} {
		assertContains(t, body, want)
	}
}

func TestComposeLatePaymentRefund_WithoutBookAgainLink(t *testing.T) {
	req := testLatePaymentRefundRequest(LocaleEN)
	req.BookAgainURL = ""
	body := renderComposed(t, ComposeLatePaymentRefund(req, composeNow))
	if strings.Contains(body, "Pick another slot") {
		t.Error("no book-again link expected without a URL")
	}
}
