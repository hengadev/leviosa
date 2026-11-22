// Package emailpreview is the catalogue of every previewable email: its
// sample data and how to render it. The email preview command and the admin
// preview endpoints both read it. Each entry composes with the real Compose*
// function and renders with the same templates as the SMTP client, so a
// preview cannot drift from what is sent.
//
// Pure: no database, settings, KMS, network or clock, and never real data —
// only the fixed sample below.
package emailpreview

import (
	"errors"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
)

// Sample data, mirroring cmd/seed/seed_data.example.json so the preview
// reads like the real page: partner Camille Martin, Séance d'ostéopathie
// at 70 € with a 20 € deposit, the practice address.
const (
	samplePartnerFirst = "Camille"
	samplePartnerName  = "Camille Martin"
	samplePartnerEmail = "camille.martin@example.com"
	sampleClientFirst  = "Jean"
	sampleClientLast   = "Dupont"
	sampleClientName   = "Jean Dupont"
	sampleClientEmail  = "jean.dupont@example.com"
	sampleClientPhone  = "+33612345678"
	sampleBookingID    = "b7k3xq"
	sampleCurrency     = "EUR"
	samplePriceCents   = 7000
	sampleDepositCents = 2000
	sampleManageURL    = "https://leviosa.example.com/bookings?token=0123456789abcdef"
	sampleDoorCode     = "4821B"
	sampleInstructions = "Sonnez à l'interphone « Cabinet Leviosa ».\n2e étage, porte de gauche."
	sampleBookAgainFR  = "https://leviosa.example.com/fr"
	sampleBookAgainEN  = "https://leviosa.example.com/en"

	// SampleProduct and SampleAddress are the sample booking's product and
	// practice address, shown in the preview command's index.
	SampleProduct = "Séance d'ostéopathie"
	SampleAddress = "12 rue des Lilas, 75011 Paris"
)

// Fixed moments, so a preview is the same on every run.
var (
	sampleNow   = time.Date(2026, time.March, 5, 10, 0, 0, 0, time.UTC)
	sampleStart = time.Date(2026, time.March, 12, 9, 0, 0, 0, time.UTC)
	sampleEnd   = sampleStart.Add(time.Hour)
)

// Audience is who receives an email.
type Audience string

const (
	AudienceClient   Audience = "client"
	AudiencePartner  Audience = "partner"
	AudiencePractice Audience = "practice"
)

// Entry describes one previewable email.
type Entry struct {
	// ID is stable, e.g. "booking_confirmation_full".
	ID string
	// Label is the French name shown to the administrator.
	Label    string
	Audience Audience
	// Locales are the languages the email exists in.
	Locales []domain.Locale
}

// Rendered is one email rendered with the sample data.
type Rendered struct {
	Subject string
	HTML    string
	// Attachments are the attachment file names.
	Attachments []string
}

// ErrUnknownPreview is wrapped in the not-found error Render returns for an
// unknown id or a locale the email does not exist in.
var ErrUnknownPreview = errors.New("unknown email preview")

type catalogueEntry struct {
	Entry
	compose func(domain.Locale) domain.ComposedEmail
}

var (
	bothLocales = []domain.Locale{domain.LocaleFR, domain.LocaleEN}
	frenchOnly  = []domain.Locale{domain.LocaleFR}
)

var catalogue = []catalogueEntry{
	{
		Entry: Entry{ID: "booking_confirmation_full", Label: "Confirmation — paiement complet", Audience: AudienceClient, Locales: bothLocales},
		compose: func(locale domain.Locale) domain.ComposedEmail {
			return confirmation(locale, samplePriceCents, 0)
		},
	},
	{
		Entry: Entry{ID: "booking_confirmation_deposit", Label: "Confirmation — acompte", Audience: AudienceClient, Locales: bothLocales},
		compose: func(locale domain.Locale) domain.ComposedEmail {
			return confirmation(locale, sampleDepositCents, samplePriceCents-sampleDepositCents)
		},
	},
	{
		Entry: Entry{ID: "booking_reminder", Label: "Rappel de rendez-vous", Audience: AudienceClient, Locales: bothLocales},
		compose: func(locale domain.Locale) domain.ComposedEmail {
			return domain.ComposeBookingReminder(domain.BookingReminderRequest{
				ToEmail: sampleClientEmail, ToFirstName: sampleClientFirst, ToLastName: sampleClientLast, Locale: locale,
				BookingID: sampleBookingID, ProductName: SampleProduct, PartnerName: samplePartnerName, Address: SampleAddress,
				StartTime: sampleStart, EndTime: sampleEnd, ManageURL: sampleManageURL,
				DoorCode: sampleDoorCode, AccessInstructions: sampleInstructions,
			}, sampleNow)
		},
	},
	{
		// The door code changed (or the access was sent again).
		Entry: Entry{ID: "booking_access_update", Label: "Mise à jour des informations d'accès", Audience: AudienceClient, Locales: bothLocales},
		compose: func(locale domain.Locale) domain.ComposedEmail {
			return domain.ComposeBookingAccessUpdate(domain.BookingAccessUpdateRequest{
				ToEmail: sampleClientEmail, ToFirstName: sampleClientFirst, ToLastName: sampleClientLast, Locale: locale,
				BookingID: sampleBookingID, ProductName: SampleProduct, PartnerName: samplePartnerName, Address: SampleAddress,
				StartTime: sampleStart, EndTime: sampleEnd, ManageURL: sampleManageURL,
				DoorCode: sampleDoorCode, AccessInstructions: sampleInstructions,
			}, sampleNow)
		},
	},
	{
		Entry: Entry{ID: "booking_cancellation_full_refund", Label: "Annulation — remboursement complet", Audience: AudienceClient, Locales: bothLocales},
		compose: func(locale domain.Locale) domain.ComposedEmail {
			return clientCancellation(locale, samplePriceCents, domain.CancelledByClient)
		},
	},
	{
		Entry: Entry{ID: "booking_cancellation_partial_refund", Label: "Annulation — remboursement partiel", Audience: AudienceClient, Locales: bothLocales},
		compose: func(locale domain.Locale) domain.ComposedEmail {
			return clientCancellation(locale, sampleDepositCents, domain.CancelledByClient)
		},
	},
	{
		Entry: Entry{ID: "booking_cancellation_no_refund", Label: "Annulation — sans remboursement", Audience: AudienceClient, Locales: bothLocales},
		compose: func(locale domain.Locale) domain.ComposedEmail {
			return clientCancellation(locale, 0, domain.CancelledByClient)
		},
	},
	{
		// Client side, full refund.
		Entry: Entry{ID: "booking_cancellation_practitioner", Label: "Annulation par le praticien", Audience: AudienceClient, Locales: bothLocales},
		compose: func(locale domain.Locale) domain.ComposedEmail {
			return clientCancellation(locale, samplePriceCents, domain.CancelledByPractitioner)
		},
	},
	{
		Entry: Entry{ID: "booking_late_payment_refund", Label: "Remboursement d'un paiement tardif", Audience: AudienceClient, Locales: bothLocales},
		compose: func(locale domain.Locale) domain.ComposedEmail {
			return domain.ComposeLatePaymentRefund(domain.LatePaymentRefundRequest{
				ToEmail: sampleClientEmail, ToFirstName: sampleClientFirst, ToLastName: sampleClientLast, Locale: locale,
				BookingID: sampleBookingID, ProductName: SampleProduct, StartTime: sampleStart, EndTime: sampleEnd,
				RefundedAmountCents: samplePriceCents, Currency: sampleCurrency,
				BookAgainURL: bookAgainURL(locale),
			}, sampleNow)
		},
	},
	{
		Entry: Entry{ID: "booking_partner_notice", Label: "Praticien — nouvelle réservation", Audience: AudiencePartner, Locales: frenchOnly},
		compose: func(domain.Locale) domain.ComposedEmail {
			return domain.ComposeBookingPartnerNotice(domain.BookingPartnerNoticeRequest{
				ToEmail: samplePartnerEmail, ToFirstName: samplePartnerFirst,
				BookingID: sampleBookingID, ProductName: SampleProduct,
				ClientName: sampleClientName, ClientEmail: sampleClientEmail, ClientPhone: sampleClientPhone,
				StartTime: sampleStart, EndTime: sampleEnd,
			}, sampleNow)
		},
	},
	{
		Entry: Entry{ID: "booking_partner_cancellation", Label: "Praticien — annulation", Audience: AudiencePartner, Locales: frenchOnly},
		compose: func(domain.Locale) domain.ComposedEmail {
			return domain.ComposeBookingPartnerCancellation(domain.BookingPartnerCancellationRequest{
				ToEmail: samplePartnerEmail, ToFirstName: samplePartnerFirst,
				BookingID: sampleBookingID, ProductName: SampleProduct, ClientName: sampleClientName,
				StartTime: sampleStart, EndTime: sampleEnd,
				CancelledBy: domain.CancelledByClient, RefundedAmountCents: sampleDepositCents, Currency: sampleCurrency,
			}, sampleNow)
		},
	},
}

// List returns every previewable email, in a stable order.
func List() []Entry {
	entries := make([]Entry, 0, len(catalogue))
	for _, e := range catalogue {
		entry := e.Entry
		entry.Locales = append([]domain.Locale(nil), e.Locales...)
		entries = append(entries, entry)
	}
	return entries
}

// Render composes and renders the email id in locale with the sample data.
// An unknown id or a locale the email does not exist in is a not-found
// error (errs.ErrDomainNotFound).
func Render(id string, locale domain.Locale) (Rendered, error) {
	for _, e := range catalogue {
		if e.ID != id {
			continue
		}
		if !hasLocale(e.Locales, locale) {
			break
		}
		email := e.compose(locale)
		html, err := domain.RenderEmailTemplate(email.Template, email.Data)
		if err != nil {
			return Rendered{}, err
		}
		attachments := make([]string, 0, len(email.Attachments))
		for _, a := range email.Attachments {
			attachments = append(attachments, a.Filename)
		}
		return Rendered{Subject: email.Subject, HTML: html, Attachments: attachments}, nil
	}
	return Rendered{}, errs.NewNotFoundErr(ErrUnknownPreview, "email preview "+id+" ("+string(locale)+")")
}

func hasLocale(locales []domain.Locale, locale domain.Locale) bool {
	for _, l := range locales {
		if l == locale {
			return true
		}
	}
	return false
}

func confirmation(locale domain.Locale, paidCents, balanceCents int) domain.ComposedEmail {
	return domain.ComposeBookingConfirmation(domain.BookingConfirmationRequest{
		ToEmail: sampleClientEmail, ToFirstName: sampleClientFirst, ToLastName: sampleClientLast, Locale: locale,
		BookingID: sampleBookingID, ProductName: SampleProduct, PartnerName: samplePartnerName, Address: SampleAddress,
		StartTime: sampleStart, EndTime: sampleEnd,
		AmountPaidOnlineCents: paidCents, BalanceDueCents: balanceCents, Currency: sampleCurrency,
		ManageURL: sampleManageURL, DoorCode: sampleDoorCode, AccessInstructions: sampleInstructions,
	}, sampleNow)
}

func clientCancellation(locale domain.Locale, refundedCents int, cancelledBy string) domain.ComposedEmail {
	return domain.ComposeBookingClientCancellation(domain.BookingClientCancellationRequest{
		ToEmail: sampleClientEmail, ToFirstName: sampleClientFirst, ToLastName: sampleClientLast, Locale: locale,
		BookingID: sampleBookingID, ProductName: SampleProduct, StartTime: sampleStart, EndTime: sampleEnd,
		CancelledBy:           cancelledBy,
		AmountPaidOnlineCents: samplePriceCents, RefundedAmountCents: refundedCents, Currency: sampleCurrency,
		BookAgainURL: bookAgainURL(locale),
	}, sampleNow)
}

func bookAgainURL(locale domain.Locale) string {
	if locale == domain.LocaleEN {
		return sampleBookAgainEN
	}
	return sampleBookAgainFR
}
