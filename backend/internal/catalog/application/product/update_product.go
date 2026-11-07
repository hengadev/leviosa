package product

import (
	"context"
	"fmt"
	"log"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

func (s *ProductService) UpdateProduct(ctx context.Context, productIDStr string, product *domain.UpdateProductRequest) error {
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return errs.NewInvalidValueErr("product ID is required and must be a valid UUID")
	}

	existingProduct, err := s.sharedRepo.GetProductByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("get existing product for update: %w", err)
	}

	if err := product.Valid(ctx); err != nil {
		return errs.NewInvalidValueErr("product")
	}

	// A new deposit must fit every active price of the product (issue 013).
	if product.Deposit.Present && product.Deposit.Value != nil {
		if err := s.validateDepositAgainstActivePrices(ctx, productID, product.Deposit.Value); err != nil {
			return err
		}
	}

	stripeUpdateNeeded := product.Name != nil || product.Description != nil
	if stripeUpdateNeeded {
		stripeReq := &domain.UpdateStripeProductRequest{
			Name:        product.Name,
			Description: product.Description,
		}
		// The repository call needs the Stripe product ID, which we got from the DB.
		if _, err := s.stripe.UpdateProduct(ctx, existingProduct.StripeProductID, stripeReq); err != nil {
			return errs.NewExternalServiceErr(err, "failed to update product in Stripe")
		}
	}

	if err := s.repo.UpdateProduct(ctx, productID, product); err != nil {
		// 5. This is the rollback step: If the DB update fails, revert the Stripe update.
		log.Printf("Service: Failed to update product %s in DB, attempting Stripe rollback: %v", productID, err)

		// The rollback payload uses the original values.
		if stripeUpdateNeeded {
			rollbackStripeReq := &domain.UpdateStripeProductRequest{
				Name:        &existingProduct.Name,
				Description: &existingProduct.Description,
			}
			if _, rollbackErr := s.stripe.UpdateProduct(ctx, existingProduct.StripeProductID, rollbackStripeReq); rollbackErr != nil {
				log.Printf("Service: Failed to rollback Stripe product %s. Data inconsistency detected! Rollback error: %v", existingProduct.StripeProductID, rollbackErr)
			}
		}

		// Return the original database error, wrapped.
		return fmt.Errorf("update product: %w", err)
	}

	return nil
}

// validateDepositAgainstActivePrices rejects a deposit that does not fit one
// of the product's active prices (a fixed deposit above a price). A product
// without active prices has nothing to check against yet; adding a price
// later re-validates the deposit (PriceService.CreatePrice).
func (s *ProductService) validateDepositAgainstActivePrices(ctx context.Context, productID uuid.UUID, deposit *domain.Deposit) error {
	amounts, err := s.sharedRepo.GetActivePriceAmounts(ctx, productID)
	if err != nil {
		return fmt.Errorf("get active prices to validate the deposit: %w", err)
	}
	for _, amount := range amounts {
		if err := deposit.ValidateAgainstPrice(amount); err != nil {
			return errs.NewInvalidValueErr(fmt.Sprintf("deposit: %s", err.Error()))
		}
	}
	return nil
}
