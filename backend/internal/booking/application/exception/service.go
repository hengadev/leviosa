// Package exception is the partner area's management of availability
// exceptions (docs/issues/012-availability-exceptions.md): closing a
// recurring availability on a date or a range of dates, listing and
// reopening them. The slot engine skips the closed occurrences
// (domain.WithoutExceptions), for the practice page and booking creation
// alike.
package exception

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
	"github.com/hengadev/encx"
)

type Service struct {
	exceptionRepo    ports.AvailabilityExceptionRepository
	availabilityRepo ports.AvailabilityRepository
	bookingRepo      ports.BookingRepository
	crypto           encx.CryptoService
	now              func() time.Time
}

// New creates the availability exception service.
func New(
	exceptionRepo ports.AvailabilityExceptionRepository,
	availabilityRepo ports.AvailabilityRepository,
	bookingRepo ports.BookingRepository,
	crypto encx.CryptoService,
) ports.AvailabilityExceptionService {
	return &Service{
		exceptionRepo:    exceptionRepo,
		availabilityRepo: availabilityRepo,
		bookingRepo:      bookingRepo,
		crypto:           crypto,
		now:              time.Now,
	}
}

func (s *Service) ListExceptions(ctx context.Context, partnerUserID, availabilityID uuid.UUID) ([]domain.AvailabilityException, error) {
	if _, err := s.ownedAvailability(ctx, partnerUserID, availabilityID); err != nil {
		return nil, err
	}
	exceptions, err := s.exceptionRepo.ListByAvailabilityIDs(ctx, []uuid.UUID{availabilityID})
	if err != nil {
		return nil, fmt.Errorf("list availability exceptions: %w", err)
	}
	return exceptions, nil
}

// CreateException closes the dates and counts the bookings already on them.
// Those bookings are left as they are: the partner cancels them one by one
// (practitioner cancellation) if needed.
func (s *Service) CreateException(ctx context.Context, partnerUserID, availabilityID uuid.UUID, startDate, endDate time.Time) (*domain.AvailabilityException, int, error) {
	availability, err := s.ownedAvailability(ctx, partnerUserID, availabilityID)
	if err != nil {
		return nil, 0, err
	}

	now := s.now()
	exception, err := domain.NewAvailabilityException(availability, startDate, endDate, domain.PracticeToday(now))
	if err != nil {
		return nil, 0, errs.NewInvalidValueErr(err.Error())
	}
	if err := s.exceptionRepo.Create(ctx, exception); err != nil {
		return nil, 0, fmt.Errorf("create availability exception: %w", err)
	}

	affected, err := s.countBookingsOn(ctx, *exception, now)
	if err != nil {
		return nil, 0, err
	}
	return exception, affected, nil
}

func (s *Service) DeleteException(ctx context.Context, partnerUserID, availabilityID, exceptionID uuid.UUID) error {
	if _, err := s.ownedAvailability(ctx, partnerUserID, availabilityID); err != nil {
		return err
	}
	exception, err := s.exceptionRepo.GetByID(ctx, exceptionID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return fmt.Errorf("availability exception not found: %w", errs.ErrRepositoryNotFound)
		}
		return fmt.Errorf("get availability exception: %w", err)
	}
	if exception.AvailabilityID != availabilityID {
		return fmt.Errorf("availability exception not found: %w", errs.ErrRepositoryNotFound)
	}
	if err := s.exceptionRepo.Delete(ctx, exceptionID); err != nil {
		return fmt.Errorf("delete availability exception: %w", err)
	}
	return nil
}

// ownedAvailability returns the availability if it belongs to the calling
// partner. Another partner's availability answers not found, so its
// existence is not disclosed.
func (s *Service) ownedAvailability(ctx context.Context, partnerUserID, availabilityID uuid.UUID) (*domain.Availability, error) {
	availabilityEncx, err := s.availabilityRepo.GetByID(ctx, availabilityID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("availability not found: %w", errs.ErrRepositoryNotFound)
		}
		return nil, fmt.Errorf("get availability: %w", err)
	}
	if availabilityEncx.UserID != partnerUserID {
		return nil, fmt.Errorf("availability not found: %w", errs.ErrRepositoryNotFound)
	}
	availability, err := domain.DecryptAvailabilityEncx(ctx, s.crypto, availabilityEncx)
	if err != nil {
		return nil, errs.NewNotDecryptedErr("availability", err)
	}
	return availability, nil
}

// countBookingsOn counts the availability's bookings starting on the closed
// dates (practice time zone) that hold their time: confirmed ones, and those
// awaiting payment while their hold runs.
func (s *Service) countBookingsOn(ctx context.Context, exception domain.AvailabilityException, now time.Time) (int, error) {
	bookingsEncx, err := s.bookingRepo.GetBookingsByAvailability(ctx, exception.AvailabilityID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return 0, nil
		}
		return 0, fmt.Errorf("list bookings of availability %s: %w", exception.AvailabilityID, err)
	}
	count := 0
	for _, bookingEncx := range bookingsEncx {
		booking, err := domain.DecryptBookingEncx(ctx, s.crypto, bookingEncx)
		if err != nil {
			return 0, fmt.Errorf("decrypt booking %s: %w", bookingEncx.ID, err)
		}
		holdsTime := booking.Status == domain.BookingStatusConfirmed ||
			(booking.Status == domain.BookingStatusAwaitingPayment && booking.OccupiesTime(now))
		if holdsTime && exception.CoversInstant(booking.SlotStartTime) {
			count++
		}
	}
	return count, nil
}
