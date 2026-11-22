package domain

import (
	"strings"
	"testing"
	"time"
)

func testAccessUpdateRequest(locale Locale) BookingAccessUpdateRequest {
	start := time.Date(2026, 11, 3, 13, 30, 0, 0, time.UTC) // 14:30 in Paris
	return BookingAccessUpdateRequest{
		ToEmail:            "client@example.com",
		ToFirstName:        "Alice",
		ToLastName:         "Smith",
		Locale:             locale,
		BookingID:          "550e8400-e29b-41d4-a716-446655440000",
		ProductName:        "Massage",
		PartnerName:        "Bob Jones",
		Address:            "12 rue de la Paix, 75002 Paris",
		StartTime:          start,
		EndTime:            start.Add(time.Hour),
		ManageURL:          "https://leviosa.test/bookings?token=abc",
		DoorCode:           testDoorCode,
		AccessInstructions: testInstructions,
	}
}

func TestComposeBookingAccessUpdate_French(t *testing.T) {
	email := ComposeBookingAccessUpdate(testAccessUpdateRequest(LocaleFR), composeNow)

	if email.Subject != "Nouveau code d'accès pour votre rendez-vous du mardi 3 novembre 2026" {
		t.Errorf("subject = %q", email.Subject)
	}
	if email.Template != BookingAccessUpdateTemplate {
		t.Errorf("template = %q", email.Template)
	}
	if len(email.Attachments) != 0 {
		t.Errorf("the access update has no attachment, got %d", len(email.Attachments))
	}
	body := renderComposed(t, email)
	for _, want := range []string{
		`lang="fr"`, "Nouveau code d&#39;accès", "Bonjour Alice,", "Le code d&#39;accès du cabinet a changé",
		"Massage", "mardi 3 novembre 2026", "14h30 – 15h30", "Bob Jones", "12 rue de la Paix, 75002 Paris",
		"Accès au cabinet", "Indications", testDoorCode,
		"Sonnez à l&#39;interphone « Cabinet ».<br />2e étage, porte de gauche.",
		"Gérer ma réservation", "https://leviosa.test/bookings?token=abc",
		"550e8400-e29b-41d4-a716-446655440000", "Tous droits réservés.",
	} {
		assertContains(t, body, want)
	}
}

func TestComposeBookingAccessUpdate_English(t *testing.T) {
	email := ComposeBookingAccessUpdate(testAccessUpdateRequest(LocaleEN), composeNow)

	if email.Subject != "New access code for your appointment on Tuesday 3 November 2026" {
		t.Errorf("subject = %q", email.Subject)
	}
	if len(email.Attachments) != 0 {
		t.Errorf("the access update has no attachment, got %d", len(email.Attachments))
	}
	body := renderComposed(t, email)
	for _, want := range []string{
		`lang="en"`, "New access code", "Hello Alice,", "The practice access code has changed",
		"Treatment", "Tuesday 3 November 2026", "14:30 – 15:30", "Bob Jones", "12 rue de la Paix, 75002 Paris",
		"Getting in", "Door code", "Directions", testDoorCode,
		"Manage my booking", "https://leviosa.test/bookings?token=abc", "All rights reserved.",
	} {
		assertContains(t, body, want)
	}
	for _, french := range []string{"Bonjour", "Rendez-vous", "Gérer"} {
		if strings.Contains(body, french) {
			t.Errorf("English access update contains French %q", french)
		}
	}
}

func TestComposeBookingAccessUpdate_MissingLocaleIsFrench(t *testing.T) {
	email := ComposeBookingAccessUpdate(testAccessUpdateRequest(""), composeNow)
	assertContains(t, renderComposed(t, email), `lang="fr"`)
}

func TestComposeBookingAccessUpdate_NoManageLinkWithoutToken(t *testing.T) {
	req := testAccessUpdateRequest(LocaleFR)
	req.ManageURL = ""
	body := renderComposed(t, ComposeBookingAccessUpdate(req, composeNow))
	if strings.Contains(body, "Gérer ma réservation") {
		t.Error("no manage button without a manage link")
	}
}
