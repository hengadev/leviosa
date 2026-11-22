package emailpreview

import (
	"errors"
	"strings"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
)

func TestList_ReturnsEveryBookingEmail(t *testing.T) {
	entries := List()
	if len(entries) != 11 {
		t.Fatalf("List returned %d entries, want 11", len(entries))
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

		switch e.Audience {
		case AudienceClient:
			if !hasLocale(e.Locales, domain.LocaleFR) || !hasLocale(e.Locales, domain.LocaleEN) || len(e.Locales) != 2 {
				t.Errorf("client email %s locales = %v, want fr and en", e.ID, e.Locales)
			}
		case AudiencePartner:
			if len(e.Locales) != 1 || e.Locales[0] != domain.LocaleFR {
				t.Errorf("partner email %s locales = %v, want fr only", e.ID, e.Locales)
			}
		default:
			t.Errorf("entry %s has audience %q", e.ID, e.Audience)
		}
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
