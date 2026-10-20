package app

import (
	"context"
	"testing"
)

func TestLoadConfig_RejectsEmptySessionSecretInStaging(t *testing.T) {
	t.Setenv("ENVIRONMENT", "staging")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	// SESSION_SECRET intentionally not set → defaults to "development-secret-key"

	_, err := LoadConfig(context.Background())
	if err == nil {
		t.Fatal("expected LoadConfig to reject empty/default SESSION_SECRET in staging, got nil error")
	}
}

func TestLoadConfig_RejectsDefaultSessionSecretInProduction(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("SESSION_SECRET", "development-secret-key")

	_, err := LoadConfig(context.Background())
	if err == nil {
		t.Fatal("expected LoadConfig to reject default SESSION_SECRET in production, got nil error")
	}
}

func TestLoadConfig_AcceptsRealSecretInStaging(t *testing.T) {
	t.Setenv("ENVIRONMENT", "staging")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("SESSION_SECRET", "a-real-secret-that-is-not-the-default")

	cfg, err := LoadConfig(context.Background())
	if err != nil {
		t.Fatalf("expected LoadConfig to succeed with a real secret in staging, got error: %v", err)
	}
	if cfg.SessionSecret != "a-real-secret-that-is-not-the-default" {
		t.Errorf("expected SessionSecret to be set, got %q", cfg.SessionSecret)
	}
}

func TestLoadConfig_DevelopmentAllowsDefaultSecret(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	// Ensure SESSION_SECRET is not set in the test environment
	t.Setenv("SESSION_SECRET", "")

	cfg, err := LoadConfig(context.Background())
	if err != nil {
		t.Fatalf("expected LoadConfig to succeed in development with default secret, got error: %v", err)
	}
	if cfg.SessionSecret != "development-secret-key" {
		t.Errorf("expected default SessionSecret in development, got %q", cfg.SessionSecret)
	}
}

func TestLoadConfig_ReadsEncryptionSettings(t *testing.T) {
	t.Setenv("ENVIRONMENT", "staging")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("SESSION_SECRET", "a-real-secret-that-is-not-the-default")
	t.Setenv("ENCX_KMS_KEY_ALIAS", "alias/leviosa-staging")
	t.Setenv("ENCX_PEPPER", "bGV2aW9zYS1sb2NhbC1kZXYtcGVwcGVyLTMyYnl0ZXM=")
	t.Setenv("AWS_REGION", "eu-central-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIDEXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
	t.Setenv("AWS_ENDPOINT_URL", "")

	cfg, err := LoadConfig(context.Background())
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	enc := cfg.Encryption
	if enc.KMSKeyAlias != "alias/leviosa-staging" || enc.Pepper != "bGV2aW9zYS1sb2NhbC1kZXYtcGVwcGVyLTMyYnl0ZXM=" ||
		enc.AWSRegion != "eu-central-1" || enc.AWSAccessKeyID != "AKIDEXAMPLE" ||
		enc.AWSSecretAccessKey != "secret" || enc.AWSEndpointURL != "" {
		t.Fatalf("unexpected encryption config: %+v", enc)
	}
	if err := enc.Validate(); err != nil {
		t.Fatalf("encryption config from a complete environment should validate: %v", err)
	}
}
