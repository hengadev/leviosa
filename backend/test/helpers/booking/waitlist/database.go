package waitlistHelpers

import (
	"context"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// ClearWaitlistTable removes all test data from the waitlist_entries table.
func ClearWaitlistTable(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(ctx, "TRUNCATE TABLE booking.waitlist_entries RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

// GetWaitlistEntryEncxByID retrieves a waitlist entry from the database by ID,
// for tests that need to assert on the persisted (encrypted) row directly.
func GetWaitlistEntryEncxByID(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id uuid.UUID,
) (*domain.WaitlistEntryEncx, error) {
	t.Helper()

	query := `
		SELECT
			id, client_id,
			product_id_encrypted, desired_range_start_encrypted, desired_range_end_encrypted,
			notes_encrypted,
			guest_first_name_encrypted, guest_last_name_encrypted,
			guest_email_encrypted, guest_phone_encrypted,
			status, notified_at, claimed_at, expired_at, cancelled_at,
			created_at, updated_at,
			dek_encrypted, key_version, metadata
		FROM booking.waitlist_entries
		WHERE id = $1
	`

	var entryEncx domain.WaitlistEntryEncx
	err := pool.QueryRow(ctx, query, id).Scan(
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

	return &entryEncx, err
}

// CountWaitlistEntries counts all waitlist entries in the database.
func CountWaitlistEntries(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (int, error) {
	t.Helper()

	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM booking.waitlist_entries").Scan(&count)
	return count, err
}
