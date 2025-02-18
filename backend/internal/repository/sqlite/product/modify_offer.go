package productRepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain/product"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/pkg/sqliteutil"
)

func (p *Repository) ModifyOffer(
	ctx context.Context,
	productType *productService.Offer,
	whereMap map[string]any,
) error {
	query, values, errs := sqliteutil.WriteUpdateQuery(*productType, whereMap)
	if len(errs) > 0 {
		return rp.NewInternalErr(errs)
	}
	result, err := p.DB.ExecContext(ctx, query, values...)
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
		return rp.NewNotUpdatedErr(err, fmt.Sprintf("product named %s", productType.Name))
	}
	return nil
}
