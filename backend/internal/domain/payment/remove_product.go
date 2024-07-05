package payment

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/product"
)

func (s *Service) RemoveProduct(ctx context.Context, eventID string) (string, error) {
	product_params := &stripe.ProductParams{
		ID:          &eventID,
		Description: stripe.String("1 X Pass valuable for all the event."),
	}
	_, err := product.Del(eventID, product_params)
	if err != nil {
		return "", fmt.Errorf("Failed to delete the product on the server: %w", err)
	}
	return eventID, nil
}
