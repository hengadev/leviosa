package priceRepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

// UpdatePrice updates specific fields of a Price in the database.
// This function takes a domain.Price struct, assuming it's already validated
// and contains the fields to update.
func (r *PriceRepository) UpdatePrice(ctx context.Context, priceID string, patch *domain.UpdatePriceRequest) error {
	// Dynamically build the SET clause and arguments
	sets := []string{}
	args := []any{}
	argCounter := 1

	if patch.Active != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", argCounter))
		args = append(args, *patch.Active)
		argCounter++
	}

	if patch.TaxRatePercent != nil {
		sets = append(sets, fmt.Sprintf("tax_rate_percent = $%d", argCounter))
		args = append(args, *patch.TaxRatePercent)
		argCounter++
	}
	// 'is_active' and 'tax_rate_percent' are the only fields updatable
	// directly on a price locally. If you allowed nickname/metadata updates
	// locally, you'd add them here.

	if len(sets) == 0 {
		return errs.NewInvalidInputErr(errors.New("no updatable fields provided for price update"))
	}

	query := fmt.Sprintf("UPDATE %s.prices SET %s WHERE id = $%d;",
		r.schema, strings.Join(sets, ", "), argCounter)
	args = append(args, priceID)

	commandTag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return errs.ClassifyPgError("update price", err)
	}

	if commandTag.RowsAffected() == 0 {
		return errs.NewRepositoryNotFoundErr(nil, "price") // Price not found by ID
	}

	return nil
}
