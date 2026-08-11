package ports

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/google/uuid"
)

// ReviewRepository defines persistence for reviews. Reviews are not encx-encrypted
// (see domain.Review) so this repository works with plain domain.Review values,
// unlike most booking repositories which work with *Encx DTOs.
type ReviewRepository interface {
	// Create stores a new review. Returns an error wrapping errs.ErrUniqueViolation
	// if a review already exists for the booking (enforced by a DB unique constraint
	// on booking_id as a defense-in-depth backstop to the application-layer check).
	Create(ctx context.Context, review *domain.Review) error

	// GetByID retrieves a review by its ID. Returns errs.ErrRepositoryNotFound if absent.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)

	// GetByBookingID retrieves the review for a given booking, if one exists.
	// Returns errs.ErrRepositoryNotFound if absent.
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Review, error)

	// ListByPartnerID retrieves reviews for a partner, paginated and optionally
	// filtered to non-hidden only.
	ListByPartnerID(ctx context.Context, partnerID uuid.UUID, filter ReviewFilter) ([]*domain.Review, error)

	// ListByProductID retrieves reviews for a product, same filtering semantics.
	ListByProductID(ctx context.Context, productID uuid.UUID, filter ReviewFilter) ([]*domain.Review, error)

	// AggregateByPartnerID computes the average rating and count across ALL
	// non-hidden reviews for a partner (not just a paginated page).
	AggregateByPartnerID(ctx context.Context, partnerID uuid.UUID) (average float64, count int, err error)

	// AggregateByProductID computes the average rating and count across ALL
	// non-hidden reviews for a product.
	AggregateByProductID(ctx context.Context, productID uuid.UUID) (average float64, count int, err error)

	// Update persists changes to an existing review (used for hide/unhide moderation).
	Update(ctx context.Context, review *domain.Review) error
}

// ReviewFilter controls pagination and visibility filtering for review listings.
type ReviewFilter struct {
	// IncludeHidden, when false (the default), restricts results to non-hidden
	// reviews — used by the public read endpoints. Staff/admin moderation views
	// pass true to see everything.
	IncludeHidden bool
	Limit         int
	Offset        int
}
