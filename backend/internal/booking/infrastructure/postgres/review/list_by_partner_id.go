package reviewRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

func (r *Repository) ListByPartnerID(ctx context.Context, partnerID uuid.UUID, filter ports.ReviewFilter) ([]*domain.Review, error) {
	return r.listByColumn(ctx, "partner_id", partnerID, filter)
}

func (r *Repository) listByColumn(ctx context.Context, column string, value uuid.UUID, filter ports.ReviewFilter) ([]*domain.Review, error) {
	query := fmt.Sprintf(`
		SELECT id, booking_id, client_id, partner_id, product_id, rating, comment, hidden, created_at, updated_at
		FROM %s.reviews
		WHERE %s = $1
	`, r.schema, column)

	args := []interface{}{value}
	argIndex := 2

	if !filter.IncludeHidden {
		query += " AND hidden = FALSE"
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ClassifyPgError("list reviews by "+column, err)
	}
	defer rows.Close()

	var reviews []*domain.Review
	for rows.Next() {
		review := &domain.Review{}
		var comment *string
		if err := rows.Scan(
			&review.ID,
			&review.BookingID,
			&review.ClientID,
			&review.PartnerID,
			&review.ProductID,
			&review.Rating,
			&comment,
			&review.Hidden,
			&review.CreatedAt,
			&review.UpdatedAt,
		); err != nil {
			return nil, errs.ClassifyPgError("scan review row", err)
		}
		if comment != nil {
			review.Comment = *comment
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, errs.ClassifyPgError("iterate review rows", err)
	}

	return reviews, nil
}
