package booking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	catalogDomain "github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// Free slots (docs/issues/010-slot-engine.md): the public slot endpoint and
// booking creation both build the slot engine's query here, so the practice
// page and the server can never disagree on what is bookable.
//
// This slice covers one-off availabilities. Recurring availabilities (issue
// 011) will be expanded into occurrences and availability exceptions (012)
// subtracted before the windows reach the engine. Bookings awaiting payment
// occupy time while their hold runs (Booking.OccupiesTime, ADR-0018).

// GetFreeSlots returns the free slots of a partner's one-off availabilities
// for a product, starting in [from, to).
func (s *BookingService) GetFreeSlots(ctx context.Context, partnerID, productID uuid.UUID, from, to time.Time) (*domain.FreeSlots, error) {
	if !from.Before(to) {
		return nil, errs.NewInvalidValueErr("from must be before to")
	}
	if s.authUserClient == nil {
		return nil, fmt.Errorf("get free slots: no authuser client configured")
	}

	partnerUserID, err := s.authUserClient.GetPartnerUserID(ctx, partnerID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("partner not found: %w", errs.ErrRepositoryNotFound)
		}
		return nil, fmt.Errorf("resolve partner: %w", err)
	}

	product, err := s.productService.GetProductByID(ctx, productID.String())
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("product not found: %w", errs.ErrRepositoryNotFound)
		}
		return nil, fmt.Errorf("get product: %w", err)
	}

	policy, err := s.bookingPolicy(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	// Nothing can be offered outside the policy window: narrow the range so
	// the repository never scans beyond it.
	earliest := now.Add(policy.MinimumNotice)
	horizonEnd := now.Add(policy.Horizon)
	result := &domain.FreeSlots{Slots: []domain.Slot{}, EarliestStart: earliest, HorizonEnd: horizonEnd}
	if from.Before(earliest) {
		from = earliest
	}
	if limit := horizonEnd.Add(time.Nanosecond); to.After(limit) {
		to = limit
	}
	if !from.Before(to) {
		return result, nil
	}

	availabilitiesEncx, err := s.availabilityRepo.GetByUserID(ctx, partnerUserID, ports.AvailabilityFilter{
		Status:    []domain.AvailabilityStatus{domain.AvailabilityStatusAvailable},
		TimeRange: &ports.TimeRange{Start: from, End: to},
	})
	if err != nil && !errors.Is(err, errs.ErrRepositoryNotFound) {
		return nil, fmt.Errorf("list partner availabilities: %w", err)
	}

	availabilities := make([]*domain.Availability, 0, len(availabilitiesEncx))
	for _, availabilityEncx := range availabilitiesEncx {
		if availabilityEncx.IsRecurring {
			continue // recurring availabilities are expanded in issue 011
		}
		availability, err := domain.DecryptAvailabilityEncx(ctx, s.crypto, availabilityEncx)
		if err != nil {
			return nil, errs.NewNotDecryptedErr("availability", err)
		}
		availabilities = append(availabilities, availability)
	}

	query, err := s.slotQuery(ctx, availabilities, product, policy, now)
	if err != nil {
		return nil, err
	}
	query.From = from
	query.To = to

	result.Slots = append(result.Slots, domain.ComputeSlots(query)...)
	return result, nil
}

// slotQuery builds the engine's query over the given availabilities: their
// windows, the bookings occupying them (each extended by its own product's
// buffer time), the product and the policy. From/To are left to the caller.
func (s *BookingService) slotQuery(
	ctx context.Context,
	availabilities []*domain.Availability,
	product *catalogDomain.ProductRes,
	policy domain.BookingPolicy,
	now time.Time,
) (domain.SlotEngineQuery, error) {
	query := domain.SlotEngineQuery{
		ProductDuration: time.Duration(product.Duration) * time.Minute,
		BufferTime:      time.Duration(product.BufferTime) * time.Minute,
		MinimumNotice:   policy.MinimumNotice,
		BookingHorizon:  policy.Horizon,
		Now:             now,
	}

	buffers := map[uuid.UUID]time.Duration{product.ID: query.BufferTime}
	for _, availability := range availabilities {
		query.Availabilities = append(query.Availabilities, domain.AvailabilityWindow{
			AvailabilityID: availability.ID,
			Start:          availability.StartTime,
			End:            availability.EndTime,
		})

		bookingsEncx, err := s.bookingRepo.GetBookingsByAvailability(ctx, availability.ID)
		if err != nil && !errors.Is(err, errs.ErrRepositoryNotFound) {
			return domain.SlotEngineQuery{}, fmt.Errorf("list bookings of availability %s: %w", availability.ID, err)
		}
		occupancies, err := s.occupancies(ctx, bookingsEncx, buffers, query.BufferTime, now)
		if err != nil {
			return domain.SlotEngineQuery{}, err
		}
		query.Occupancies = append(query.Occupancies, occupancies...)
	}
	return query, nil
}

// occupancies decrypts bookings and returns the intervals of those occupying
// time at now (confirmed ones, and those awaiting payment while their hold
// runs), each extended by its own product's buffer time.
func (s *BookingService) occupancies(
	ctx context.Context,
	bookingsEncx []*domain.BookingEncx,
	buffers map[uuid.UUID]time.Duration,
	fallbackBuffer time.Duration,
	now time.Time,
) ([]domain.SlotOccupancy, error) {
	var occupancies []domain.SlotOccupancy
	for _, bookingEncx := range bookingsEncx {
		booking, err := domain.DecryptBookingEncx(ctx, s.crypto, bookingEncx)
		if err != nil {
			return nil, fmt.Errorf("decrypt existing booking %s: %w", bookingEncx.ID, err)
		}
		if !booking.OccupiesTime(now) {
			continue
		}
		occupancies = append(occupancies, booking.Occupancy(s.bufferOf(ctx, booking.ProductID, buffers, fallbackBuffer)))
	}
	return occupancies, nil
}

// bufferOf returns the buffer time of a booked product, memoised per query.
// A product that can no longer be read (removed from the catalog) falls back
// to the queried product's buffer rather than failing the whole query.
func (s *BookingService) bufferOf(ctx context.Context, productID uuid.UUID, cache map[uuid.UUID]time.Duration, fallback time.Duration) time.Duration {
	if buffer, ok := cache[productID]; ok {
		return buffer
	}
	buffer := fallback
	if product, err := s.productService.GetProductByID(ctx, productID.String()); err == nil {
		buffer = time.Duration(product.BufferTime) * time.Minute
	} else {
		slog.WarnContext(ctx, "booked product unreadable; using the queried product's buffer time",
			"product_id", productID, "err", err)
	}
	cache[productID] = buffer
	return buffer
}

// bookingPolicy reads the current policy; without a provider (tests, tools)
// the defaults apply.
func (s *BookingService) bookingPolicy(ctx context.Context) (domain.BookingPolicy, error) {
	if s.policyProvider == nil {
		return domain.DefaultBookingPolicy(), nil
	}
	policy, err := s.policyProvider.GetBookingPolicy(ctx)
	if err != nil {
		return domain.BookingPolicy{}, fmt.Errorf("get booking policy: %w", err)
	}
	return policy, nil
}
