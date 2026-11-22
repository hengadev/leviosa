package domain

import (
	"strings"
	"testing"
)

// These emails keep the subject, template and data the SMTP client built
// inline before they were composed here.

func TestComposeOTPEmail(t *testing.T) {
	email := ComposeOTPEmail(OTPEmailRequest{
		ToEmail: "client@example.com", OTP: "428193", FromEmail: "contact@leviosa.test", LogoURL: "https://leviosa.test/logo.png",
	}, composeNow)

	if email.Subject != "Your OTP Code" {
		t.Errorf("subject = %q", email.Subject)
	}
	if email.Template != "otp" {
		t.Errorf("template = %q", email.Template)
	}
	want := OTPEmailData{OTP: "428193", Year: 2026, LogoURL: "https://leviosa.test/logo.png"}
	if got, ok := email.Data.(OTPEmailData); !ok || got != want {
		t.Errorf("data = %+v, want %+v", email.Data, want)
	}
	if email.ReplyTo != "" || len(email.Attachments) != 0 {
		t.Errorf("unexpected reply-to %q or attachments %d", email.ReplyTo, len(email.Attachments))
	}
}

// The template reads Address and InstagramPath: without them in the data it
// failed to execute, so no login code email could be sent.
func TestComposeOTPEmail_Renders(t *testing.T) {
	email := ComposeOTPEmail(OTPEmailRequest{ToEmail: "client@example.com", OTP: "428193"}, composeNow)
	html, err := RenderEmailTemplate(email.Template, email.Data)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(html, "428193") {
		t.Error("rendered email does not show the code")
	}
	if strings.Contains(html, "Instagram") {
		t.Error("an empty Instagram link is shown")
	}
}

func testPaymentNotificationRequest() PaymentNotificationRequest {
	return PaymentNotificationRequest{
		ToEmail: "client@example.com", ToFirstName: "Alice", ToLastName: "Smith",
		Amount: "€70.00", Product: "Massage", PaymentDate: "05 March 2026",
		FromEmail: "contact@leviosa.test", CompanyName: "Leviosa", LogoURL: "https://leviosa.test/logo.png",
		RetryURL: "https://leviosa.test/bookings",
	}
}

func TestComposePaymentNotificationEmail(t *testing.T) {
	email := ComposePaymentNotificationEmail(testPaymentNotificationRequest(), composeNow)

	if email.Subject != "Payment Confirmation" {
		t.Errorf("subject = %q", email.Subject)
	}
	if email.Template != "payment" {
		t.Errorf("template = %q", email.Template)
	}
	want := PaymentEmailData{
		FirstName: "Alice", LastName: "Smith", Amount: "€70.00", Product: "Massage", PaymentDate: "05 March 2026",
		CompanyName: "Leviosa", Year: 2026, LogoURL: "https://leviosa.test/logo.png",
	}
	if got, ok := email.Data.(PaymentEmailData); !ok || got != want {
		t.Errorf("data = %+v, want %+v", email.Data, want)
	}
	renderComposed(t, email)
}

func TestComposePaymentFailedEmail(t *testing.T) {
	email := ComposePaymentFailedEmail(testPaymentNotificationRequest(), composeNow)

	if email.Subject != "Payment Failed" {
		t.Errorf("subject = %q", email.Subject)
	}
	if email.Template != "payment_failed" {
		t.Errorf("template = %q", email.Template)
	}
	data, ok := email.Data.(PaymentFailedEmailData)
	if !ok {
		t.Fatalf("data is %T", email.Data)
	}
	if data.RetryURL != "https://leviosa.test/bookings" || data.FirstName != "Alice" || data.Amount != "€70.00" ||
		data.Product != "Massage" || data.PaymentDate != "05 March 2026" || data.Year != 2026 {
		t.Errorf("data = %+v", data)
	}
	body := renderComposed(t, email)
	for _, want := range []string{"Bonjour Alice,", "Massage", "€70.00", "05 March 2026", `href="https://leviosa.test/bookings"`} {
		assertContains(t, body, want)
	}
}

func TestComposeContactFormEmail(t *testing.T) {
	email := ComposeContactFormEmail(ContactFormRequest{
		ToEmail: "contact@leviosa.test", FromEmail: "contact@leviosa.test",
		Name: "Marie Leroy", ContactEmail: "marie.leroy@example.com", ContactPhone: "+33698765432",
		Message: "Bonjour,\nAvez-vous des créneaux le samedi ?", CompanyName: "Leviosa", SubmittedAt: "05/03/2026 10:00",
	}, composeNow)

	if email.Subject != "Nouveau message de contact — Marie Leroy" {
		t.Errorf("subject = %q", email.Subject)
	}
	if email.Template != "contact_form" {
		t.Errorf("template = %q", email.Template)
	}
	if email.ReplyTo != "marie.leroy@example.com" {
		t.Errorf("reply-to = %q, want the visitor's email", email.ReplyTo)
	}
	want := ContactFormEmailData{
		Name: "Marie Leroy", ContactEmail: "marie.leroy@example.com", ContactPhone: "+33698765432",
		Message: "Bonjour,\nAvez-vous des créneaux le samedi ?", SubmittedAt: "05/03/2026 10:00",
		CompanyName: "Leviosa", Year: 2026,
	}
	if got, ok := email.Data.(ContactFormEmailData); !ok || got != want {
		t.Errorf("data = %+v, want %+v", email.Data, want)
	}
	renderComposed(t, email)
}

func TestComposeContactFormEmail_NoReplyToWithoutVisitorEmail(t *testing.T) {
	email := ComposeContactFormEmail(ContactFormRequest{Name: "Marie Leroy", Message: "Bonjour"}, composeNow)
	if email.ReplyTo != "" {
		t.Errorf("reply-to = %q, want none", email.ReplyTo)
	}
}

func TestComposePaymentDisputeAlertEmail(t *testing.T) {
	email := ComposePaymentDisputeAlertEmail(PaymentDisputeAlertRequest{
		ToEmail: "contact@leviosa.test", BookingID: "550e8400-e29b-41d4-a716-446655440000",
		ClientName: "Alice Smith", ProductName: "Massage", Amount: "€70.00", DisputeReason: "fraudulent",
	}, composeNow)

	if email.Subject != "Payment dispute opened — booking 550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("subject = %q", email.Subject)
	}
	if email.Template != "payment_dispute_alert" {
		t.Errorf("template = %q", email.Template)
	}
	want := PaymentDisputeAlertEmailData{
		BookingID: "550e8400-e29b-41d4-a716-446655440000", ClientName: "Alice Smith", ProductName: "Massage",
		Amount: "€70.00", DisputeReason: "fraudulent", Year: 2026,
	}
	if got, ok := email.Data.(PaymentDisputeAlertEmailData); !ok || got != want {
		t.Errorf("data = %+v, want %+v", email.Data, want)
	}
	renderComposed(t, email)
}
