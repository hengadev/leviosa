package stripeService

import "context"

type PaymentProductInfo struct {
	ID          string
	Name        string
	Description string
	Price       int64
	PriceID     string
}

// the interface that I need to implement for all object that can be bought
type Payment interface {
	GetPaymentInfo(ctx context.Context) *PaymentProductInfo
}
