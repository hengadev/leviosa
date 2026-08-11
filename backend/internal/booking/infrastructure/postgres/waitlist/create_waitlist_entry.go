package waitlistRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

func (r *Repository) Create(ctx context.Context, entry *domain.WaitlistEntryEncx) error {
	query := fmt.Sprintf(`
		INSERT INTO %s.waitlist_entries (
			id, client_id,
			product_id_encrypted, desired_range_start_encrypted, desired_range_end_encrypted,
			notes_encrypted,
			guest_first_name_encrypted, guest_last_name_encrypted,
			guest_email_encrypted, guest_phone_encrypted,
			status, notified_at, claimed_at, expired_at, cancelled_at,
			created_at, updated_at,
			dek_encrypted, key_version, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20
		)
	`, r.schema)

	_, err := r.pool.Exec(ctx, query,
		entry.ID,
		entry.ClientID,
		entry.ProductIDEncrypted,
		entry.DesiredRangeStartEncrypted,
		entry.DesiredRangeEndEncrypted,
		entry.NotesEncrypted,
		entry.GuestFirstNameEncrypted,
		entry.GuestLastNameEncrypted,
		entry.GuestEmailEncrypted,
		entry.GuestPhoneEncrypted,
		entry.Status,
		entry.NotifiedAt,
		entry.ClaimedAt,
		entry.ExpiredAt,
		entry.CancelledAt,
		entry.CreatedAt,
		entry.UpdatedAt,
		entry.DEKEncrypted,
		entry.KeyVersion,
		entry.Metadata,
	)
	if err != nil {
		return errs.ClassifyPgError("create waitlist entry", err)
	}

	return nil
}
