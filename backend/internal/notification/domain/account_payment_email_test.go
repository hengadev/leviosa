package domain

import (
	"strings"
	"testing"
)

// The emails sent outside the booking flow: French subjects and bodies, the
// booking emails' "LEVIOSA" header and footer, no placeholder content.

// assertFinished checks a rendered email has the booking emails' header and
// footer and none of the old placeholder content.
func assertFinished(t *testing.T, body string) {
	t.Helper()
	for _, want := range []string{">LEVIOSA</h1>", "&copy; 2026 Leviosa. Tous droits réservés."} {
		assertContains(t, body, want)
	}
	for _, banned := range []string{"<img", "placeholder", "Placeholder", "[Client Name]", "Instagram", "Merci de vous être inscrit"} {
		if strings.Contains(body, banned) {
			t.Errorf("rendered email contains %q", banned)
		}
	}
}

func TestComposeOTPEmail(t *testing.T) {
	email := ComposeOTPEmail(OTPEmailRequest{
		ToEmail: "client@example.com", OTP: "428193", ValidForMinutes: 5, FromEmail: "contact@leviosa.test", LogoURL: "https://leviosa.test/logo.png",
	}, composeNow)

	if email.Subject != "Votre code de vérification Leviosa" {
		t.Errorf("subject = %q", email.Subject)
	}
	if email.Template != "otp" {
		t.Errorf("template = %q", email.Template)
	}
	want := OTPEmailData{OTP: "428193", ValidForMinutes: 5, Year: 2026}
	if got, ok := email.Data.(OTPEmailData); !ok || got != want {
		t.Errorf("data = %+v, want %+v", email.Data, want)
	}
	if email.ReplyTo != "" || len(email.Attachments) != 0 {
		t.Errorf("unexpected reply-to %q or attachments %d", email.ReplyTo, len(email.Attachments))
	}
}

// The template must execute with the data it is given: a field it reads that
// the data lacks once made every login code email fail to render. Its wording
// fits a password reset, an email check and a guest claim, and states how
// long the code stays valid when it is known.
func TestComposeOTPEmail_Renders(t *testing.T) {
	tests := []struct {
		minutes int
		want    string
	}{
		{10, "Ce code est valable 10 minutes."},
		{1, "Ce code est valable 1 minute."},
		{0, "Ce code n'est valable que quelques minutes."},
	}
	for _, tt := range tests {
		email := ComposeOTPEmail(OTPEmailRequest{ToEmail: "client@example.com", OTP: "428193", ValidForMinutes: tt.minutes}, composeNow)
		html, err := RenderEmailTemplate(email.Template, email.Data)
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		for _, want := range []string{"428193", "Votre code de vérification", tt.want} {
			assertContains(t, html, want)
		}
		assertFinished(t, html)
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

	if email.Subject != "Échec de votre paiement" {
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
	for _, want := range []string{"Échec de votre paiement", "Bonjour Alice,", "Massage", "€70.00", "05 March 2026", `href="https://leviosa.test/bookings"`} {
		assertContains(t, body, want)
	}
	assertFinished(t, body)
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

	if email.Subject != "Litige de paiement — réservation 550e8400-e29b-41d4-a716-446655440000" {
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
	body := renderComposed(t, email)
	for _, want := range []string{"Litige de paiement", "Alice Smith", "Massage", "€70.00", "fraudulent", "550e8400-e29b-41d4-a716-446655440000"} {
		assertContains(t, body, want)
	}
	assertFinished(t, body)
}
