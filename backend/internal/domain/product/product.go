package productService

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/stripe"
	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

// (massage standard, massage hibiscus etc...)
type Product struct {
	ID          string // this is like the productID in the products table
	Price       int64
	PriceID     string
	Name        string
	Description string
	Picture     string
	Type        ProductType
}

func (p *Product) GetPaymentInfo(ctx context.Context) *stripeService.PaymentProductInfo {
	return &stripeService.PaymentProductInfo{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}
}

func (p Product) Valid(ctx context.Context) errsx.Map {
	var errs errsx.Map
	if p.Price <= 0 {
		errs.Set("price", "price value should be >= 0")
	}
	return errs
}

// - create product (service to create object + store in database)
// - get the *Product
// - then get the Payment info from the *Product
// - use the stripe service to create a new stripe product with *Product and get the stripe PriceID
// - store the priceID in the database
