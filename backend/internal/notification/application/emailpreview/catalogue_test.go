package emailpreview

import (
	"errors"
	"strings"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
)

// wantLocales is the language each email outside the booking flow is sent
// in today; every other client email exists in French and English.
var wantLocales = map[string][]domain.Locale{
	"payment_receipt":       englishOnly,
	"payment_failed":        englishOnly,
	"contact_form":          frenchOnly,
	"payment_dispute_alert": englishOnly,
}

func TestList_ReturnsEveryEmail(t *testing.T) {
	entries := List()
	if len(entries) != 15 {
		t.Fatalf("List returned %d entries, want 15", len(entries))
	}

	seen := map[string]bool{}
	for _, e := range entries {
		if e.ID == "" || e.Label == "" {
			t.Errorf("entry %+v has an empty id or label", e)
		}
		if seen[e.ID] {
			t.Errorf("duplicate id %q", e.ID)
		}
		seen[e.ID] = true

		if want, ok := wantLocales[e.ID]; ok {
			if len(e.Locales) != 1 || e.Locales[0] != want[0] {
				t.Errorf("%s locales = %v, want %v", e.ID, e.Locales, want)
			}
		}

		switch e.Audience {
		case AudienceClient:
			if _, ok := wantLocales[e.ID]; ok {
				break
			}
			if !hasLocale(e.Locales, domain.LocaleFR) || !hasLocale(e.Locales, domain.LocaleEN) || len(e.Locales) != 2 {
				t.Errorf("client email %s locales = %v, want fr and en", e.ID, e.Locales)
			}
		case AudiencePartner:
			if len(e.Locales) != 1 || e.Locales[0] != domain.LocaleFR {
				t.Errorf("partner email %s locales = %v, want fr only", e.ID, e.Locales)
			}
		case AudiencePractice:
			if _, ok := wantLocales[e.ID]; !ok {
				t.Errorf("practice email %s has no expected locale", e.ID)
			}
		default:
			t.Errorf("entry %s has audience %q", e.ID, e.Audience)
		}
	}

	for id := range wantLocales {
		if !seen[id] {
			t.Errorf("List misses %s", id)
		}
	}
}

func TestList_OtherEmailsAudiencesAndLabels(t *testing.T) {
	want := map[string]struct {
		label    string
		audience Audience
	}{
		"payment_receipt":       {"Reçu de paiement", AudienceClient},
		"payment_failed":        {"Échec du paiement", AudienceClient},
		"contact_form":          {"Message du formulaire de contact", AudiencePractice},
		"payment_dispute_alert": {"Litige de paiement", AudiencePractice},
	}
	for _, e := range List() {
		w, ok := want[e.ID]
		if !ok {
			continue
		}
		if e.Label != w.label || e.Audience != w.audience {
			t.Errorf("%s = (%q, %s), want (%q, %s)", e.ID, e.Label, e.Audience, w.label, w.audience)
		}
	}
}

func TestRender_OtherEmailsShowTheirSampleData(t *testing.T) {
	cases := []struct {
		id      string
		locale  domain.Locale
		subject string
		body    []string
	}{
		// The receipt template shows no payment detail today, only the year.
		{"payment_receipt", domain.LocaleEN, "Payment Confirmation", []string{"&copy; 2026"}},
		{"payment_failed", domain.LocaleEN, "Payment Failed", []string{sampleAmount, sampleRetryURL}},
		{"contact_form", domain.LocaleFR, "Nouveau message de contact — Marie Leroy", []string{sampleVisitorName, sampleVisitorEmail, "samedi matin"}},
		{"payment_dispute_alert", domain.LocaleEN, "Payment dispute opened — booking " + sampleBookingID, []string{sampleDisputeReason, sampleClientName}},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			r, err := Render(tc.id, tc.locale)
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if r.Subject != tc.subject {
				t.Errorf("subject = %q, want %q", r.Subject, tc.subject)
			}
			for _, want := range tc.body {
				if !strings.Contains(r.HTML, want) {
					t.Errorf("HTML misses %q", want)
				}
			}
			if len(r.Attachments) != 0 {
				t.Errorf("attachments = %v, want none", r.Attachments)
			}
		})
	}
}

func TestList_ReturnsACopy(t *testing.T) {
	List()[0].Locales[0] = "xx"
	if List()[0].Locales[0] != domain.LocaleFR {
		t.Fatal("List exposed the catalogue's locales slice")
	}
}

func TestRender_EveryEntryAndLocale(t *testing.T) {
	for _, e := range List() {
		for _, locale := range e.Locales {
			t.Run(e.ID+"_"+string(locale), func(t *testing.T) {
				r, err := Render(e.ID, locale)
				if err != nil {
					t.Fatalf("Render: %v", err)
				}
				if strings.TrimSpace(r.Subject) == "" {
					t.Error("empty subject")
				}
				if !strings.Contains(r.HTML, "<html") {
					t.Errorf("HTML does not look like an email: %.80q", r.HTML)
				}
			})
		}
	}
}

func TestRender_ShowsTheSampleDoorCode(t *testing.T) {
	for _, id := range []string{"booking_confirmation_full", "booking_confirmation_deposit", "booking_reminder", "booking_access_update"} {
		for _, locale := range bothLocales {
			r, err := Render(id, locale)
			if err != nil {
				t.Fatalf("Render(%s, %s): %v", id, locale, err)
			}
			if !strings.Contains(r.HTML, sampleDoorCode) {
				t.Errorf("%s (%s) does not show the sample door code", id, locale)
			}
		}
	}
}

func TestRender_ConfirmationListsItsCalendarAttachment(t *testing.T) {
	r, err := Render("booking_confirmation_full", domain.LocaleFR)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(r.Attachments) != 1 || !strings.HasSuffix(r.Attachments[0], ".ics") {
		t.Fatalf("attachments = %v, want one .ics file", r.Attachments)
	}
}

func TestRender_NotFound(t *testing.T) {
	cases := []struct {
		name   string
		id     string
		locale domain.Locale
	}{
		{"unknown id", "no_such_email", domain.LocaleFR},
		{"unsupported locale", "booking_reminder", "de"},
		{"partner email in English", "booking_partner_notice", domain.LocaleEN},
		{"English-only email in French", "payment_receipt", domain.LocaleFR},
		{"French-only practice email in English", "contact_form", domain.LocaleEN},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Render(tc.id, tc.locale)
			if !errors.Is(err, errs.ErrDomainNotFound) {
				t.Fatalf("err = %v, want a not-found error", err)
			}
		})
	}
}
