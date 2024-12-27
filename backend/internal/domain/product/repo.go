package productService

import "context"

type Reader interface{}
type Writer interface {
	AddProduct(ctx context.Context, product *Product) error
	AddPriceID(ctx context.Context, productID string, priceID string) error
}

type ReadWriter interface {
	Reader
	Writer
}
