package productRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/product"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (p *Repository) AddProduct(ctx context.Context, product *productService.Product) error {
	query := `
            INSERT INTO products (
                id,
                price,
                name,
                description,
                picture,
                type,
            ) VALUES (?, ?, ?, ?, ?, ?)`

	result, err := p.DB.ExecContext(ctx, query,
		product.ID,
		product.Price,
		product.Name,
		product.Description,
		product.Picture,
		product.Type,
	)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}

	// Check if the insert was successful
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotCreatedErr(err, "product")
	}
	return nil
}
