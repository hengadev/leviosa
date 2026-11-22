package smtp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/ports"

	"gopkg.in/gomail.v2"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	// From is the sender of the emails that name none (booking emails, the
	// check email). Empty falls back to Username, which suits a mailbox
	// login (Zoho) but not a relay's (Brevo's is an @smtp-brevo.com ID).
	From string
}

// Sender is the address the emails that name none are sent from: From,
// else Username.
func (c SMTPConfig) Sender() string {
	if c.From != "" {
		return c.From
	}
	return c.Username
}

type SMTPClient struct {
	config SMTPConfig
	// send delivers a built message; the SMTP dialer unless a test replaces
	// it.
	send func(*gomail.Message) error
}

var _ ports.EmailService = (*SMTPClient)(nil)

func NewSMTPClient(config SMTPConfig) *SMTPClient {
	client := &SMTPClient{config: config}
	client.send = func(m *gomail.Message) error {
		return gomail.NewDialer(config.Host, config.Port, config.Username, config.Password).DialAndSend(m)
	}
	return client
}

// SendCheckEmail sends one short test email to `to` through the same message
// and dialer path as every other email (the sender is the configured From,
// as for booking emails). The check-email run mode uses it to prove
// the SMTP settings end to end.
func (c *SMTPClient) SendCheckEmail(ctx context.Context, to string) error {
	data := struct {
		SentAt string
	}{
		SentAt: time.Now().UTC().Format(time.RFC3339),
	}

	emailReq := &domain.EmailRequest{
		To:       to,
		Subject:  "Leviosa SMTP check",
		Template: "smtp_check",
		Data:     data,
	}

	return c.sendEmail(ctx, emailReq, "")
}

func (c *SMTPClient) SendOTPEmail(ctx context.Context, req domain.OTPEmailRequest) error {
	return c.sendComposedFrom(ctx, req.ToEmail, req.FromEmail, domain.ComposeOTPEmail(req, time.Now()))
}

func (c *SMTPClient) SendWelcomeEmail(ctx context.Context, req domain.WelcomeEmailRequest) error {
	data := struct {
		FirstName   string
		LastName    string
		CompanyName string
		Year        int
		LogoURL     string
	}{
		FirstName:   req.ToFirstName,
		LastName:    req.ToLastName,
		CompanyName: req.CompanyName,
		Year:        time.Now().Year(),
		LogoURL:     req.LogoURL,
	}

	emailReq := &domain.EmailRequest{
		To:       req.ToEmail,
		Subject:  fmt.Sprintf("Welcome to %s", req.CompanyName),
		Template: "welcome",
		Data:     data,
	}

	return c.sendEmail(ctx, emailReq, req.FromEmail)
}

func (c *SMTPClient) SendVerifyEmailEmail(ctx context.Context, req domain.VerifyEmailRequest) error {
	data := struct {
		FirstName   string
		LastName    string
		CompanyName string
		Year        int
		LogoURL     string
	}{
		FirstName:   req.ToFirstName,
		LastName:    req.ToLastName,
		CompanyName: req.CompanyName,
		Year:        time.Now().Year(),
		LogoURL:     req.LogoURL,
	}

	emailReq := &domain.EmailRequest{
		To:       req.ToEmail,
		Subject:  "Verify Your Email",
		Template: "verify_email",
		Data:     data,
	}

	return c.sendEmail(ctx, emailReq, req.FromEmail)
}

func (c *SMTPClient) SendEventNotificationEmail(ctx context.Context, req domain.EventNotificationRequest) error {
	data := struct {
		FirstName   string
		LastName    string
		Event       string
		Details     string
		CompanyName string
		Year        int
		LogoURL     string
	}{
		FirstName:   req.ToFirstName,
		LastName:    req.ToLastName,
		Event:       req.Event,
		Details:     req.Details,
		CompanyName: req.CompanyName,
		Year:        time.Now().Year(),
		LogoURL:     req.LogoURL,
	}

	emailReq := &domain.EmailRequest{
		To:       req.ToEmail,
		Subject:  fmt.Sprintf("Event Notification: %s", req.Event),
		Template: "event_notification",
		Data:     data,
	}

	return c.sendEmail(ctx, emailReq, req.FromEmail)
}

func (c *SMTPClient) SendPaymentNotificationEmail(ctx context.Context, req domain.PaymentNotificationRequest) error {
	return c.sendComposedFrom(ctx, req.ToEmail, req.FromEmail, domain.ComposePaymentNotificationEmail(req, time.Now()))
}

func (c *SMTPClient) SendPaymentFailedEmail(ctx context.Context, req domain.PaymentNotificationRequest) error {
	return c.sendComposedFrom(ctx, req.ToEmail, req.FromEmail, domain.ComposePaymentFailedEmail(req, time.Now()))
}

// SendBookingConfirmationEmail sends the client confirmation of a booking,
// in the booking's locale, with the appointment as an .ics attachment.
func (c *SMTPClient) SendBookingConfirmationEmail(ctx context.Context, req domain.BookingConfirmationRequest) error {
	return c.sendComposed(ctx, req.ToEmail, domain.ComposeBookingConfirmation(req, time.Now()))
}

// SendBookingPartnerNoticeEmail tells the partner a new booking was
// confirmed (French).
func (c *SMTPClient) SendBookingPartnerNoticeEmail(ctx context.Context, req domain.BookingPartnerNoticeRequest) error {
	return c.sendComposed(ctx, req.ToEmail, domain.ComposeBookingPartnerNotice(req, time.Now()))
}

// SendLatePaymentRefundEmail tells a client, in the booking's locale, that
// their payment arrived after the hold and was refunded in full because the
// slot was no longer available.
func (c *SMTPClient) SendLatePaymentRefundEmail(ctx context.Context, req domain.LatePaymentRefundRequest) error {
	return c.sendComposed(ctx, req.ToEmail, domain.ComposeLatePaymentRefund(req, time.Now()))
}

// sendComposed sends an email composed by the domain, from the configured
// From, as every booking email.
func (c *SMTPClient) sendComposed(ctx context.Context, to string, email domain.ComposedEmail) error {
	return c.sendComposedFrom(ctx, to, "", email)
}

// sendComposedFrom sends an email composed by the domain from `from`; empty
// means the configured From.
func (c *SMTPClient) sendComposedFrom(ctx context.Context, to, from string, email domain.ComposedEmail) error {
	return c.sendEmail(ctx, &domain.EmailRequest{
		To:          to,
		ReplyTo:     email.ReplyTo,
		Subject:     email.Subject,
		Template:    email.Template,
		Data:        email.Data,
		Attachments: email.Attachments,
	}, from)
}

// SendBookingClientCancellationEmail tells a client, in the booking's
// locale, that their booking is cancelled and what was refunded.
func (c *SMTPClient) SendBookingClientCancellationEmail(ctx context.Context, req domain.BookingClientCancellationRequest) error {
	return c.sendComposed(ctx, req.ToEmail, domain.ComposeBookingClientCancellation(req, time.Now()))
}

// SendBookingPartnerCancellationEmail tells the partner, in French, that a
// booking was cancelled.
func (c *SMTPClient) SendBookingPartnerCancellationEmail(ctx context.Context, req domain.BookingPartnerCancellationRequest) error {
	return c.sendComposed(ctx, req.ToEmail, domain.ComposeBookingPartnerCancellation(req, time.Now()))
}

// SendBookingReminderEmail sends the client reminder of a confirmed booking,
// in the booking's locale.
func (c *SMTPClient) SendBookingReminderEmail(ctx context.Context, req domain.BookingReminderRequest) error {
	return c.sendComposed(ctx, req.ToEmail, domain.ComposeBookingReminder(req, time.Now()))
}

// SendBookingAccessUpdateEmail tells a client with an upcoming booking, in
// the booking's locale, that the practice access code changed, with the new
// access. No attachment.
func (c *SMTPClient) SendBookingAccessUpdateEmail(ctx context.Context, req domain.BookingAccessUpdateRequest) error {
	return c.sendComposed(ctx, req.ToEmail, domain.ComposeBookingAccessUpdate(req, time.Now()))
}

// SendContactFormEmail sends a visitor's message to the practice; replies go
// to the visitor when they left an email.
func (c *SMTPClient) SendContactFormEmail(ctx context.Context, req domain.ContactFormRequest) error {
	return c.sendComposedFrom(ctx, req.ToEmail, req.FromEmail, domain.ComposeContactFormEmail(req, time.Now()))
}

func (c *SMTPClient) SendPaymentDisputeAlertEmail(ctx context.Context, req domain.PaymentDisputeAlertRequest) error {
	return c.sendComposedFrom(ctx, req.ToEmail, req.FromEmail, domain.ComposePaymentDisputeAlertEmail(req, time.Now()))
}

func (c *SMTPClient) sendEmail(ctx context.Context, request *domain.EmailRequest, fromEmail string) error {
	if request == nil {
		return errs.ErrInvalidValue
	}

	if request.To == "" {
		return fmt.Errorf("recipient email cannot be empty: %w", errs.ErrInvalidValue)
	}

	if request.Subject == "" {
		return fmt.Errorf("email subject cannot be empty: %w", errs.ErrInvalidValue)
	}

	if fromEmail == "" {
		fromEmail = c.config.Sender()
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fromEmail)
	m.SetHeader("To", request.To)
	m.SetHeader("Subject", request.Subject)

	if request.ReplyTo != "" {
		m.SetHeader("Reply-To", request.ReplyTo)
	}

	if len(request.CarbonCopy) > 0 {
		addresses := make([]string, 0, len(request.CarbonCopy))
		for email, name := range request.CarbonCopy {
			addresses = append(addresses, m.FormatAddress(email, name))
		}
		m.SetHeader("Cc", addresses...)
	}

	// Embed Instagram image
	instagramPath, err := writeTempFile(domain.InstagramImage, "instagram.png")
	if err != nil {
		return fmt.Errorf("write instagram temp file: %w", err)
	}
	defer os.Remove(instagramPath)
	m.Embed(instagramPath, gomail.Rename("instagram.png"))

	// Embed additional images if specified
	for path, rename := range request.Images {
		m.Embed(path, gomail.Rename(rename))
	}

	if request.Template != "" {
		htmlBody, err := c.renderTemplate(request.Template, request.Data)
		if err != nil {
			return fmt.Errorf("render email template: %w", err)
		}
		m.SetBody("text/html", htmlBody)
	}

	// Attachments make the message multipart/mixed.
	for _, attachment := range request.Attachments {
		attachment := attachment
		m.Attach(attachment.Filename,
			gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(attachment.Data)
				return err
			}),
			gomail.SetHeader(map[string][]string{
				"Content-Type": {attachment.ContentType + `; name="` + attachment.Filename + `"`},
			}),
		)
	}

	if err := c.send(m); err != nil {
		return fmt.Errorf("send email via SMTP: %w", err)
	}

	return nil
}

func (c *SMTPClient) renderTemplate(templateName string, data interface{}) (string, error) {
	return domain.RenderEmailTemplate(templateName, data)
}

func writeTempFile(data []byte, filename string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("file data cannot be empty: %w", errs.ErrInvalidValue)
	}

	ext := filepath.Ext(filename)
	stem := filename[:len(filename)-len(ext)]

	f, err := os.CreateTemp("", stem+"*"+ext)
	if err != nil {
		return "", fmt.Errorf("create temp file %s: %w", filename, err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("write temp file %s: %w", filename, err)
	}

	return f.Name(), nil
}
