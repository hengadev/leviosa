package domain

import (
	"bytes"
	"html/template"
	"testing"
)

func TestPaymentFailedTemplateRenders(t *testing.T) {
	tmpl := parseTemplateOrFail(t, "payment_failed")

	data := struct {
		FirstName   string
		LastName    string
		Amount      string
		Product     string
		PaymentDate string
		CompanyName string
		Year        int
		LogoURL     string
		RetryURL    string
	}{
		FirstName:   "Alice",
		LastName:    "Smith",
		Amount:      "€50.00",
		Product:     "Massage Therapy",
		PaymentDate: "15 June 2026",
		CompanyName: "Leviosa",
		Year:        2026,
		LogoURL:     "https://example.com/logo.png",
		RetryURL:    "https://app.leviosa.com/bookings",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute payment_failed template: %v", err)
	}

	assertContains(t, buf.String(), "€50.00")
	assertContains(t, buf.String(), "Massage Therapy")
	assertContains(t, buf.String(), "15 June 2026")
	assertContains(t, buf.String(), "Réessayer le paiement")
	assertContains(t, buf.String(), "https://app.leviosa.com/bookings")
	assertContains(t, buf.String(), "2026")

	// Without a retry URL, the CTA button must not render.
	dataNoURL := data
	dataNoURL.RetryURL = ""
	var bufNoURL bytes.Buffer
	if err := tmpl.Execute(&bufNoURL, dataNoURL); err != nil {
		t.Fatalf("failed to execute payment_failed template without RetryURL: %v", err)
	}
	if bytes.Contains(bufNoURL.Bytes(), []byte("Réessayer le paiement")) {
		t.Error("retry button should not appear when RetryURL is empty")
	}
}

func TestContactFormTemplateRenders(t *testing.T) {
	tmpl := parseTemplateOrFail(t, "contact_form")

	data := struct {
		Name         string
		ContactEmail string
		ContactPhone string
		Message      string
		SubmittedAt  string
		CompanyName  string
		Year         int
		LogoURL      string
	}{
		Name:         "Alice Dupont",
		ContactEmail: "alice@example.com",
		ContactPhone: "0612345678",
		Message:      "Bonjour, je voudrais des informations sur vos tarifs.",
		SubmittedAt:  "11/08/2026 14:30",
		CompanyName:  "Leviosa",
		Year:         2026,
		LogoURL:      "https://example.com/logo.png",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute contact_form template: %v", err)
	}

	assertContains(t, buf.String(), "Alice Dupont")
	assertContains(t, buf.String(), "alice@example.com")
	assertContains(t, buf.String(), "0612345678")
	assertContains(t, buf.String(), "Bonjour, je voudrais des informations sur vos tarifs.")
	assertContains(t, buf.String(), "11/08/2026 14:30")
	assertContains(t, buf.String(), "Leviosa")
	assertContains(t, buf.String(), "2026")
}

func TestContactFormTemplate_OmitsMissingContactFields(t *testing.T) {
	tmpl := parseTemplateOrFail(t, "contact_form")

	data := struct {
		Name         string
		ContactEmail string
		ContactPhone string
		Message      string
		SubmittedAt  string
		CompanyName  string
		Year         int
		LogoURL      string
	}{
		Name:         "Bob",
		ContactPhone: "0612345678",
		Message:      "Appelez-moi svp.",
		SubmittedAt:  "11/08/2026 14:30",
		CompanyName:  "Leviosa",
		Year:         2026,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute contact_form template without email: %v", err)
	}

	if bytes.Contains(buf.Bytes(), []byte("Email</td>")) {
		t.Error("email row should not render when ContactEmail is empty")
	}
	assertContains(t, buf.String(), "0612345678")
}

func parseTemplateOrFail(t *testing.T, name string) *template.Template {
	t.Helper()
	tmpl, err := template.ParseFS(EmailTemplates, "templates/"+name+".html")
	if err != nil {
		t.Fatalf("failed to parse template %s: %v", name, err)
	}
	return tmpl
}

func assertContains(t *testing.T, body, substr string) {
	t.Helper()
	if !bytes.Contains([]byte(body), []byte(substr)) {
		t.Errorf("template output missing expected substring %q", substr)
	}
}

func TestSMTPCheckTemplateRenders(t *testing.T) {
	tmpl := parseTemplateOrFail(t, "smtp_check")

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ SentAt string }{SentAt: "2026-10-20T10:00:00Z"}); err != nil {
		t.Fatalf("failed to execute smtp_check template: %v", err)
	}

	assertContains(t, buf.String(), "check-email")
	assertContains(t, buf.String(), "2026-10-20T10:00:00Z")
}
