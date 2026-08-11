package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// GetBillingPortalLink generates a Stripe Billing Portal session URL for the authenticated
// user to manage their saved payment method. The user must already have a Stripe customer ID
// (set when the account was completed, or via OAuth signup) — Billing Portal sessions can only
// be created for an existing Stripe customer, so this does not create one on the fly.
func (s *UserService) GetBillingPortalLink(ctx context.Context, userID uuid.UUID, returnURL string) (string, error) {
	// Get encrypted user from repository
	userEncx, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get user by ID: %w", err)
	}

	// Decrypt user
	user, err := domain.DecryptUserEncx(ctx, s.crypto, userEncx)
	if err != nil {
		return "", errs.NewNotDecryptedErr("user", err)
	}

	if user.StripeCustomerID == "" {
		return "", errs.NewNotFoundErr(errors.New("user has no associated stripe customer"), "stripe customer")
	}

	// Create the Billing Portal session
	url, err := s.stripe.CreateBillingPortalSession(ctx, user.StripeCustomerID, returnURL)
	if err != nil {
		return "", fmt.Errorf("create billing portal session: %w", err)
	}

	return url, nil
}
