package productPayment

import (
	"context"
	"fmt"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v87"
)

func (c *service) CreateProduct(ctx context.Context, req domain.CreateStripeProductRequest) (*domain.PaymentProduct, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	// One key per creation, so stripe-go's retries of this request are not
	// applied twice. Stripe caps keys at 255 characters: the key must not grow
	// with the product's name or description.
	idempotencyKey := "prod_" + uuid.NewString()

	params := &stripe.ProductCreateParams{
		Name:        stripe.String(req.Name),
		Description: stripe.String(req.Description),
		Active:      stripe.Bool(true),
		Metadata:    req.Metadata,
	}
	params.SetIdempotencyKey(idempotencyKey)

	stripeProduct, err := c.V1Products.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe product creation failed: %w", err)
	}

	var createdAt time.Time
	if stripeProduct.Created > 0 {
		t := time.Unix(stripeProduct.Created, 0)
		createdAt = t
	}

	var updatedAt time.Time
	if stripeProduct.Updated > 0 {
		updatedAt = time.Unix(stripeProduct.Updated, 0)
	}

	return &domain.PaymentProduct{
		ID:          stripeProduct.ID,
		Name:        stripeProduct.Name,
		Description: stripeProduct.Description,
		Active:      stripeProduct.Active,
		Metadata:    stripeProduct.Metadata,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
