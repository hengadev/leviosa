package stripe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gostripe "github.com/stripe/stripe-go/v82"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stripeGoLibrary points the stripe-go SDK's global API backend at an
// httptest server for the duration of a test, then restores the previous
// backend/key.
//
// The Service in this package (and the wider Stripe SDK v82 usage here) calls
// package-level helpers like paymentintent.New/Get, which read the SDK's
// process-global backend and API key rather than any per-Service client
// configuration. Redirecting that global backend is therefore the only way to
// exercise CreatePaymentIntent/RetrievePaymentIntent against a fake server in
// a unit test.
func withStripeTestBackend(t *testing.T, serverURL string) {
	t.Helper()

	prevKey := gostripe.Key
	prevBackend := gostripe.GetBackend(gostripe.APIBackend)

	gostripe.Key = "sk_test_dummy"
	backend := gostripe.GetBackendWithConfig(gostripe.APIBackend, &gostripe.BackendConfig{
		URL: gostripe.String(serverURL),
	})
	gostripe.SetBackend(gostripe.APIBackend, backend)

	t.Cleanup(func() {
		gostripe.Key = prevKey
		gostripe.SetBackend(gostripe.APIBackend, prevBackend)
	})
}

// TestCreatePaymentIntent_SetsReceiptEmail verifies that CreatePaymentIntent
// sends the receipt_email parameter to Stripe when one is provided, so Stripe
// can email the client a hosted receipt automatically once payment succeeds.
func TestCreatePaymentIntent_SetsReceiptEmail(t *testing.T) {
	t.Run("includes receipt_email when provided", func(t *testing.T) {
		var gotReceiptEmail string
		var sawReceiptEmail bool

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())
			sawReceiptEmail = r.Form.Has("receipt_email")
			gotReceiptEmail = r.Form.Get("receipt_email")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "pi_test_123",
				"object":        "payment_intent",
				"client_secret": "pi_test_123_secret_abc",
				"status":        "requires_payment_method",
				"amount":        1000,
				"currency":      "eur",
			})
		}))
		defer server.Close()

		withStripeTestBackend(t, server.URL)

		svc := NewService("sk_test_dummy", "", "whsec_dummy")

		id, secret, err := svc.CreatePaymentIntent(context.Background(), 1000, "eur", "Booking payment", "client@example.com", nil)

		require.NoError(t, err)
		assert.Equal(t, "pi_test_123", id)
		assert.Equal(t, "pi_test_123_secret_abc", secret)
		assert.True(t, sawReceiptEmail, "expected receipt_email to be sent to Stripe")
		assert.Equal(t, "client@example.com", gotReceiptEmail)
	})

	t.Run("omits receipt_email when not provided", func(t *testing.T) {
		var sawReceiptEmail bool

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())
			sawReceiptEmail = r.Form.Has("receipt_email")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "pi_test_456",
				"object":        "payment_intent",
				"client_secret": "pi_test_456_secret_abc",
				"status":        "requires_payment_method",
				"amount":        2000,
				"currency":      "eur",
			})
		}))
		defer server.Close()

		withStripeTestBackend(t, server.URL)

		svc := NewService("sk_test_dummy", "", "whsec_dummy")

		_, _, err := svc.CreatePaymentIntent(context.Background(), 2000, "eur", "Guestless booking payment", "", nil)

		require.NoError(t, err)
		assert.False(t, sawReceiptEmail, "expected receipt_email to be omitted when not provided")
	})
}
