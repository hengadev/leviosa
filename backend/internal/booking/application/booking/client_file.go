package booking

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

// GetClientFile returns the client file of the given booking's client, as
// seen by the booking's partner (issue 042): contact details plus upcoming
// and past sessions with that practitioner. Computed from the partner's
// bookings; nothing is stored (ADR-0019: no notes, no health data).
func (s *BookingService) GetClientFile(ctx context.Context, partnerID, bookingID uuid.UUID) (*domain.ClientFileResponse, error) {
	booking, err := s.GetBooking(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.PartnerID != partnerID {
		return nil, errs.ErrRepositoryNotFound
	}

	bookings, err := s.GetPartnerBookingsEnriched(ctx, partnerID, ports.BookingFilter{})
	if err != nil {
		return nil, err
	}
	for _, g := range domain.GroupClientBookings(bookings) {
		for _, b := range g.Bookings {
			if b.ID == bookingID {
				file := domain.BuildClientFile(g, s.now())
				return &file, nil
			}
		}
	}
	return nil, errs.ErrRepositoryNotFound
}

// SearchPartnerClients returns the distinct clients of the practitioner
// matching a name or phone fragment (issue 043 wires this into the booking
// form).
func (s *BookingService) SearchPartnerClients(ctx context.Context, partnerID uuid.UUID, query string) ([]domain.ClientSummary, error) {
	bookings, err := s.GetPartnerBookingsEnriched(ctx, partnerID, ports.BookingFilter{})
	if err != nil {
		return nil, fmt.Errorf("search partner clients: %w", err)
	}
	return domain.SearchClientSummaries(domain.GroupClientBookings(bookings), query), nil
}
