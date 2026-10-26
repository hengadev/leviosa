package stripe

import (
	"context"
	"log"
	"os"
	"testing"

	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stripeMockContainer is shared across this package's tests that need a real
// (mocked) Stripe API server, per ADR-0005: Stripe is an external paid API,
// so tests run against stripe/stripe-mock rather than the live test-mode API.
var stripeMockContainer *tu.StripeMockContainer

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tu.SetupStripeMock(ctx, nil)
	if err != nil {
		log.Fatalf("failed to start stripe-mock container: %v", err)
	}
	stripeMockContainer = container

	code := m.Run()

	tu.TeardownStripeMock(ctx, nil, stripeMockContainer)
	os.Exit(code)
}

// TestRetrievePaymentIntent_ReturnsReceiptURL exercises RetrievePaymentIntent
// against a real (mocked) Stripe API server and verifies that the underlying
// charge's hosted receipt_url is surfaced on the returned PaymentIntentInfo,
// via the "latest_charge" expand added to the retrieval call.
func TestRetrievePaymentIntent_ReturnsReceiptURL(t *testing.T) {
	svc := NewService("sk_test_dummy", stripeMockContainer.URL, "whsec_dummy")

	info, err := svc.RetrievePaymentIntent(context.Background(), "pi_test_receipt")

	require.NoError(t, err)
	assert.Equal(t, "pi_test_receipt", info.ID)
	assert.NotEmpty(t, info.ReceiptURL, "expected receipt_url to be populated from the expanded latest_charge")
}
