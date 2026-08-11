package ports

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/google/uuid"
)

// ReviewService defines the review sub-module's use cases (issue 009).
type ReviewService interface {
	// CreateReview submits a review for a booking. Enforces, mirroring the
	// booking-relationship authorization pattern from ADR-0013:
	//   - the booking exists (errs.ErrRepositoryNotFound otherwise)
	//   - clientID is the booking's client (errs.ErrForbidden otherwise)
	//   - the booking status is `completed` (errs.ErrUnprocessableEntity otherwise)
	//   - at most one review per booking (errs.ErrConflict otherwise)
	CreateReview(ctx context.Context, bookingID, clientID uuid.UUID, rating int, comment string) (*domain.Review, error)

	// GetPublicReviewsForPartner returns non-hidden reviews and an aggregate rating
	// for a partner, computed on read.
	GetPublicReviewsForPartner(ctx context.Context, partnerID uuid.UUID, limit, offset int) (*domain.PublicReviewListResponse, error)

	// GetPublicReviewsForProduct returns non-hidden reviews and an aggregate rating
	// for a product, computed on read.
	GetPublicReviewsForProduct(ctx context.Context, productID uuid.UUID, limit, offset int) (*domain.PublicReviewListResponse, error)

	// HideReview hides a review from public display (staff/admin only — enforced at
	// the route layer via RequireAdmin). The record is retained, not deleted.
	HideReview(ctx context.Context, reviewID uuid.UUID) (*domain.Review, error)

	// UnhideReview reverses a previous HideReview call.
	UnhideReview(ctx context.Context, reviewID uuid.UUID) (*domain.Review, error)
}
