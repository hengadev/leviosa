package booking

import (
	"context"
	"fmt"
	"sync"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
)

// MockPaymentService is a simple mock implementation of PaymentService for testing
type MockPaymentService struct {
	mu             sync.Mutex // concurrent bookings create intents at once
	paymentIntents map[string]*ports.PaymentIntentInfo
	refunds        []MockRefund
	cancels        []string
}

// MockRefund is one refund the booking service asked the payment port for.
type MockRefund struct {
	PaymentIntentID string
	Amount          int
	Reason          string
}

// RefundsFor returns the refunds made on a PaymentIntent.
func (m *MockPaymentService) RefundsFor(paymentIntentID string) []MockRefund {
	m.mu.Lock()
	defer m.mu.Unlock()
	var refunds []MockRefund
	for _, refund := range m.refunds {
		if refund.PaymentIntentID == paymentIntentID {
			refunds = append(refunds, refund)
		}
	}
	return refunds
}

// CancelsOf counts the cancel calls made on a PaymentIntent.
func (m *MockPaymentService) CancelsOf(paymentIntentID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, id := range m.cancels {
		if id == paymentIntentID {
			count++
		}
	}
	return count
}

// IntentStatus returns a PaymentIntent's status.
func (m *MockPaymentService) IntentStatus(paymentIntentID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if intent, ok := m.paymentIntents[paymentIntentID]; ok {
		return intent.Status
	}
	return ""
}

// NewMockPaymentService creates a new mock payment service
func NewMockPaymentService() *MockPaymentService {
	return &MockPaymentService{
		paymentIntents: make(map[string]*ports.PaymentIntentInfo),
	}
}

// CreatePaymentIntent mocks creating a Stripe payment intent
func (m *MockPaymentService) CreatePaymentIntent(ctx context.Context, amount int, currency, description, receiptEmail string, metadata map[string]string) (string, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	paymentIntentID := "pi_test_" + uuid.New().String()[:8]
	clientSecret := "pi_test_secret_" + uuid.New().String()[:8]

	m.paymentIntents[paymentIntentID] = &ports.PaymentIntentInfo{
		ID:           paymentIntentID,
		Status:       ports.PaymentIntentStatusRequiresPaymentMethod,
		Amount:       amount,
		Currency:     currency,
		ClientSecret: clientSecret,
		Description:  description,
		Metadata:     metadata,
	}

	return paymentIntentID, clientSecret, nil
}

// ConfirmPaymentIntent mocks confirming a payment intent
func (m *MockPaymentService) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	intent, exists := m.paymentIntents[paymentIntentID]
	if !exists {
		return fmt.Errorf("payment intent not found: %s", paymentIntentID)
	}

	intent.Status = ports.PaymentIntentStatusSucceeded
	return nil
}

// RetrievePaymentIntent mocks retrieving a payment intent
func (m *MockPaymentService) RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (*ports.PaymentIntentInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	intent, exists := m.paymentIntents[paymentIntentID]
	if !exists {
		return nil, fmt.Errorf("payment intent not found: %s", paymentIntentID)
	}

	return intent, nil
}

// RefundPayment mocks creating a refund
func (m *MockPaymentService) RefundPayment(ctx context.Context, paymentIntentID string, amount int, reason string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	intent, exists := m.paymentIntents[paymentIntentID]
	if !exists {
		return "", fmt.Errorf("payment intent not found: %s", paymentIntentID)
	}

	if intent.Status != ports.PaymentIntentStatusSucceeded {
		return "", fmt.Errorf("payment intent must be succeeded to refund")
	}

	m.refunds = append(m.refunds, MockRefund{PaymentIntentID: paymentIntentID, Amount: amount, Reason: reason})
	refundID := "re_test_" + uuid.New().String()[:8]
	return refundID, nil
}

// CancelPaymentIntent mocks canceling a payment intent
func (m *MockPaymentService) CancelPaymentIntent(ctx context.Context, paymentIntentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	intent, exists := m.paymentIntents[paymentIntentID]
	if !exists {
		return fmt.Errorf("payment intent not found: %s", paymentIntentID)
	}

	m.cancels = append(m.cancels, paymentIntentID)
	intent.Status = ports.PaymentIntentStatusCanceled
	return nil
}

// VerifyWebhookSignature mocks verifying a webhook signature
func (m *MockPaymentService) VerifyWebhookSignature(payload []byte, signature string) (*ports.WebhookEvent, error) {
	// For testing purposes, always return a valid webhook event
	return &ports.WebhookEvent{
		ID:              "evt_test_" + uuid.New().String(),
		Type:            "payment_intent.succeeded",
		PaymentIntentID: "pi_test_" + uuid.New().String(),
		Status:          "succeeded",
		Amount:          1000,
		Currency:        "usd",
		Metadata:        make(map[string]string),
	}, nil
}
