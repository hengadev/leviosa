package reviewRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

func (r *Repository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Review, error) {
	query := fmt.Sprintf(`
		SELECT id, booking_id, client_id, partner_id, product_id, rating, comment, hidden, created_at, updated_at
		FROM %s.reviews
		WHERE booking_id = $1
	`, r.schema)

	review := &domain.Review{}
	var comment *string
	err := r.pool.QueryRow(ctx, query, bookingID).Scan(
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
	)
	if err != nil {
		return nil, errs.ClassifyPgError("get review by booking id", err)
	}
	if comment != nil {
		review.Comment = *comment
	}

	return review, nil
}
