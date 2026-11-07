package price

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// CreatePrice creates a new Stripe Price and stores its details locally.
func (s *PriceService) CreatePrice(ctx context.Context, productIDStr string, request *domain.CreatePriceRequest) (string, error) {
	// 1. Validate internal product ID
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return "", errs.NewInvalidValueErr(fmt.Sprintf("internal product ID is invalid: %s", err.Error()))
	}

	// 2. Validate incoming price data (errs.CreatePriceInput is analogous to errs.Price here)
	// You might define a `CreatePriceInput` struct in `errs` if different from `Price` directly.
	// For simplicity, let's assume `input` has relevant fields for `Price.Valid` check.
	// A new `errs.Price` will be constructed for `Valid` call.
	tempPrice := domain.Price{
		ProductID: productID,
		Amount:    request.Amount,
		// Currencies are stored in uppercase ("EUR"), the code booking looks up.
		Currency:       strings.ToUpper(strings.TrimSpace(request.Currency)),
		Interval:       domain.Interval(request.Interval),
		TaxRatePercent: request.TaxRatePercent,
	}

	if err := tempPrice.Valid(ctx); err != nil {
		return "", errs.NewInvalidValueErr(fmt.Errorf("price validation failed: %w", err).Error())
	}

	// A new price must still cover the product's fixed deposit (issue 013).
	// Stripe price amounts are immutable, so this is the only way a price
	// changes under an existing deposit.
	product, err := s.sharedRepo.GetProductByID(ctx, productID)
	if err != nil {
		return "", fmt.Errorf("get product: %w", err)
	}
	if err := product.Deposit.ValidateAgainstPrice(request.Amount); err != nil {
		return "", errs.NewInvalidValueErr(fmt.Sprintf("price does not cover the product's deposit: %s", err.Error()))
	}

	// 3. Get Stripe Product ID for the given internal product ID.
	stripeProductID, _, err := s.sharedRepo.GetStripeProductAndPriceIDs(ctx, productID)
	if err != nil {
		return "", fmt.Errorf("get Stripe Product ID: %w", err)
	}

	// 4. Create Price in Stripe.
	stripeReq := domain.CreateStripePriceRequest{
		ProductID: stripeProductID, // Use Stripe's product ID
		Amount:    int64(request.Amount),
		Currency:  tempPrice.Currency,
		Interval:  request.Interval,
		Active:    true, // Default to active when creating prices
		Nickname:  "",   // Set from request.Nickname if needed
		Metadata:  request.Metadata,
	}

	// Set optional fields if provided
	if request.Nickname != nil {
		stripeReq.Nickname = *request.Nickname
	}
	// paymentPrice, err := s.stripe.CreatePrice(ctx, stripeReq)
	paymentPrice, err := s.stripe.CreatePrice(ctx, stripeReq)
	if err != nil {
		return "", errs.NewExternalServiceErr(fmt.Errorf("failed to create price in Stripe: %w", err), "Stripe API")
	}

	// 5. Store new Price details in your database. The currency and interval
	// are the ones asked for: Stripe answers with a lowercase currency, and
	// with no interval at all for a one-time price.
	now := time.Now()
	newPrice := &domain.Price{
		ID:             uuid.New(),      // Generate a new internal ID
		ProductID:      productID,       // Your internal product ID
		StripePriceID:  paymentPrice.ID, // Stripe's price ID
		Amount:         int(paymentPrice.Amount),
		Currency:       tempPrice.Currency,
		Interval:       tempPrice.Interval,
		IsActive:       paymentPrice.Active,
		TaxRatePercent: request.TaxRatePercent, // Local-only field; not synced to Stripe
		CreatedAt:      now,                    // DB will set this, but for returning, use now or Stripe's created time
		UpdatedAt:      now,
	}

	if err := s.repo.CreatePrice(ctx, newPrice); err != nil {
		// IMPORTANT: If DB creation fails, consider rolling back Stripe price creation
		// or flagging it for manual review/cleanup. This adds complexity.
		// For now, we'll just return the error.
		return "", errs.NewQueryFailedErr(fmt.Errorf("failed to save price to database: %w", err))
	}

	return newPrice.ID.String(), nil
}
