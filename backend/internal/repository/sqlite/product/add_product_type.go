package productRepository

import (
	"context"
	"errors"

	productService "github.com/GaryHY/event-reservation-app/internal/domain/product"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (p *Repository) AddProductType(ctx context.Context, productType *productService.ProductType) error {
	query := `
        INSERT INTO product_types (
            name,
            description
        ) VALUES (?, ?);`
	result, err := p.DB.ExecContext(
		ctx,
		query,
		productType.Name,
		productType.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}

	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotCreatedErr(errors.New("no rows affected by insertion statement"), "product type")
	}
	return nil
}
