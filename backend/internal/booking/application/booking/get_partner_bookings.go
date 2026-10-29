package booking

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
)

func (s *BookingService) GetPartnerBookingsEnriched(ctx context.Context, partnerID uuid.UUID, filter ports.BookingFilter) ([]domain.PartnerBookingResponse, error) {
	bookings, err := s.GetPartnerBookings(ctx, partnerID, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]domain.PartnerBookingResponse, 0, len(bookings))
	for _, b := range bookings {
		resp := domain.PartnerBookingResponse{
			ID:              b.ID,
			ClientID:        b.ClientID,
			ClientName:      "Utilisateur inconnu",
			ProductName:     "Produit inconnu",
			RoomName:        "Salle inconnue",
			SlotStartTime:   b.SlotStartTime,
			SlotEndTime:     b.SlotEndTime,
			Status:          b.Status,
			PaymentStatus:   b.PaymentStatus,
			TotalPriceCents: b.TotalPriceCents,
			Currency:        b.Currency,
			CompletedAt:     b.CompletedAt,
		}

		// Contact details (name, email, phone) only — never notes: Leviosa
		// stores no health data (ADR-0019).
		if b.IsGuestBooking() {
			resp.ClientName = b.GuestDisplayName()
			resp.ClientEmail = b.GuestEmail
			resp.ClientPhone = b.GuestPhone
		} else if s.authUserClient != nil {
			name, email, phone, err := s.authUserClient.GetUserContact(ctx, *b.ClientID)
			if err == nil {
				if name != "" {
					resp.ClientName = name
				}
				resp.ClientEmail = email
				resp.ClientPhone = phone
			}
		}
		if s.productService != nil {
			resp.ProductName = s.resolveProductName(ctx, b.ProductID)
		}
		if s.roomService != nil {
			resp.RoomName = s.resolveRoomName(ctx, b.RoomID)
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// GetPartnerBookings returns the partner's appointments: bookings awaiting
// payment or expired are never shown to the partner (ADR-0018).
func (s *BookingService) GetPartnerBookings(ctx context.Context, partnerID uuid.UUID, filter ports.BookingFilter) ([]*domain.Booking, error) {
	filter.ExcludeStatus = domain.NonAppointmentStatuses()
	bookingsEncx, err := s.bookingRepo.GetByPartnerID(ctx, partnerID, filter)
	if err != nil {
		return nil, fmt.Errorf("get partner bookings: %w", err)
	}

	// Decrypt each booking
	bookings := make([]*domain.Booking, 0, len(bookingsEncx))
	for _, bookingEncx := range bookingsEncx {
		booking, err := domain.DecryptBookingEncx(ctx, s.crypto, bookingEncx)
		if err != nil {
			return nil, fmt.Errorf("decrypt booking %s: %w", bookingEncx.ID, err)
		}
		bookings = append(bookings, booking)
	}

	return bookings, nil
}
