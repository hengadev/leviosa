package reviewRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

func (r *Repository) Update(ctx context.Context, review *domain.Review) error {
	query := fmt.Sprintf(`
		UPDATE %s.reviews
		SET rating = $1, comment = $2, hidden = $3, updated_at = $4
		WHERE id = $5
	`, r.schema)

	tag, err := r.pool.Exec(ctx, query,
		review.Rating,
		review.Comment,
		review.Hidden,
		review.UpdatedAt,
		review.ID,
	)
	if err != nil {
		return errs.ClassifyPgError("update review", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update review: %w", errs.ErrRepositoryNotFound)
	}

	return nil
}
