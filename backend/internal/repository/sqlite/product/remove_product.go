package productRepository

import (
	"context"
	"errors"

	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (p *Repository) RemoveProduct(ctx context.Context, productID string) error {
	result, err := p.DB.ExecContext(ctx, "DELETE FROM users WHERE id = ?;", productID)
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
		return rp.NewNotDeletedErr(err, "product")
	}
	return nil
}
