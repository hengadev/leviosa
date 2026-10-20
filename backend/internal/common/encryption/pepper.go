package encryption

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/hengadev/encx"
)

var (
	// ErrPepperMissing means no pepper is configured.
	ErrPepperMissing = errors.New(EnvPepper + " is not set")
	// ErrPepperInvalid means the configured pepper is not base64 of exactly encx.PepperLength bytes.
	ErrPepperInvalid = errors.New(EnvPepper + " is invalid")
	// ErrPepperReadOnly is returned by StorePepper: the pepper lives in configuration
	// (Infisical in deployed environments) and the app never creates or replaces one.
	ErrPepperReadOnly = errors.New("the pepper comes from " + EnvPepper + " and cannot be stored by the app")
)

// PepperStore is encx's secret store for the pepper, backed by configuration instead of a
// secrets service (ADR-0015). It serves one pepper whatever alias encx asks for, and it
// refuses to store one: encx generates and stores a new pepper when none exists, which
// would silently invalidate every password and searchable hash, so a missing or invalid
// pepper must stop startup instead.
type PepperStore struct {
	pepper []byte
}

// NewPepperStore decodes encoded (standard base64, surrounding whitespace ignored) and
// checks that it is exactly encx.PepperLength bytes.
func NewPepperStore(encoded string) (*PepperStore, error) {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return nil, ErrPepperMissing
	}

	pepper, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: not standard base64: %v", ErrPepperInvalid, err)
	}
	if len(pepper) != encx.PepperLength {
		return nil, fmt.Errorf("%w: decodes to %d bytes, want %d", ErrPepperInvalid, len(pepper), encx.PepperLength)
	}

	return &PepperStore{pepper: pepper}, nil
}

// PepperExists reports whether a pepper is configured. A store without one returns
// ErrPepperMissing rather than false, so encx never takes its "generate a new pepper" path.
func (s *PepperStore) PepperExists(_ context.Context, _ string) (bool, error) {
	if err := s.check(); err != nil {
		return false, err
	}
	return true, nil
}

// GetPepper returns a copy of the configured pepper.
func (s *PepperStore) GetPepper(_ context.Context, _ string) ([]byte, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	return append([]byte(nil), s.pepper...), nil
}

// StorePepper always fails with ErrPepperReadOnly.
func (s *PepperStore) StorePepper(_ context.Context, _ string, _ []byte) error {
	return ErrPepperReadOnly
}

// GetStoragePath names the variable the pepper is read from.
func (s *PepperStore) GetStoragePath(_ string) string {
	return "env:" + EnvPepper
}

func (s *PepperStore) check() error {
	if s == nil || len(s.pepper) == 0 {
		return ErrPepperMissing
	}
	if len(s.pepper) != encx.PepperLength {
		return fmt.Errorf("%w: %d bytes, want %d", ErrPepperInvalid, len(s.pepper), encx.PepperLength)
	}
	return nil
}
