package bookingRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

func (r *Repository) Create(ctx context.Context, bookingEncx *domain.BookingEncx) error {
	return r.insert(ctx, r.pool, bookingEncx)
}

// insert writes a new booking through exec: the pool, or the transaction
// CreateGuarded holds.
func (r *Repository) insert(ctx context.Context, exec execer, bookingEncx *domain.BookingEncx) error {
	query := fmt.Sprintf(`
		INSERT INTO %s.bookings (
			id, availability_id, client_id, user_id, room_id,
			product_id_encrypted, slot_start_time_encrypted, slot_end_time_encrypted,
			client_notes_encrypted, partner_notes_encrypted,
			total_price_cents, currency, payment_status, payment_intent_id,
			status, cancelled_at, cancellation_reason_encrypted, completed_at,
			guest_first_name_encrypted, guest_last_name_encrypted,
			guest_email_encrypted, guest_phone_encrypted,
			token,
			reminded_at,
			created_at, updated_at,
			dek_encrypted, key_version, metadata,
			hold_expires_at, locale, amount_paid_online_cents
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22,
			$23,
			$24,
			$25, $26, $27, $28, $29,
			$30, $31, $32
		)
	`, r.schema)

	_, err := exec.Exec(ctx, query,
		bookingEncx.ID,
		bookingEncx.AvailabilityID,
		bookingEncx.ClientID,
		bookingEncx.PartnerID,
		bookingEncx.RoomID,
		bookingEncx.ProductIDEncrypted,
		bookingEncx.SlotStartTimeEncrypted,
		bookingEncx.SlotEndTimeEncrypted,
		bookingEncx.ClientNotesEncrypted,
		bookingEncx.PartnerNotesEncrypted,
		bookingEncx.TotalPriceCents,
		bookingEncx.Currency,
		bookingEncx.PaymentStatus,
		bookingEncx.PaymentIntentID,
		bookingEncx.Status,
		bookingEncx.CancelledAt,
		bookingEncx.CancellationReasonEncrypted,
		bookingEncx.CompletedAt,
		bookingEncx.GuestFirstNameEncrypted,
		bookingEncx.GuestLastNameEncrypted,
		bookingEncx.GuestEmailEncrypted,
		bookingEncx.GuestPhoneEncrypted,
		bookingEncx.Token,
		bookingEncx.RemindedAt,
		bookingEncx.CreatedAt,
		bookingEncx.UpdatedAt,
		bookingEncx.DEKEncrypted,
		bookingEncx.KeyVersion,
		bookingEncx.Metadata,
		bookingEncx.HoldExpiresAt,
		bookingEncx.Locale,
		bookingEncx.AmountPaidOnlineCents,
	)
	if err != nil {
		return errs.ClassifyPgError("create booking", err)
	}

	return nil
}
