package productService

import "context"

type Reader interface {
	GetProduct(ctx context.Context, productID string) (*Product, error)
	GetOffer(ctx context.Context, productID int) (*Offer, error)
}

type Writer interface {
	AddProduct(ctx context.Context, product *Product) error
	AddOffer(ctx context.Context, productType *Offer) error
	AddPriceID(ctx context.Context, productID string, priceID string) error
	ModifyProduct(ctx context.Context, product *Product, whereMap map[string]any, prohibitedFields ...string) error
	ModifyOffer(ctx context.Context, productType *Offer, whereMap map[string]any, prohibitedFields ...string) error
	RemoveProduct(ctx context.Context, productID string) error
	RemoveOffer(ctx context.Context, productID int) error
}

type ReadWriter interface {
	Reader
	Writer
}
