package sharedRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// GetActivePriceAmounts returns the amounts, in cents, of a product's active
// prices. A product without active prices yields an empty slice, not an
// error. The product service uses it to check a new fixed deposit against
// every price the product can currently be sold at (issue 013).
func (r *SharedRepository) GetActivePriceAmounts(ctx context.Context, productID uuid.UUID) ([]int, error) {
	query := fmt.Sprintf(`
	SELECT amount
	FROM %s.prices
	WHERE product_id = $1 AND is_active = TRUE;`, r.schema)

	rows, err := r.pool.Query(ctx, query, productID)
	if err != nil {
		return nil, errs.ClassifyPgError(fmt.Sprintf("get active price amounts for product %s", productID), err)
	}
	defer rows.Close()

	amounts := []int{}
	for rows.Next() {
		var amount int
		if err := rows.Scan(&amount); err != nil {
			return nil, errs.NewDBQueryErr(fmt.Errorf("failed to scan price amount for product %s: %w", productID, err))
		}
		amounts = append(amounts, amount)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.NewDBQueryErr(fmt.Errorf("error iterating over price amounts for product %s: %w", productID, err))
	}
	return amounts, nil
}
