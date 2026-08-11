package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
)

// fakeNotificationService implements ports.NotificationService, recording the
// last SendContactFormNotification call so tests can assert on it.
type fakeNotificationService struct {
	called     bool
	gotName    string
	gotEmail   string
	gotPhone   string
	gotMessage string
	returnErr  error
}

func (f *fakeNotificationService) SendOTPEmail(ctx context.Context, email, otp string) error {
	return nil
}
func (f *fakeNotificationService) SendWelcomeEmail(ctx context.Context, email, firstName, lastName string) error {
	return nil
}
func (f *fakeNotificationService) SendVerifyEmailEmail(ctx context.Context, email, firstName, lastName string) error {
	return nil
}
func (f *fakeNotificationService) SendEventNotificationEmail(ctx context.Context, email, firstName, lastName, eventName, eventDetails string) error {
	return nil
}
func (f *fakeNotificationService) SendPaymentNotificationEmail(ctx context.Context, email, firstName, lastName, amount, product, paymentDate string) error {
	return nil
}
func (f *fakeNotificationService) SendContactFormNotification(ctx context.Context, name, contactEmail, contactPhone, message string) error {
	f.called = true
	f.gotName = name
	f.gotEmail = contactEmail
	f.gotPhone = contactPhone
	f.gotMessage = message
	return f.returnErr
}
func (f *fakeNotificationService) SendOTPBySMS(ctx context.Context, phoneNumber, otp string) error {
	return nil
}
func (f *fakeNotificationService) SendGenericSMS(ctx context.Context, phoneNumber, message string) error {
	return nil
}
func (f *fakeNotificationService) InvalidateSettingsCache() {}

func newTestRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.WithValue(req.Context(), ctxutil.LoggerKey, logger)
	return req.WithContext(ctx)
}

func TestSubmitContact_Success(t *testing.T) {
	svc := &fakeNotificationService{}
	h := New(svc)

	body := `{"name":"Alice Dupont","contact_email":"alice@example.com","message":"Bonjour, une question sur vos tarifs."}`
	req := newTestRequest(t, body)
	rec := httptest.NewRecorder()

	h.SubmitContact(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if !svc.called {
		t.Fatal("expected SendContactFormNotification to be called")
	}
	if svc.gotName != "Alice Dupont" {
		t.Errorf("name = %q, want %q", svc.gotName, "Alice Dupont")
	}
	if svc.gotEmail != "alice@example.com" {
		t.Errorf("email = %q, want %q", svc.gotEmail, "alice@example.com")
	}
}

func TestSubmitContact_RequiresName(t *testing.T) {
	svc := &fakeNotificationService{}
	h := New(svc)

	body := `{"contact_email":"alice@example.com","message":"Bonjour"}`
	req := newTestRequest(t, body)
	rec := httptest.NewRecorder()

	h.SubmitContact(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if svc.called {
		t.Error("service should not be called when validation fails")
	}
}

func TestSubmitContact_RequiresMessage(t *testing.T) {
	svc := &fakeNotificationService{}
	h := New(svc)

	body := `{"name":"Alice","contact_email":"alice@example.com"}`
	req := newTestRequest(t, body)
	rec := httptest.NewRecorder()

	h.SubmitContact(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSubmitContact_RequiresEmailOrPhone(t *testing.T) {
	svc := &fakeNotificationService{}
	h := New(svc)

	body := `{"name":"Alice","message":"Bonjour"}`
	req := newTestRequest(t, body)
	rec := httptest.NewRecorder()

	h.SubmitContact(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSubmitContact_AcceptsPhoneOnly(t *testing.T) {
	svc := &fakeNotificationService{}
	h := New(svc)

	body := `{"name":"Alice","contact_phone":"0612345678","message":"Rappelez-moi svp."}`
	req := newTestRequest(t, body)
	rec := httptest.NewRecorder()

	h.SubmitContact(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if svc.gotPhone != "0612345678" {
		t.Errorf("phone = %q, want %q", svc.gotPhone, "0612345678")
	}
}

func TestSubmitContact_NormalizesPhoneWhitespace(t *testing.T) {
	svc := &fakeNotificationService{}
	h := New(svc)

	body := `{"name":"Alice","contact_phone":"06 12 34 56 78","message":"Rappelez-moi svp."}`
	req := newTestRequest(t, body)
	rec := httptest.NewRecorder()

	h.SubmitContact(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if svc.gotPhone != "0612345678" {
		t.Errorf("phone = %q, want spaces stripped to %q", svc.gotPhone, "0612345678")
	}
}

func TestSubmitContact_RejectsInvalidEmail(t *testing.T) {
	svc := &fakeNotificationService{}
	h := New(svc)

	body := `{"name":"Alice","contact_email":"not-an-email","message":"Bonjour"}`
	req := newTestRequest(t, body)
	rec := httptest.NewRecorder()

	h.SubmitContact(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSubmitContact_RejectsWrongContentType(t *testing.T) {
	svc := &fakeNotificationService{}
	h := New(svc)

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("{}"))
	rec := httptest.NewRecorder()

	h.SubmitContact(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnsupportedMediaType)
	}
}

func TestSubmitContact_PropagatesServiceError(t *testing.T) {
	svc := &fakeNotificationService{returnErr: errors.New("smtp down")}
	h := New(svc)

	body := `{"name":"Alice","contact_email":"alice@example.com","message":"Bonjour"}`
	req := newTestRequest(t, body)
	rec := httptest.NewRecorder()

	h.SubmitContact(rec, req)

	if rec.Code < 500 {
		t.Fatalf("status = %d, want a 5xx when the notification service fails", rec.Code)
	}
}

func TestSubmitContact_RejectsUnknownFields(t *testing.T) {
	svc := &fakeNotificationService{}
	h := New(svc)

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(map[string]any{
		"name":          "Alice",
		"contact_email": "alice@example.com",
		"message":       "Bonjour",
		"unexpected":    "field",
	})
	req := newTestRequest(t, buf.String())
	rec := httptest.NewRecorder()

	h.SubmitContact(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
