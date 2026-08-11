package helpers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// AuditEntryRow is a test-only read model for a row in audit.entries. The
// audit.Logger port intentionally exposes no query method (append-only,
// write-only by design — see internal/common/audit), so integration tests
// verify what was written by querying the table directly instead of going
// through the port.
type AuditEntryRow struct {
	ActorID    uuid.UUID
	ActorRole  string
	Action     string
	TargetType string
	TargetID   string
	Payload    map[string]any
}

// ClearAuditEntriesTable truncates audit.entries between tests.
func ClearAuditEntriesTable(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(ctx, "TRUNCATE TABLE audit.entries RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}

// GetLatestAuditEntry returns the most recent audit.entries row matching the
// given target type/ID, or ok=false if none exists.
func GetLatestAuditEntry(t *testing.T, ctx context.Context, pool *pgxpool.Pool, targetType, targetID string) (AuditEntryRow, bool) {
	t.Helper()

	var (
		row     AuditEntryRow
		payload []byte
	)

	err := pool.QueryRow(ctx, `
		SELECT actor_id, actor_role, action, target_type, target_id, payload
		FROM audit.entries
		WHERE target_type = $1 AND target_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, targetType, targetID).Scan(
		&row.ActorID,
		&row.ActorRole,
		&row.Action,
		&row.TargetType,
		&row.TargetID,
		&payload,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return AuditEntryRow{}, false
		}
		require.NoError(t, err)
		return AuditEntryRow{}, false
	}

	if len(payload) > 0 {
		require.NoError(t, json.Unmarshal(payload, &row.Payload))
	}

	return row, true
}

// CountAuditEntries returns how many audit.entries rows match the given
// target type/ID (useful for asserting exactly-once writes).
func CountAuditEntries(t *testing.T, ctx context.Context, pool *pgxpool.Pool, targetType, targetID string) int {
	t.Helper()

	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit.entries WHERE target_type = $1 AND target_id = $2
	`, targetType, targetID).Scan(&count)
	require.NoError(t, err)
	return count
}
