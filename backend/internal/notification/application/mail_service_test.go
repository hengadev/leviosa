package application

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/ports"
	settingsDomain "github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	settingsPorts "github.com/Leviosa-care/leviosa/backend/internal/settings/ports"
)

// stubSettingsService implements settingsPorts.SettingsService, returning
// configurable company info and errors for the fields SendContactFormEmail needs.
type stubSettingsService struct {
	email    string
	name     string
	logoURL  string
	emailErr error
	nameErr  error
	logoErr  error
}

var _ settingsPorts.SettingsService = (*stubSettingsService)(nil)

func (s *stubSettingsService) GetCompanyLogo(ctx context.Context) (*settingsDomain.GetCompanyLogoResponse, error) {
	if s.logoErr != nil {
		return nil, s.logoErr
	}
	return &settingsDomain.GetCompanyLogoResponse{LogoURL: s.logoURL}, nil
}
func (s *stubSettingsService) GetCompanyName(ctx context.Context) (*settingsDomain.GetCompanyNameResponse, error) {
	if s.nameErr != nil {
		return nil, s.nameErr
	}
	return &settingsDomain.GetCompanyNameResponse{Name: s.name}, nil
}
func (s *stubSettingsService) GetCompanyEmail(ctx context.Context) (*settingsDomain.GetCompanyEmailResponse, error) {
	if s.emailErr != nil {
		return nil, s.emailErr
	}
	return &settingsDomain.GetCompanyEmailResponse{Email: s.email}, nil
}
func (s *stubSettingsService) GetCompanyTelephone(ctx context.Context) (*settingsDomain.GetCompanyTelephoneResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) GetCompanyLegalAddress(ctx context.Context) (*settingsDomain.GetCompanyLegalAddressResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) GetCompanyInstagram(ctx context.Context) (*settingsDomain.GetCompanyInstagramResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) GetOTPDuration(ctx context.Context) (*settingsDomain.GetOTPDurationResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) GetOTPLength(ctx context.Context) (*settingsDomain.GetOTPLengthResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) GetOTPMaxAttempts(ctx context.Context) (*settingsDomain.GetOTPMaxAttemptsResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) GetAccessTokenDuration(ctx context.Context) (*settingsDomain.GetAccessTokenDurationResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) GetRefreshTokenDuration(ctx context.Context) (*settingsDomain.GetRefreshTokenDurationResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetCompanyName(ctx context.Context, req *settingsDomain.SetCompanyNameRequest) (*settingsDomain.SetCompanyNameResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetCompanyEmail(ctx context.Context, req *settingsDomain.SetCompanyEmailRequest) (*settingsDomain.SetCompanyEmailResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetCompanyTelephone(ctx context.Context, req *settingsDomain.SetCompanyTelephoneRequest) (*settingsDomain.SetCompanyTelephoneResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetCompanyLegalAddress(ctx context.Context, req *settingsDomain.SetCompanyLegalAddressRequest) (*settingsDomain.SetCompanyLegalAddressResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetCompanyInstagram(ctx context.Context, req *settingsDomain.SetCompanyInstagramRequest) (*settingsDomain.SetCompanyInstagramResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetCompanyLogo(ctx context.Context, file io.Reader, fileSize int64, contentType string) (*settingsDomain.SetCompanyLogoResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetOTPDuration(ctx context.Context, req *settingsDomain.SetOTPDurationRequest) (*settingsDomain.SetOTPDurationResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetOTPLength(ctx context.Context, req *settingsDomain.SetOTPLengthRequest) (*settingsDomain.SetOTPLengthResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetOTPMaxAttempts(ctx context.Context, req *settingsDomain.SetOTPMaxAttemptsRequest) (*settingsDomain.SetOTPMaxAttemptsResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetAccessTokenDuration(ctx context.Context, req *settingsDomain.SetAccessTokenDurationRequest) (*settingsDomain.SetAccessTokenDurationResponse, error) {
	return nil, nil
}
func (s *stubSettingsService) SetRefreshTokenDuration(ctx context.Context, req *settingsDomain.SetRefreshTokenDurationRequest) (*settingsDomain.SetRefreshTokenDurationResponse, error) {
	return nil, nil
}

// fakeEmailClient implements ports.EmailService, recording the last
// ContactFormRequest it received so tests can assert on it.
type fakeEmailClient struct {
	lastContactForm domain.ContactFormRequest
	contactFormErr  error
}

var _ ports.EmailService = (*fakeEmailClient)(nil)

func (f *fakeEmailClient) SendOTPEmail(ctx context.Context, req domain.OTPEmailRequest) error {
	return nil
}
func (f *fakeEmailClient) SendWelcomeEmail(ctx context.Context, req domain.WelcomeEmailRequest) error {
	return nil
}
func (f *fakeEmailClient) SendVerifyEmailEmail(ctx context.Context, req domain.VerifyEmailRequest) error {
	return nil
}
func (f *fakeEmailClient) SendEventNotificationEmail(ctx context.Context, req domain.EventNotificationRequest) error {
	return nil
}
func (f *fakeEmailClient) SendPaymentNotificationEmail(ctx context.Context, req domain.PaymentNotificationRequest) error {
	return nil
}
func (f *fakeEmailClient) SendPaymentFailedEmail(ctx context.Context, req domain.PaymentNotificationRequest) error {
	return nil
}
func (f *fakeEmailClient) SendBookingConfirmationEmail(ctx context.Context, req domain.BookingConfirmationRequest) error {
	return nil
}
func (f *fakeEmailClient) SendBookingCancellationEmail(ctx context.Context, req domain.BookingCancellationRequest) error {
	return nil
}
func (f *fakeEmailClient) SendBookingReminderEmail(ctx context.Context, req domain.BookingReminderRequest) error {
	return nil
}
func (f *fakeEmailClient) SendContactFormEmail(ctx context.Context, req domain.ContactFormRequest) error {
	f.lastContactForm = req
	return f.contactFormErr
}

func TestMailService_SendContactFormEmail_Success(t *testing.T) {
	settings := &stubSettingsService{
		email:   "staff@leviosa.care",
		name:    "Leviosa",
		logoURL: "https://cdn.example.com/logo.png",
	}
	emailClient := &fakeEmailClient{}
	provider := NewSettingsProvider(settings)
	mailService := NewMailService(emailClient, provider)

	err := mailService.SendContactFormEmail(context.Background(), "Alice Dupont", "alice@example.com", "", "Bonjour, je voudrais des informations.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := emailClient.lastContactForm
	if got.ToEmail != "staff@leviosa.care" {
		t.Errorf("ToEmail = %q, want staff inbox", got.ToEmail)
	}
	if got.FromEmail != "staff@leviosa.care" {
		t.Errorf("FromEmail = %q, want company email", got.FromEmail)
	}
	if got.Name != "Alice Dupont" {
		t.Errorf("Name = %q, want %q", got.Name, "Alice Dupont")
	}
	if got.ContactEmail != "alice@example.com" {
		t.Errorf("ContactEmail = %q, want %q", got.ContactEmail, "alice@example.com")
	}
	if got.ContactPhone != "" {
		t.Errorf("ContactPhone = %q, want empty", got.ContactPhone)
	}
	if !strings.Contains(got.Message, "informations") {
		t.Errorf("Message = %q, want it to contain the submitted text", got.Message)
	}
	if got.CompanyName != "Leviosa" {
		t.Errorf("CompanyName = %q, want %q", got.CompanyName, "Leviosa")
	}
	if got.LogoURL != "https://cdn.example.com/logo.png" {
		t.Errorf("LogoURL = %q, want logo URL", got.LogoURL)
	}
	if got.SubmittedAt == "" {
		t.Error("SubmittedAt should not be empty")
	}
}

func TestMailService_SendContactFormEmail_PropagatesSettingsError(t *testing.T) {
	settings := &stubSettingsService{emailErr: errors.New("settings unavailable")}
	emailClient := &fakeEmailClient{}
	provider := NewSettingsProvider(settings)
	mailService := NewMailService(emailClient, provider)

	err := mailService.SendContactFormEmail(context.Background(), "Bob", "bob@example.com", "", "Hello")
	if err == nil {
		t.Fatal("expected an error when company email lookup fails")
	}
}

func TestMailService_SendContactFormEmail_PropagatesEmailClientError(t *testing.T) {
	settings := &stubSettingsService{email: "staff@leviosa.care", name: "Leviosa", logoURL: "logo.png"}
	emailClient := &fakeEmailClient{contactFormErr: errors.New("smtp down")}
	provider := NewSettingsProvider(settings)
	mailService := NewMailService(emailClient, provider)

	err := mailService.SendContactFormEmail(context.Background(), "Bob", "bob@example.com", "", "Hello")
	if err == nil {
		t.Fatal("expected an error when the email client fails")
	}
}
