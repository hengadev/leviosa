package waitlistRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WaitlistEntryEncx, error) {
	query := fmt.Sprintf(`
		SELECT
			id, client_id,
			product_id_encrypted, desired_range_start_encrypted, desired_range_end_encrypted,
			notes_encrypted,
			guest_first_name_encrypted, guest_last_name_encrypted,
			guest_email_encrypted, guest_phone_encrypted,
			status, notified_at, claimed_at, expired_at, cancelled_at,
			created_at, updated_at,
			dek_encrypted, key_version, metadata
		FROM %s.waitlist_entries
		WHERE id = $1
	`, r.schema)

	entryEncx := &domain.WaitlistEntryEncx{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&entryEncx.ID,
		&entryEncx.ClientID,
		&entryEncx.ProductIDEncrypted,
		&entryEncx.DesiredRangeStartEncrypted,
		&entryEncx.DesiredRangeEndEncrypted,
		&entryEncx.NotesEncrypted,
		&entryEncx.GuestFirstNameEncrypted,
		&entryEncx.GuestLastNameEncrypted,
		&entryEncx.GuestEmailEncrypted,
		&entryEncx.GuestPhoneEncrypted,
		&entryEncx.Status,
		&entryEncx.NotifiedAt,
		&entryEncx.ClaimedAt,
		&entryEncx.ExpiredAt,
		&entryEncx.CancelledAt,
		&entryEncx.CreatedAt,
		&entryEncx.UpdatedAt,
		&entryEncx.DEKEncrypted,
		&entryEncx.KeyVersion,
		&entryEncx.Metadata,
	)
	if err != nil {
		return nil, errs.ClassifyPgError("get waitlist entry by id", err)
	}

	return entryEncx, nil
}
