package productRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/product"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (p *Repository) GetProductType(ctx context.Context, productID int) (*productService.ProductType, error) {
	var productType productService.ProductType
	query := `
        SELECT
            name,
            description
        FROM product_types
        WHERE id = ?;`
	err := p.DB.QueryRowContext(ctx, query, productID).Scan(
		&productType.Name,
		&productType.Description,
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
	return &productType, nil
}
