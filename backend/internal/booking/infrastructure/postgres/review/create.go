package reviewRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

func (r *Repository) Create(ctx context.Context, review *domain.Review) error {
	query := fmt.Sprintf(`
		INSERT INTO %s.reviews (
			id, booking_id, client_id, partner_id, product_id,
			rating, comment, hidden, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`, r.schema)

	_, err := r.pool.Exec(ctx, query,
		review.ID,
		review.BookingID,
		review.ClientID,
		review.PartnerID,
		review.ProductID,
		review.Rating,
		review.Comment,
		review.Hidden,
		review.CreatedAt,
		review.UpdatedAt,
	)
	if err != nil {
		return errs.ClassifyPgError("create review", err)
	}

	return nil
}
