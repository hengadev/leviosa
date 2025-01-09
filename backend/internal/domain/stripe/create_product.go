package stripeService

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/product"
)

// CreateProduct create a stripe product with a generated price using stripe's Product API.
func (s *Service) CreateProduct(ctx context.Context, object Payment) (string, error) {
	paymentInfo := object.GetPaymentInfo(ctx)
	product_params := &stripe.ProductParams{
		ID:          stripe.String(paymentInfo.ID),
		Name:        stripe.String(paymentInfo.Name),
		Description: stripe.String(paymentInfo.Description),
	}
	product, err := product.New(product_params)
	if err != nil {
		return "", domain.NewNotCreatedErr(err)
	}
	return product.ID, nil
}
