package booking

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// Next working day (issue 037): the first Paris date after today on which
// the practitioner has an open availability occurrence or an appointment
// booking, and that day's appointments, so Aujourd'hui can show it.

// minimumNextWorkingDayLookAhead bounds the domain function's scan from
// below: even a short booking horizon is looked across for 60 days, so a
// weekly rhythm always finds its next day.
const minimumNextWorkingDayLookAhead = 60

// GetPartnerNextWorkingDay returns the partner's next working day and its
// appointment bookings (the same shape as the partner bookings list). The
// date is empty when no day is worked within the look-ahead.
func (s *BookingService) GetPartnerNextWorkingDay(ctx context.Context, partnerID uuid.UUID) (*domain.PartnerNextWorkingDayResponse, error) {
	if s.authUserClient == nil {
		return nil, fmt.Errorf("get next working day: no authuser client configured")
	}
	partnerUserID, err := s.authUserClient.GetPartnerUserID(ctx, partnerID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("partner not found: %w", errs.ErrRepositoryNotFound)
		}
		return nil, fmt.Errorf("resolve partner: %w", err)
	}

	policy, err := s.bookingPolicy(ctx)
	if err != nil {
		return nil, err
	}
	lookAhead := int(policy.Horizon.Hours() / 24)
	if lookAhead < minimumNextWorkingDayLookAhead {
		lookAhead = minimumNextWorkingDayLookAhead
	}

	today := s.now().In(domain.PracticeLocation)
	from := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, domain.PracticeLocation).AddDate(0, 0, 1)
	to := from.AddDate(0, 0, lookAhead)

	availabilitiesEncx, err := s.availabilityRepo.GetByUserID(ctx, partnerUserID, ports.AvailabilityFilter{
		Status:    []domain.AvailabilityStatus{domain.AvailabilityStatusAvailable},
		TimeRange: &ports.TimeRange{Start: from, End: to},
	})
	if err != nil && !errors.Is(err, errs.ErrRepositoryNotFound) {
		return nil, fmt.Errorf("list partner availabilities: %w", err)
	}
	availabilities := make([]*domain.Availability, 0, len(availabilitiesEncx))
	for _, availabilityEncx := range availabilitiesEncx {
		availability, err := domain.DecryptAvailabilityEncx(ctx, s.crypto, availabilityEncx)
		if err != nil {
			return nil, errs.NewNotDecryptedErr("availability", err)
		}
		availabilities = append(availabilities, availability)
	}
	exceptions, err := s.exceptionsOf(ctx, availabilities...)
	if err != nil {
		return nil, err
	}

	// The bookings decide too: an appointment on a day without availability
	// makes it a working day. TimeRange is not applied by the repository
	// (slot times are encrypted); the domain function keeps the days.
	bookings, err := s.GetPartnerBookings(ctx, partnerID, ports.BookingFilter{})
	if err != nil {
		return nil, err
	}

	result := &domain.PartnerNextWorkingDayResponse{Bookings: []domain.PartnerBookingResponse{}}
	day, ok := domain.NextWorkingDay(availabilities, exceptions, bookings, today, lookAhead)
	if !ok {
		return result, nil
	}
	result.Date = day.Format("2006-01-02")

	dayBookings, err := s.GetPartnerBookingsEnriched(ctx, partnerID, ports.BookingFilter{})
	if err != nil {
		return nil, err
	}
	dayStart := civilParis(day)
	for _, b := range dayBookings {
		if b.Status != domain.BookingStatusCancelled && civilParis(b.SlotStartTime.In(domain.PracticeLocation)) == dayStart {
			result.Bookings = append(result.Bookings, b)
		}
	}
	sort.Slice(result.Bookings, func(i, j int) bool {
		return result.Bookings[i].SlotStartTime.Before(result.Bookings[j].SlotStartTime)
	})
	return result, nil
}

// civilParis returns the YYYY-MM-DD date of t in the practice time zone.
func civilParis(t time.Time) string {
	return t.Format("2006-01-02")
}
