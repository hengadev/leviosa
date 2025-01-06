package productService

import "context"

type Reader interface {
	GetProduct(ctx context.Context, productID string) (*Product, error)
	GetProductType(ctx context.Context, productID int) (*ProductType, error)
}

type Writer interface {
	AddProduct(ctx context.Context, product *Product) error
	AddProductType(ctx context.Context, productType *ProductType) error
	AddPriceID(ctx context.Context, productID string, priceID string) error
	ModifyProduct(ctx context.Context, product *Product, whereMap map[string]any, prohibitedFields ...string) error
	ModifyProductType(ctx context.Context, productType *ProductType, whereMap map[string]any, prohibitedFields ...string) error
	RemoveProduct(ctx context.Context, productID string) error
	RemoveProductType(ctx context.Context, productID int) error
}

type ReadWriter interface {
	Reader
	Writer
}
