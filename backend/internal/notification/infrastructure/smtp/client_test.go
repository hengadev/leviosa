package smtp

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"

	"gopkg.in/gomail.v2"
)

// capturingClient returns an SMTP client whose messages are written to the
// returned buffer instead of being sent.
func capturingClient(t *testing.T) (*SMTPClient, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	c := NewSMTPClient(SMTPConfig{Host: "smtp.invalid", Port: 587, Username: "noreply@leviosa.test"})
	c.send = func(m *gomail.Message) error {
		buf.Reset()
		_, err := m.WriteTo(&buf)
		return err
	}
	return c, &buf
}

type mimePart struct {
	contentType string
	disposition string
	body        []byte
}

// leafParts walks a MIME tree and returns its non-multipart parts, decoded.
func leafParts(t *testing.T, contentType string, body io.Reader) []mimePart {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("parse content type %q: %v", contentType, err)
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		t.Fatalf("expected a multipart type, got %q", mediaType)
	}
	var parts []mimePart
	reader := multipart.NewReader(body, params["boundary"])
	for {
		p, err := reader.NextRawPart()
		if err == io.EOF {
			return parts
		}
		if err != nil {
			t.Fatalf("read part: %v", err)
		}
		partType := p.Header.Get("Content-Type")
		if strings.HasPrefix(partType, "multipart/") {
			parts = append(parts, leafParts(t, partType, p)...)
			continue
		}
		raw, err := io.ReadAll(p)
		if err != nil {
			t.Fatalf("read part body: %v", err)
		}
		switch p.Header.Get("Content-Transfer-Encoding") {
		case "base64":
			raw, err = base64.StdEncoding.DecodeString(strings.Join(strings.Fields(string(raw)), ""))
			if err != nil {
				t.Fatalf("decode base64 part: %v", err)
			}
		case "quoted-printable":
			raw, err = io.ReadAll(quotedprintable.NewReader(bytes.NewReader(raw)))
			if err != nil {
				t.Fatalf("decode quoted-printable part: %v", err)
			}
		}
		parts = append(parts, mimePart{contentType: partType, disposition: p.Header.Get("Content-Disposition"), body: raw})
	}
}

func TestSendEmail_WithAttachment_IsMultipartMixed(t *testing.T) {
	c, buf := capturingClient(t)
	ics := []byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n")

	err := c.sendEmail(context.Background(), &domain.EmailRequest{
		To:       "client@example.com",
		Subject:  "Test",
		Template: "smtp_check",
		Data:     struct{ SentAt string }{SentAt: "now"},
		Attachments: []domain.EmailAttachment{{
			Filename:    "rendez-vous.ics",
			ContentType: "text/calendar; charset=utf-8; method=PUBLISH",
			Data:        ics,
		}},
	}, "")
	if err != nil {
		t.Fatalf("sendEmail: %v", err)
	}

	msg, err := mail.ReadMessage(buf)
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	topType := msg.Header.Get("Content-Type")
	if !strings.HasPrefix(topType, "multipart/mixed") {
		t.Fatalf("message with an attachment must be multipart/mixed, got %q", topType)
	}

	var attachment *mimePart
	var html *mimePart
	for _, p := range leafParts(t, topType, msg.Body) {
		p := p
		switch {
		case strings.HasPrefix(p.contentType, "text/calendar"):
			attachment = &p
		case strings.HasPrefix(p.contentType, "text/html"):
			html = &p
		}
	}
	if html == nil {
		t.Fatal("expected the HTML body part")
	}
	if attachment == nil {
		t.Fatal("expected a text/calendar attachment part")
	}
	if !strings.Contains(attachment.contentType, "method=PUBLISH") {
		t.Errorf("attachment content type should keep its parameters, got %q", attachment.contentType)
	}
	if !strings.HasPrefix(attachment.disposition, "attachment") || !strings.Contains(attachment.disposition, `filename="rendez-vous.ics"`) {
		t.Errorf("unexpected attachment disposition %q", attachment.disposition)
	}
	if !bytes.Equal(attachment.body, ics) {
		t.Errorf("attachment body = %q, want %q", attachment.body, ics)
	}
}

func TestSendEmail_WithoutAttachment_HasNoAttachmentPart(t *testing.T) {
	c, buf := capturingClient(t)
	if err := c.SendCheckEmail(context.Background(), "client@example.com"); err != nil {
		t.Fatalf("SendCheckEmail: %v", err)
	}
	if strings.Contains(buf.String(), "Content-Disposition: attachment") {
		t.Error("an email without attachments must not contain an attachment part")
	}
}

func TestSendEmail_SenderIsFromElseUsername(t *testing.T) {
	sender := func(c *SMTPClient, buf *bytes.Buffer) string {
		t.Helper()
		if err := c.SendCheckEmail(context.Background(), "client@example.com"); err != nil {
			t.Fatalf("SendCheckEmail: %v", err)
		}
		msg, err := mail.ReadMessage(buf)
		if err != nil {
			t.Fatalf("parse message: %v", err)
		}
		return msg.Header.Get("From")
	}

	c, buf := capturingClient(t)
	if got := sender(c, buf); got != "noreply@leviosa.test" {
		t.Errorf("without From, From = %q, want the SMTP username", got)
	}

	c.config.From = "contact@leviosa.test"
	if got := sender(c, buf); got != "contact@leviosa.test" {
		t.Errorf("From = %q, want the configured From", got)
	}
}

func TestSendBookingConfirmationEmail_LocalizedWithCalendar(t *testing.T) {
	start := time.Date(2026, 11, 3, 13, 30, 0, 0, time.UTC) // 14:30 in Paris
	req := domain.BookingConfirmationRequest{
		ToEmail:               "client@example.com",
		ToFirstName:           "Alice",
		Locale:                domain.LocaleEN,
		BookingID:             "550e8400-e29b-41d4-a716-446655440000",
		ProductName:           "Massage",
		PartnerName:           "Bob Jones",
		Address:               "12 rue de la Paix, 75002 Paris",
		StartTime:             start,
		EndTime:               start.Add(time.Hour),
		AmountPaidOnlineCents: 2000,
		BalanceDueCents:       4000,
		Currency:              "EUR",
		ManageURL:             "https://leviosa.test/bookings?token=tok",
	}

	c, buf := capturingClient(t)
	if err := c.SendBookingConfirmationEmail(context.Background(), req); err != nil {
		t.Fatalf("SendBookingConfirmationEmail: %v", err)
	}

	msg, err := mail.ReadMessage(buf)
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	subject, err := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
	if err != nil {
		t.Fatalf("decode subject: %v", err)
	}
	if subject != "Your appointment on Tuesday 3 November 2026 is confirmed" {
		t.Errorf("unexpected subject %q", subject)
	}
	if got := msg.Header.Get("To"); got != "client@example.com" {
		t.Errorf("To = %q", got)
	}

	var html, ics string
	for _, p := range leafParts(t, msg.Header.Get("Content-Type"), msg.Body) {
		switch {
		case strings.HasPrefix(p.contentType, "text/html"):
			html = string(p.body)
		case strings.HasPrefix(p.contentType, "text/calendar"):
			ics = string(p.body)
		}
	}
	for _, want := range []string{"Appointment confirmed", "14:30 – 15:30", "€20.00", "€40.00", "Manage my booking", "https://leviosa.test/bookings?token=tok"} {
		if !strings.Contains(html, want) {
			t.Errorf("HTML body missing %q", want)
		}
	}
	for _, want := range []string{"DTSTART:20261103T133000Z", "DTEND:20261103T143000Z", "UID:booking-550e8400-e29b-41d4-a716-446655440000@leviosa"} {
		if !strings.Contains(ics, want) {
			t.Errorf(".ics attachment missing %q", want)
		}
	}
}

func TestSendBookingPartnerNoticeEmail_French(t *testing.T) {
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.UTC) // 10:00 in Paris (summer time)
	c, buf := capturingClient(t)
	err := c.SendBookingPartnerNoticeEmail(context.Background(), domain.BookingPartnerNoticeRequest{
		ToEmail:     "partner@example.com",
		ToFirstName: "Bob",
		BookingID:   "b-1",
		ProductName: "Massage",
		ClientName:  "Alice Smith",
		ClientEmail: "alice@example.com",
		ClientPhone: "+33612345678",
		StartTime:   start,
		EndTime:     start.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("SendBookingPartnerNoticeEmail: %v", err)
	}
	msg, err := mail.ReadMessage(buf)
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	if strings.HasPrefix(msg.Header.Get("Content-Type"), "multipart/mixed") {
		t.Error("the partner notice has no attachment")
	}
	subject, _ := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
	if subject != "Nouvelle réservation : Massage, mercredi 1 juillet 2026 à 10h00" {
		t.Errorf("unexpected subject %q", subject)
	}
}

// parseSent reads the captured message and returns its headers' decoded
// subject and its HTML body.
func parseSent(t *testing.T, buf *bytes.Buffer) (*mail.Message, string, string) {
	t.Helper()
	msg, err := mail.ReadMessage(buf)
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	subject, err := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
	if err != nil {
		t.Fatalf("decode subject: %v", err)
	}
	var html string
	for _, p := range leafParts(t, msg.Header.Get("Content-Type"), msg.Body) {
		if strings.HasPrefix(p.contentType, "text/html") {
			html = string(p.body)
		}
	}
	return msg, subject, html
}

func TestSendContactFormEmail_FromCompanyReplyToVisitor(t *testing.T) {
	c, buf := capturingClient(t)
	err := c.SendContactFormEmail(context.Background(), domain.ContactFormRequest{
		ToEmail: "inbox@leviosa.test", FromEmail: "contact@leviosa.test",
		Name: "Marie Leroy", ContactEmail: "marie.leroy@example.com", ContactPhone: "+33698765432",
		Message: "Bonjour,\nAvez-vous des créneaux le samedi ?", SubmittedAt: "05/03/2026 10:00",
	})
	if err != nil {
		t.Fatalf("SendContactFormEmail: %v", err)
	}

	msg, subject, html := parseSent(t, buf)
	if subject != "Nouveau message de contact — Marie Leroy" {
		t.Errorf("subject = %q", subject)
	}
	if got := msg.Header.Get("From"); got != "contact@leviosa.test" {
		t.Errorf("From = %q, want the request's FromEmail", got)
	}
	if got := msg.Header.Get("To"); got != "inbox@leviosa.test" {
		t.Errorf("To = %q", got)
	}
	if got := msg.Header.Get("Reply-To"); got != "marie.leroy@example.com" {
		t.Errorf("Reply-To = %q, want the visitor's email", got)
	}
	for _, want := range []string{"Marie Leroy", "marie.leroy@example.com", "05/03/2026 10:00"} {
		if !strings.Contains(html, want) {
			t.Errorf("body misses %q", want)
		}
	}
}

func TestSendContactFormEmail_NoReplyToWithoutVisitorEmail(t *testing.T) {
	c, buf := capturingClient(t)
	err := c.SendContactFormEmail(context.Background(), domain.ContactFormRequest{
		ToEmail: "inbox@leviosa.test", FromEmail: "contact@leviosa.test", Name: "Marie Leroy", Message: "Bonjour",
	})
	if err != nil {
		t.Fatalf("SendContactFormEmail: %v", err)
	}
	msg, _, _ := parseSent(t, buf)
	if got := msg.Header.Get("Reply-To"); got != "" {
		t.Errorf("Reply-To = %q, want none", got)
	}
}

func TestSendPaymentDisputeAlertEmail_FromRequestSender(t *testing.T) {
	c, buf := capturingClient(t)
	c.config.From = "noreply@leviosa.test"
	err := c.SendPaymentDisputeAlertEmail(context.Background(), domain.PaymentDisputeAlertRequest{
		ToEmail: "inbox@leviosa.test", BookingID: "b7k3xq", ClientName: "Jean Dupont",
		ProductName: "Massage", Amount: "€70.00", DisputeReason: "fraudulent", FromEmail: "contact@leviosa.test",
	})
	if err != nil {
		t.Fatalf("SendPaymentDisputeAlertEmail: %v", err)
	}

	msg, subject, html := parseSent(t, buf)
	if subject != "Payment dispute opened — booking b7k3xq" {
		t.Errorf("subject = %q", subject)
	}
	if got := msg.Header.Get("From"); got != "contact@leviosa.test" {
		t.Errorf("From = %q, want the request's FromEmail", got)
	}
	if got := msg.Header.Get("Reply-To"); got != "" {
		t.Errorf("Reply-To = %q, want none", got)
	}
	for _, want := range []string{"b7k3xq", "Jean Dupont", "Massage", "€70.00", "fraudulent"} {
		if !strings.Contains(html, want) {
			t.Errorf("body misses %q", want)
		}
	}
}

func TestSendPaymentFailedEmail_FallsBackToConfiguredFrom(t *testing.T) {
	c, buf := capturingClient(t)
	c.config.From = "noreply@leviosa.test"
	err := c.SendPaymentFailedEmail(context.Background(), domain.PaymentNotificationRequest{
		ToEmail: "client@example.com", ToFirstName: "Jean", Amount: "€70.00", Product: "Massage",
		PaymentDate: "05 March 2026", RetryURL: "https://leviosa.test/bookings",
	})
	if err != nil {
		t.Fatalf("SendPaymentFailedEmail: %v", err)
	}
	msg, subject, html := parseSent(t, buf)
	if subject != "Payment Failed" {
		t.Errorf("subject = %q", subject)
	}
	if got := msg.Header.Get("From"); got != "noreply@leviosa.test" {
		t.Errorf("From = %q, want the configured From when the request names none", got)
	}
	if !strings.Contains(html, "https://leviosa.test/bookings") {
		t.Error("body misses the retry link")
	}
}
