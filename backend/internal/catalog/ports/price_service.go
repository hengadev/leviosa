package ports

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
)

type PriceService interface {
	PrivatePriceService
	PublicPriceService
	// GetPublishedProductPrices serves the public price route: the active
	// prices of a published product, not found for a draft or archived one.
	GetPublishedProductPrices(ctx context.Context, productID string) ([]*domain.Price, error)
}

// PublicPriceService contains read-only methods for price lookups.
// Used by other services (e.g., booking) for cross-module queries.
type PublicPriceService interface {
	GetPrice(ctx context.Context, priceID string) (*domain.Price, error)
	GetPriceByStripeID(ctx context.Context, stripePriceID string) (*domain.Price, error)
	GetAllPrices(ctx context.Context) ([]*domain.Price, error)
	GetPricesByProductID(ctx context.Context, productID string) ([]*domain.Price, error)
	GetActiveOneTimePriceByProductID(ctx context.Context, productID string, currency string) (*domain.Price, error)
}

// PrivatePriceService contains write methods for price management.
// Used by catalog handlers for admin operations.
type PrivatePriceService interface {
	CreatePrice(ctx context.Context, productID string, request *domain.CreatePriceRequest) (string, error)
	UpdatePrice(ctx context.Context, priceID string, input domain.UpdatePriceRequest) (*domain.Price, error)
	DeactivatePrice(ctx context.Context, priceID string) error
	// ReplaceOneTimePrice changes a product's one-time price: a new price,
	// the previous active ones in that currency deactivated.
	ReplaceOneTimePrice(ctx context.Context, productID string, request *domain.CreatePriceRequest) (string, error)
}
