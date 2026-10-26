package stripe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

		svc := NewService("sk_test_dummy", server.URL, "whsec_dummy")

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

		svc := NewService("sk_test_dummy", server.URL, "whsec_dummy")

		_, _, err := svc.CreatePaymentIntent(context.Background(), 2000, "eur", "Guestless booking payment", "", nil)

		require.NoError(t, err)
		assert.False(t, sawReceiptEmail, "expected receipt_email to be omitted when not provided")
	})
}
