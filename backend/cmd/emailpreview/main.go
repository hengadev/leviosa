// Command emailpreview renders every booking email (issue 022) to HTML
// files with fixed sample data — the seed placeholders (partner, product,
// price, deposit, address) — so the operator can send them to the client
// for review. Offline: no database, KMS, SMTP or network; nothing is sent.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
)

// Sample data, mirroring cmd/seed/seed_data.example.json so the preview
// reads like the real page: partner Camille Martin, Séance d'ostéopathie
// at 70 € with a 20 € deposit, the practice address.
const (
	samplePartnerFirst = "Camille"
	samplePartnerName  = "Camille Martin"
	sampleClientEmail  = "jean.dupont@example.com"
	sampleClientPhone  = "+33612345678"
	sampleProduct      = "Séance d'ostéopathie"
	sampleAddress      = "12 rue des Lilas, 75011 Paris"
	sampleCurrency     = "EUR"
	samplePriceCents   = 7000
	sampleDepositCents = 2000
	sampleManageURL    = "https://leviosa.example.com/booking/t/0123456789abcdef"
	sampleBookAgainFR  = "https://leviosa.example.com/fr"
	sampleBookAgainEN  = "https://leviosa.example.com/en"
)

// A fixed moment, so the preview is reproducible between runs.
var sampleNow = time.Date(2026, time.March, 5, 10, 0, 0, 0, time.UTC)

// preview is one email to write: a file name and the composed email (its
// attachments are listed in the index, not rendered).
type preview struct {
	File  string
	Email domain.ComposedEmail
}

// Run writes every booking email variant to outDir and returns the path of
// the index linking them.
func Run(outDir string) (string, error) {
	previews := allPreviews()

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	indexPath := filepath.Join(outDir, "index.html")
	var rows []string
	for _, p := range previews {
		html, err := render(p.Email)
		if err != nil {
			return "", fmt.Errorf("render %s: %w", p.File, err)
		}
		if err := os.WriteFile(filepath.Join(outDir, p.File), []byte(html), 0o644); err != nil {
			return "", fmt.Errorf("write %s: %w", p.File, err)
		}
		rows = append(rows, indexRow(p))
	}
	if err := os.WriteFile(indexPath, []byte(indexHTML(rows)), 0o644); err != nil {
		return "", fmt.Errorf("write index.html: %w", err)
	}
	return indexPath, nil
}

// allPreviews builds every booking email variant: FR and EN for the
// client-facing ones, French for the partner notices.
func allPreviews() []preview {
	start := time.Date(2026, time.March, 12, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	bookingID := "b7k3xq"
	var previews []preview
	add := func(file string, email domain.ComposedEmail) {
		previews = append(previews, preview{File: file, Email: email})
	}

	for _, locale := range []domain.Locale{domain.LocaleFR, domain.LocaleEN} {
		lang := strings.ToLower(string(locale))

		// Confirmation: paid in full, and with a deposit.
		add("booking_confirmation_full_"+lang+".html", domain.ComposeBookingConfirmation(domain.BookingConfirmationRequest{
			ToEmail: sampleClientEmail, ToFirstName: "Jean", ToLastName: "Dupont", Locale: locale,
			BookingID: bookingID, ProductName: sampleProduct, PartnerName: samplePartnerName, Address: sampleAddress,
			StartTime: start, EndTime: end,
			AmountPaidOnlineCents: samplePriceCents, Currency: sampleCurrency,
			ManageURL: sampleManageURL,
		}, sampleNow))
		add("booking_confirmation_deposit_"+lang+".html", domain.ComposeBookingConfirmation(domain.BookingConfirmationRequest{
			ToEmail: sampleClientEmail, ToFirstName: "Jean", ToLastName: "Dupont", Locale: locale,
			BookingID: bookingID, ProductName: sampleProduct, PartnerName: samplePartnerName, Address: sampleAddress,
			StartTime: start, EndTime: end,
			AmountPaidOnlineCents: sampleDepositCents, BalanceDueCents: samplePriceCents - sampleDepositCents, Currency: sampleCurrency,
			ManageURL: sampleManageURL,
		}, sampleNow))

		add("booking_reminder_"+lang+".html", domain.ComposeBookingReminder(domain.BookingReminderRequest{
			ToEmail: sampleClientEmail, ToFirstName: "Jean", ToLastName: "Dupont", Locale: locale,
			BookingID: bookingID, ProductName: sampleProduct, PartnerName: samplePartnerName, Address: sampleAddress,
			StartTime: start, EndTime: end, ManageURL: sampleManageURL,
		}, sampleNow))

		// Client cancellation: full, partial and no refund.
		cancellation := func(file string, refunded int, cancelledBy string) {
			add(file, domain.ComposeBookingClientCancellation(domain.BookingClientCancellationRequest{
				ToEmail: sampleClientEmail, ToFirstName: "Jean", ToLastName: "Dupont", Locale: locale,
				BookingID: bookingID, ProductName: sampleProduct, StartTime: start, EndTime: end,
				CancelledBy:           cancelledBy,
				AmountPaidOnlineCents: samplePriceCents, RefundedAmountCents: refunded, Currency: sampleCurrency,
				BookAgainURL: bookAgainURL(locale),
			}, sampleNow))
		}
		cancellation("booking_cancellation_full_refund_"+lang+".html", samplePriceCents, domain.CancelledByClient)
		cancellation("booking_cancellation_partial_refund_"+lang+".html", sampleDepositCents, domain.CancelledByClient)
		cancellation("booking_cancellation_no_refund_"+lang+".html", 0, domain.CancelledByClient)
		// Practitioner cancellation, client side: full refund.
		cancellation("booking_cancellation_practitioner_"+lang+".html", samplePriceCents, domain.CancelledByPractitioner)

		add("booking_late_payment_refund_"+lang+".html", domain.ComposeLatePaymentRefund(domain.LatePaymentRefundRequest{
			ToEmail: sampleClientEmail, ToFirstName: "Jean", ToLastName: "Dupont", Locale: locale,
			BookingID: bookingID, ProductName: sampleProduct, StartTime: start, EndTime: end,
			RefundedAmountCents: samplePriceCents, Currency: sampleCurrency,
			BookAgainURL: bookAgainURL(locale),
		}, sampleNow))
	}

	// Partner notices are French.
	add("booking_partner_notice_fr.html", domain.ComposeBookingPartnerNotice(domain.BookingPartnerNoticeRequest{
		ToEmail: "camille.martin@example.com", ToFirstName: samplePartnerFirst,
		BookingID: bookingID, ProductName: sampleProduct,
		ClientName: "Jean Dupont", ClientEmail: sampleClientEmail, ClientPhone: sampleClientPhone,
		StartTime: start, EndTime: end,
	}, sampleNow))
	add("booking_partner_cancellation_fr.html", domain.ComposeBookingPartnerCancellation(domain.BookingPartnerCancellationRequest{
		ToEmail: "camille.martin@example.com", ToFirstName: samplePartnerFirst,
		BookingID: bookingID, ProductName: sampleProduct, ClientName: "Jean Dupont",
		StartTime: start, EndTime: end,
		CancelledBy: domain.CancelledByClient, RefundedAmountCents: sampleDepositCents, Currency: sampleCurrency,
	}, sampleNow))

	return previews
}

func bookAgainURL(locale domain.Locale) string {
	if locale == domain.LocaleEN {
		return sampleBookAgainEN
	}
	return sampleBookAgainFR
}

// render produces the HTML exactly as the SMTP client would send it: the
// same embedded template executed with the composed data. Booking templates
// carry their styles inline and reference no external assets.
func render(email domain.ComposedEmail) (string, error) {
	tmpl, err := template.ParseFS(domain.EmailTemplates, "templates/"+email.Template+".html")
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, email.Data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

func indexRow(p preview) string {
	attachments := ""
	if len(p.Email.Attachments) > 0 {
		names := make([]string, 0, len(p.Email.Attachments))
		for _, a := range p.Email.Attachments {
			names = append(names, html.EscapeString(a.Filename))
		}
		attachments = fmt.Sprintf(`<div style="color:#666;">Joint&nbsp;: %s</div>`, strings.Join(names, ", "))
	}
	file := html.EscapeString(p.File)
	return fmt.Sprintf(`<li style="margin-bottom:14px;"><a href="%s">%s</a> — %s%s</li>`,
		file, file, html.EscapeString(p.Email.Subject), attachments)
}

func indexHTML(rows []string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
  <head>
    <meta charset="UTF-8" />
    <title>Aperçu des emails — Leviosa</title>
  </head>
  <body style="font-family: Arial, sans-serif; max-width: 720px; margin: 40px auto;">
    <h1>Aperçu des emails de réservation</h1>
    <p>Données d'exemple (seed) — Camille Martin, %s, %s. Rien n'est envoyé.</p>
    <ul style="list-style: none; padding: 0;">
      %s
    </ul>
  </body>
</html>
`, sampleProduct, sampleAddress, strings.Join(rows, "\n      "))
}

func main() {
	outDir := flag.String("out", "tmp/email-preview", "output directory (created if missing)")
	flag.Parse()

	indexPath, err := Run(*outDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Email previews written to %s\nOpen %s\n", *outDir, indexPath)
}
