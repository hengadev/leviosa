package productRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/product"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (p *Repository) GetProduct(ctx context.Context, productID string) (*productService.Product, error) {
	var product productService.Product
	query := `
        SELECT 
            price,
            price_id,
            name,
            description,
            picture,
            type
        FROM products
        WHERE id = ?;`
	err := p.DB.QueryRowContext(ctx, query, productID).Scan(
		&product.Price,
		&product.PriceID,
		&product.Name,
		&product.Description,
		&product.Picture,
		&product.Type,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, rp.NewNotFoundErr(err, "pending user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return &product, nil
}
