package encryption

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hengadev/encx"
)

var _ encx.KeyManagementService = kmsKeyProvider{}

// encx creates a KMS key when the alias doesn't resolve; the app's provider never does.
func TestKMSKeyProvider_RefusesToCreateKeys(t *testing.T) {
	_, err := kmsKeyProvider{}.CreateKey(context.Background(), "alias/leviosa-dev")
	if !errors.Is(err, ErrKeyCreationRefused) {
		t.Fatalf("CreateKey error = %v, want ErrKeyCreationRefused", err)
	}
}

func TestNewCrypto_InvalidConfigFailsBeforeAnyKMSCall(t *testing.T) {
	cfg := validConfig()
	cfg.Pepper = ""
	cfg.AWSEndpointURL = "http://127.0.0.1:1" // nothing listens: a KMS call would fail differently

	_, err := NewCrypto(context.Background(), cfg)
	if !errors.Is(err, ErrPepperMissing) {
		t.Fatalf("NewCrypto error = %v, want ErrPepperMissing", err)
	}
	if !strings.Contains(err.Error(), "encryption config") {
		t.Fatalf("NewCrypto error = %v, want a config error", err)
	}
}
