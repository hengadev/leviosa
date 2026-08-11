package reviewRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

func (r *Repository) AggregateByPartnerID(ctx context.Context, partnerID uuid.UUID) (float64, int, error) {
	return r.aggregateByColumn(ctx, "partner_id", partnerID)
}

func (r *Repository) AggregateByProductID(ctx context.Context, productID uuid.UUID) (float64, int, error) {
	return r.aggregateByColumn(ctx, "product_id", productID)
}

// aggregateByColumn computes the average rating and count across all non-hidden
// reviews matching the given column/value. COALESCE guards against NULL from AVG()
// when there are zero matching rows.
func (r *Repository) aggregateByColumn(ctx context.Context, column string, value uuid.UUID) (float64, int, error) {
	query := fmt.Sprintf(`
		SELECT COALESCE(AVG(rating), 0), COUNT(*)
		FROM %s.reviews
		WHERE %s = $1 AND hidden = FALSE
	`, r.schema, column)

	var average float64
	var count int
	if err := r.pool.QueryRow(ctx, query, value).Scan(&average, &count); err != nil {
		return 0, 0, errs.ClassifyPgError("aggregate reviews by "+column, err)
	}

	return average, count, nil
}
