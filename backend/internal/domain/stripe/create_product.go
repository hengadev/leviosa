package stripeService

import (
	"fmt"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/price"
	"github.com/stripe/stripe-go/v79/product"
)

type PaymentProductInfo struct {
	ID          string
	Name        string
	Description string
	Price       int64
}

type Payment interface {
	GetPaymentInfo() *PaymentProductInfo
}

func (s *Service) CreateProduct(v Payment) (string, error) {
	p := v.GetPaymentInfo()
	product_params := &stripe.ProductParams{
		ID:          &p.ID,
		Name:        stripe.String(p.Name),
		Description: stripe.String(p.Description),
	}
	product, err := product.New(product_params)
	if err != nil {
		return "", fmt.Errorf("Failed to create new product on server: %w ", err)
	}
	price_params := &stripe.PriceParams{
		Currency:   stripe.String(string(stripe.CurrencyEUR)),
		Product:    stripe.String(product.ID),
		UnitAmount: stripe.Int64(p.Price),
	}
	price, err := price.New(price_params)
	if err != nil {
		return "", fmt.Errorf("Failed to create new price on server: %w ", err)
	}
	return price.ID, nil
}
