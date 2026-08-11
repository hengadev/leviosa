package review

import (
	"context"
	"errors"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

// HideReview hides a review from public display. The route layer is responsible for
// restricting this to staff/admin (RequireAdmin) — this method performs no role
// check itself.
func (s *service) HideReview(ctx context.Context, reviewID uuid.UUID) (*domain.Review, error) {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, errs.ErrRepositoryNotFound
		}
		return nil, fmt.Errorf("get review to hide: %w", err)
	}

	if err := review.Hide(); err != nil {
		return nil, errs.NewUnprocessableEntityErr(err.Error())
	}

	if err := s.reviewRepo.Update(ctx, review); err != nil {
		return nil, fmt.Errorf("hide review: %w", err)
	}

	return review, nil
}

// UnhideReview reverses a previous HideReview call. The route layer is responsible
// for restricting this to staff/admin (RequireAdmin).
func (s *service) UnhideReview(ctx context.Context, reviewID uuid.UUID) (*domain.Review, error) {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, errs.ErrRepositoryNotFound
		}
		return nil, fmt.Errorf("get review to unhide: %w", err)
	}

	if err := review.Unhide(); err != nil {
		return nil, errs.NewUnprocessableEntityErr(err.Error())
	}

	if err := s.reviewRepo.Update(ctx, review); err != nil {
		return nil, fmt.Errorf("unhide review: %w", err)
	}

	return review, nil
}
