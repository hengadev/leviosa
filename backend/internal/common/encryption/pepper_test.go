package encryption

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/hengadev/encx"
)

var _ encx.SecretManagementService = (*PepperStore)(nil)

func validPepper() []byte {
	return []byte("0123456789abcdef0123456789abcdef")
}

func TestNewPepperStore_ServesConfiguredPepper(t *testing.T) {
	ctx := context.Background()
	store, err := NewPepperStore(base64.StdEncoding.EncodeToString(validPepper()))
	if err != nil {
		t.Fatalf("NewPepperStore: %v", err)
	}

	exists, err := store.PepperExists(ctx, "leviosa")
	if err != nil || !exists {
		t.Fatalf("PepperExists = %v, %v; want true, nil", exists, err)
	}

	got, err := store.GetPepper(ctx, "leviosa")
	if err != nil {
		t.Fatalf("GetPepper: %v", err)
	}
	if !bytes.Equal(got, validPepper()) {
		t.Fatalf("GetPepper = %q, want %q", got, validPepper())
	}

	// The caller must not be able to alter the stored pepper through the returned slice.
	got[0] ^= 0xff
	again, _ := store.GetPepper(ctx, "leviosa")
	if !bytes.Equal(again, validPepper()) {
		t.Fatal("GetPepper returned the store's own slice")
	}
}

func TestNewPepperStore_TrimsSurroundingWhitespace(t *testing.T) {
	encoded := "  " + base64.StdEncoding.EncodeToString(validPepper()) + "\n"
	if _, err := NewPepperStore(encoded); err != nil {
		t.Fatalf("NewPepperStore with surrounding whitespace: %v", err)
	}
}

func TestNewPepperStore_RejectsMissingPepper(t *testing.T) {
	for _, encoded := range []string{"", "   "} {
		_, err := NewPepperStore(encoded)
		if !errors.Is(err, ErrPepperMissing) {
			t.Errorf("NewPepperStore(%q) error = %v, want ErrPepperMissing", encoded, err)
		}
	}
}

func TestNewPepperStore_RejectsInvalidBase64(t *testing.T) {
	_, err := NewPepperStore("not base64 at all!")
	if !errors.Is(err, ErrPepperInvalid) {
		t.Fatalf("error = %v, want ErrPepperInvalid", err)
	}
}

func TestNewPepperStore_RejectsWrongLength(t *testing.T) {
	for _, n := range []int{16, 31, 33, 64} {
		_, err := NewPepperStore(base64.StdEncoding.EncodeToString(make([]byte, n)))
		if !errors.Is(err, ErrPepperInvalid) {
			t.Errorf("%d-byte pepper: error = %v, want ErrPepperInvalid", n, err)
		}
	}
}

func TestPepperStore_RefusesToStorePepper(t *testing.T) {
	store, err := NewPepperStore(base64.StdEncoding.EncodeToString(validPepper()))
	if err != nil {
		t.Fatalf("NewPepperStore: %v", err)
	}

	err = store.StorePepper(context.Background(), "leviosa", make([]byte, encx.PepperLength))
	if !errors.Is(err, ErrPepperReadOnly) {
		t.Fatalf("StorePepper error = %v, want ErrPepperReadOnly", err)
	}

	got, _ := store.GetPepper(context.Background(), "leviosa")
	if !bytes.Equal(got, validPepper()) {
		t.Fatal("StorePepper changed the configured pepper")
	}
}

// A zero-value store holds no pepper: every read path must fail rather than let encx
// fall through to generating (and trying to store) a new one.
func TestPepperStore_ZeroValueFailsEveryReadPath(t *testing.T) {
	var store PepperStore
	ctx := context.Background()

	if _, err := store.PepperExists(ctx, "leviosa"); !errors.Is(err, ErrPepperMissing) {
		t.Errorf("PepperExists error = %v, want ErrPepperMissing", err)
	}
	if _, err := store.GetPepper(ctx, "leviosa"); !errors.Is(err, ErrPepperMissing) {
		t.Errorf("GetPepper error = %v, want ErrPepperMissing", err)
	}
	if err := store.StorePepper(ctx, "leviosa", make([]byte, encx.PepperLength)); !errors.Is(err, ErrPepperReadOnly) {
		t.Errorf("StorePepper error = %v, want ErrPepperReadOnly", err)
	}
}

func TestPepperStore_StoragePathNamesTheVariable(t *testing.T) {
	var store PepperStore
	if got := store.GetStoragePath("leviosa"); got != "env:"+EnvPepper {
		t.Fatalf("GetStoragePath = %q, want %q", got, "env:"+EnvPepper)
	}
}
