package review

import (
	"context"
	"errors"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

// CreateReview submits a review for a booking. See ports.ReviewService for the
// authorization rules enforced here, which mirror ADR-0013's booking-relationship
// pattern but additionally require the booking to be `completed` specifically.
func (s *service) CreateReview(ctx context.Context, bookingID, clientID uuid.UUID, rating int, comment string) (*domain.Review, error) {
	booking, err := s.bookingSvc.GetBooking(ctx, bookingID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, errs.ErrRepositoryNotFound
		}
		return nil, fmt.Errorf("get booking for review: %w", err)
	}

	if booking.ClientID == nil || *booking.ClientID != clientID {
		return nil, errs.NewForbiddenErr(domain.ErrReviewNotBookingOwner.Error())
	}

	if booking.Status != domain.BookingStatusCompleted {
		return nil, errs.NewUnprocessableEntityErr(domain.ErrBookingNotCompleted.Error())
	}

	existing, err := s.reviewRepo.GetByBookingID(ctx, bookingID)
	if err != nil && !errors.Is(err, errs.ErrRepositoryNotFound) {
		return nil, fmt.Errorf("check existing review: %w", err)
	}
	if existing != nil {
		return nil, errs.NewConflictErr(domain.ErrReviewAlreadyExists)
	}

	review, err := domain.NewReview(bookingID, clientID, booking.PartnerID, booking.ProductID, rating, comment)
	if err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		if errors.Is(err, errs.ErrUniqueViolation) {
			return nil, errs.NewConflictErr(domain.ErrReviewAlreadyExists)
		}
		return nil, fmt.Errorf("create review: %w", err)
	}

	return review, nil
}
