package authPayment

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/stripe/stripe-go/v82"
)

// CreateBillingPortalSession creates a Stripe Billing Portal session scoped to the given
// customer, for payment-method management. The customer is redirected to returnURL when
// they're done.
func (s *service) CreateBillingPortalSession(ctx context.Context, customerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	session, err := s.client.V1BillingPortalSessions.Create(ctx, params)
	if err != nil {
		return "", errs.ClassifyStripeError("create billing portal session", err)
	}

	return session.URL, nil
}
