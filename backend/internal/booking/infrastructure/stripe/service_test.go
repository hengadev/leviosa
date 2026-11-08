package stripe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v87/webhook"
)

// TestCreatePaymentIntent_SetsReceiptEmail verifies that CreatePaymentIntent
// sends the receipt_email parameter to Stripe when one is provided, so Stripe
// can email the client a hosted receipt automatically once payment succeeds.
func TestCreatePaymentIntent_SetsReceiptEmail(t *testing.T) {
	t.Run("includes receipt_email when provided", func(t *testing.T) {
		var gotReceiptEmail, gotAmount, gotCurrency string
		var sawReceiptEmail bool

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())
			sawReceiptEmail = r.Form.Has("receipt_email")
			gotReceiptEmail = r.Form.Get("receipt_email")
			gotAmount = r.Form.Get("amount")
			gotCurrency = r.Form.Get("currency")

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

		id, secret, err := svc.CreatePaymentIntent(context.Background(), 1000, "EUR", "Booking payment", "client@example.com", nil)

		require.NoError(t, err)
		assert.Equal(t, "1000", gotAmount, "the PaymentIntent charges the amount given (the amount due online)")
		assert.Equal(t, "eur", gotCurrency, "Stripe currency codes are lowercase")
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

// TestVerifyWebhookSignature_AcceptsTheEndpointAPIVersion pins the API
// version of the staging and production webhook endpoints. stripe-go rejects
// an event from another release line (basil, dahlia, ...) before any
// handling, so a stripe-go major bump that leaves endive fails here instead
// of leaving every paid booking unconfirmed.
func TestVerifyWebhookSignature_AcceptsTheEndpointAPIVersion(t *testing.T) {
	const secret = "whsec_test"
	payload := []byte(`{
		"id": "evt_test_1",
		"object": "event",
		"api_version": "2026-09-30.endive",
		"type": "payment_intent.succeeded",
		"data": {"object": {
			"id": "pi_test_1",
			"object": "payment_intent",
			"status": "succeeded",
			"amount": 2000,
			"currency": "eur",
			"metadata": {"booking_id": "6b38f5c4-4673-4c70-b2e8-bf797857b63d"}
		}}
	}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: secret})

	event, err := NewService("sk_test", "", secret).VerifyWebhookSignature(signed.Payload, signed.Header)
	require.NoError(t, err)
	assert.Equal(t, "pi_test_1", event.PaymentIntentID)
	assert.Equal(t, "succeeded", event.Status)
	assert.Equal(t, 2000, event.Amount)
	assert.Equal(t, "6b38f5c4-4673-4c70-b2e8-bf797857b63d", event.Metadata["booking_id"])
}
