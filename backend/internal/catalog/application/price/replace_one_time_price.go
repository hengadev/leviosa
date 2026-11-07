package price

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

// ReplaceOneTimePrice changes a product's price: a Stripe price cannot be
// edited, so a new one-time price is created and the product's other active
// one-time prices in that currency are deactivated. Booking then reads the
// new one; the old ones stay for the bookings and payments that refer to them.
//
// The new price is created first: if a deactivation then fails, the product
// still has a price, and the error says which one is left active.
func (s *PriceService) ReplaceOneTimePrice(ctx context.Context, productIDStr string, request *domain.CreatePriceRequest) (string, error) {
	if request.Interval != string(domain.OneTime) {
		return "", errs.NewInvalidValueErr("interval: a product's session price is one_time")
	}

	previous, err := s.repo.GetPricesByProductID(ctx, productIDStr, true)
	if err != nil && !errors.Is(err, errs.ErrRepositoryNotFound) {
		return "", fmt.Errorf("get the product's active prices: %w", err)
	}

	newID, err := s.CreatePrice(ctx, productIDStr, request)
	if err != nil {
		return "", err
	}

	currency := strings.ToUpper(strings.TrimSpace(request.Currency))
	for _, p := range previous {
		if !p.IsActive || p.Interval != domain.OneTime || !strings.EqualFold(p.Currency, currency) {
			continue
		}
		if err := s.DeactivatePrice(ctx, p.ID.String()); err != nil {
			return newID, fmt.Errorf("new price %s created, but price %s is still active: %w", newID, p.ID, err)
		}
	}
	return newID, nil
}
