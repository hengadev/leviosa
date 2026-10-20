package testutils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/hengadev/encx"

	"github.com/Leviosa-care/leviosa/backend/internal/common/encryption"
)

// TestKMSKeyAlias is the alias of the KEK that SetupCrypto creates.
const TestKMSKeyAlias = "alias/leviosa-test"

// CryptoSetup is encx provisioned the way the backend provisions it (ADR-0015): the KEK is
// a KMS key (LocalStack) addressed by alias, the pepper a configuration value, and the
// crypto service is built by encryption.NewCrypto.
type CryptoSetup struct {
	// Crypto is the crypto service to hand to the code under test.
	Crypto encx.CryptoService
	// KMS is the LocalStack container holding the KEK.
	KMS *KMSContainer
	// Config is what Crypto was built from: TestKMSKeyAlias, the generated pepper and the
	// LocalStack endpoint and credentials.
	Config encryption.Config
}

// SetupCrypto starts LocalStack KMS, creates the KEK behind TestKMSKeyAlias, generates a
// random pepper and builds encx from them. Call it once per test package, from TestMain
// with a nil t, and release it with TeardownCrypto.
func SetupCrypto(ctx context.Context, t *testing.T) (*CryptoSetup, error) {
	kmsContainer, err := SetupKMS(ctx, t)
	if err != nil {
		return nil, err
	}
	setup := &CryptoSetup{KMS: kmsContainer}

	fail := func(err error) (*CryptoSetup, error) {
		TeardownCrypto(ctx, t, setup)
		return nil, err
	}

	if _, err := kmsContainer.CreateKey(ctx, TestKMSKeyAlias); err != nil {
		return fail(err)
	}

	pepper, err := newPepper()
	if err != nil {
		return fail(err)
	}

	setup.Config = kmsContainer.EncryptionConfig(TestKMSKeyAlias, pepper)
	if setup.Config.KeyMetadataDir, err = os.MkdirTemp("", "leviosa-encx-test-"); err != nil {
		return fail(fmt.Errorf("create encx key metadata directory: %w", err))
	}

	if setup.Crypto, err = encryption.NewCrypto(ctx, setup.Config); err != nil {
		return fail(err)
	}
	return setup, nil
}

// TeardownCrypto terminates the KMS container and removes encx's key metadata directory;
// nil is a no-op.
func TeardownCrypto(ctx context.Context, t *testing.T, s *CryptoSetup) {
	if s == nil {
		return
	}
	TeardownKMS(ctx, t, s.KMS)
	if s.Config.KeyMetadataDir != "" {
		if err := os.RemoveAll(s.Config.KeyMetadataDir); err != nil && t != nil {
			t.Logf("remove encx key metadata directory: %v", err)
		}
	}
}

// newPepper returns a random pepper encoded as the backend reads it from ENCX_PEPPER.
func newPepper() (string, error) {
	pepper := make([]byte, encx.PepperLength)
	if _, err := rand.Read(pepper); err != nil {
		return "", fmt.Errorf("generate pepper: %w", err)
	}
	return base64.StdEncoding.EncodeToString(pepper), nil
}
