package bookingRepository

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

// GetBookingsByAvailability retrieves all bookings for a specific availability.
// Used for slot overlap detection in the slot-based booking system.
// Returns bookings sorted by slot_start_time ASC.
func (r *Repository) GetBookingsByAvailability(ctx context.Context, availabilityID uuid.UUID) ([]*domain.BookingEncx, error) {
	return r.bookingsByAvailability(ctx, r.pool, availabilityID)
}

// bookingsByAvailability reads an availability's bookings through q.
func (r *Repository) bookingsByAvailability(ctx context.Context, q querier, availabilityID uuid.UUID) ([]*domain.BookingEncx, error) {
	return r.selectBookings(ctx, q, "get bookings by availability", `WHERE availability_id = $1`, availabilityID)
}

// occupyingBookingsByPartner reads a partner's bookings that may occupy time
// (every status but cancelled and expired) through q: the pool, or the
// transaction CreateGuarded holds. Slot times are encrypted, so the time
// filtering happens after decryption (Booking.OccupiesTime and the engine).
func (r *Repository) occupyingBookingsByPartner(ctx context.Context, q querier, partnerID uuid.UUID) ([]*domain.BookingEncx, error) {
	return r.selectBookings(ctx, q, "get occupying bookings by partner",
		`WHERE user_id = $1 AND status NOT IN ($2, $3)`,
		partnerID, domain.BookingStatusCancelled, domain.BookingStatusExpired)
}

// selectBookings reads bookings matching where (with its args) through q,
// sorted by slot_start_time_encrypted (only stable, not chronological).
func (r *Repository) selectBookings(ctx context.Context, q querier, operation, where string, args ...any) ([]*domain.BookingEncx, error) {
	query := `
		SELECT
			id,
			availability_id,
			client_id,
			user_id,
			room_id,
			total_price_cents,
			currency,
			payment_status,
			payment_intent_id,
			status,
			cancelled_at,
			completed_at,
			created_at,
			updated_at,
			-- Encrypted fields
			product_id_encrypted,
			slot_start_time_encrypted,
			slot_end_time_encrypted,
			client_notes_encrypted,
			partner_notes_encrypted,
			cancellation_reason_encrypted,
			guest_first_name_encrypted,
			guest_last_name_encrypted,
			guest_email_encrypted,
			guest_phone_encrypted,
			token,
			reminded_at,
			hold_expires_at,
			locale,
			-- Encryption metadata
			dek_encrypted,
			key_version,
			metadata
		FROM booking.bookings
		` + where + `
		ORDER BY slot_start_time_encrypted ASC
	`

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ClassifyPgError(operation, err)
	}
	defer rows.Close()

	var bookingsEncx []*domain.BookingEncx
	for rows.Next() {
		bookingEncx := &domain.BookingEncx{}

		err := rows.Scan(
			&bookingEncx.ID,
			&bookingEncx.AvailabilityID,
			&bookingEncx.ClientID,
			&bookingEncx.PartnerID,
			&bookingEncx.RoomID,
			&bookingEncx.TotalPriceCents,
			&bookingEncx.Currency,
			&bookingEncx.PaymentStatus,
			&bookingEncx.PaymentIntentID,
			&bookingEncx.Status,
			&bookingEncx.CancelledAt,
			&bookingEncx.CompletedAt,
			&bookingEncx.CreatedAt,
			&bookingEncx.UpdatedAt,
			// Encrypted fields
			&bookingEncx.ProductIDEncrypted,
			&bookingEncx.SlotStartTimeEncrypted,
			&bookingEncx.SlotEndTimeEncrypted,
			&bookingEncx.ClientNotesEncrypted,
			&bookingEncx.PartnerNotesEncrypted,
			&bookingEncx.CancellationReasonEncrypted,
			&bookingEncx.GuestFirstNameEncrypted,
			&bookingEncx.GuestLastNameEncrypted,
			&bookingEncx.GuestEmailEncrypted,
			&bookingEncx.GuestPhoneEncrypted,
			&bookingEncx.Token,
			&bookingEncx.RemindedAt,
			&bookingEncx.HoldExpiresAt,
			&bookingEncx.Locale,
			// Encryption metadata
			&bookingEncx.DEKEncrypted,
			&bookingEncx.KeyVersion,
			&bookingEncx.Metadata,
		)
		if err != nil {
			return nil, errs.ClassifyPgError("scan booking row", err)
		}

		bookingsEncx = append(bookingsEncx, bookingEncx)
	}

	if err := rows.Err(); err != nil {
		return nil, errs.ClassifyPgError("iterate booking rows", err)
	}

	// Return empty slice instead of nil if no bookings found
	if bookingsEncx == nil {
		return []*domain.BookingEncx{}, nil
	}

	return bookingsEncx, nil
}
