package app

import (
	"context"
	"slices"
	"testing"

	smtpClient "github.com/Leviosa-care/leviosa/backend/internal/notification/infrastructure/smtp"
	twilioClient "github.com/Leviosa-care/leviosa/backend/internal/notification/infrastructure/twilio"
)

func setTwilioEnv(t *testing.T) {
	t.Helper()
	t.Setenv("TWILIO_ACCOUNT_SID", "ACaccount")
	t.Setenv("TWILIO_API_KEY_SID", "SKapikey")
	t.Setenv("TWILIO_API_KEY_SECRET", "apikeysecret")
	t.Setenv("TWILIO_REGION", "ie1")
	t.Setenv("TWILIO_EDGE", "dublin")
	t.Setenv("TWILIO_SENDER", "Leviosa")
	t.Setenv("TWILIO_AUTH_TOKEN", "authtokenmustnotbeused")
}

func TestTwilioConfigFromEnvReadsAPIKeyRegionEdgeAndSender(t *testing.T) {
	setTwilioEnv(t)

	want := twilioClient.Config{
		AccountSID:   "ACaccount",
		APIKeySID:    "SKapikey",
		APIKeySecret: "apikeysecret",
		Region:       "ie1",
		Edge:         "dublin",
		Sender:       "Leviosa",
	}
	if got := TwilioConfigFromEnv(); got != want {
		t.Fatalf("TwilioConfigFromEnv() = %+v, want %+v", got, want)
	}
	if missing := MissingTwilioEnv(want); len(missing) != 0 {
		t.Fatalf("expected nothing missing, got %v", missing)
	}
}

func TestMissingTwilioEnvNamesOnlyRequiredEmptyVariables(t *testing.T) {
	cfg := twilioClient.Config{AccountSID: "ACaccount", Region: "ie1"}

	got := MissingTwilioEnv(cfg)
	want := []string{"TWILIO_API_KEY_SID", "TWILIO_API_KEY_SECRET", "TWILIO_SENDER"}
	if !slices.Equal(got, want) {
		t.Fatalf("MissingTwilioEnv() = %v, want %v", got, want)
	}
}

func TestSMTPConfigFromEnv(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.zoho.eu")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_USERNAME", "noreply@example.com")
	t.Setenv("SMTP_PASSWORD", "smtppassword")

	want := smtpClient.SMTPConfig{Host: "smtp.zoho.eu", Port: 465, Username: "noreply@example.com", Password: "smtppassword"}
	if got := SMTPConfigFromEnv(); got != want {
		t.Fatalf("SMTPConfigFromEnv() = %+v, want %+v", got, want)
	}

	if missing := MissingSMTPEnv(smtpClient.SMTPConfig{Host: "smtp.zoho.eu", Port: 587}); !slices.Equal(missing, []string{"SMTP_USERNAME", "SMTP_PASSWORD"}) {
		t.Fatalf("MissingSMTPEnv() = %v", missing)
	}
}

func TestLoadConfigUsesProviderConfigs(t *testing.T) {
	t.Setenv("ENVIRONMENT", "staging")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("SESSION_SECRET", "a-real-secret-that-is-not-the-default")
	setTwilioEnv(t)
	t.Setenv("SMTP_USERNAME", "noreply@example.com")

	cfg, err := LoadConfig(context.Background())
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Twilio != TwilioConfigFromEnv() {
		t.Fatalf("cfg.Twilio = %+v, want the TwilioConfigFromEnv values", cfg.Twilio)
	}
	if cfg.SMTP != SMTPConfigFromEnv() {
		t.Fatalf("cfg.SMTP = %+v, want the SMTPConfigFromEnv values", cfg.SMTP)
	}
}
