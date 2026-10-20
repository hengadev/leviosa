package encryption_test

import (
	"context"
	"log"
	"os"
	"testing"

	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
)

var kmsContainer *tu.KMSContainer

// TestMain starts one LocalStack KMS container for the suite; each test creates its own
// key and alias so tests never share a KEK.
func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	kmsContainer, err = tu.SetupKMS(ctx, nil)
	if err != nil {
		log.Fatalf("setup LocalStack KMS: %v", err)
	}

	code := m.Run()
	tu.TeardownKMS(ctx, nil, kmsContainer)
	os.Exit(code)
}
