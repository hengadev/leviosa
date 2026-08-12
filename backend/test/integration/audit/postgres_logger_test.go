package audit_test

import (
	"context"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"
	th "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestPostgresLogger_Write TEST_PATH=test/integration/audit/postgres_logger_test.go

// TestPostgresLogger_Write confirms the core acceptance criterion of issue
// 019: the Postgres adapter appends a correctly-populated audit entry.
func TestPostgresLogger_Write(t *testing.T) {
	ctx := context.Background()
	th.ClearAuditEntriesTable(t, ctx, testPool)

	logger := audit.NewPostgresLogger(testPool)

	actorID := uuid.New()
	targetID := uuid.New().String()

	err := logger.Write(ctx, audit.Entry{
		ActorID:    actorID,
		ActorRole:  "administrator",
		Action:     "user.role_updated",
		TargetType: "user",
		TargetID:   targetID,
		Payload: map[string]any{
			"previous_role": "standard",
			"new_role":      "premium",
		},
	})
	require.NoError(t, err)

	entry, ok := th.GetLatestAuditEntry(t, ctx, testPool, "user", targetID)
	require.True(t, ok, "expected the written entry to be readable back")

	assert.Equal(t, actorID, entry.ActorID)
	assert.Equal(t, "administrator", entry.ActorRole)
	assert.Equal(t, "user.role_updated", entry.Action)
	assert.Equal(t, "user", entry.TargetType)
	assert.Equal(t, targetID, entry.TargetID)
	assert.Equal(t, "standard", entry.Payload["previous_role"])
	assert.Equal(t, "premium", entry.Payload["new_role"])
}

// TestPostgresLogger_Write_MultipleEntriesAreAppended confirms writes are
// purely additive (append-only) — writing two entries for the same target
// results in two rows, not an overwrite, since there is no update path.
func TestPostgresLogger_Write_MultipleEntriesAreAppended(t *testing.T) {
	ctx := context.Background()
	th.ClearAuditEntriesTable(t, ctx, testPool)

	logger := audit.NewPostgresLogger(testPool)
	targetID := uuid.New().String()

	for i := 0; i < 3; i++ {
		err := logger.Write(ctx, audit.Entry{
			ActorID:    uuid.New(),
			ActorRole:  "administrator",
			Action:     "settings.field_updated",
			TargetType: "setting",
			TargetID:   targetID,
			Payload:    map[string]any{"attempt": i},
		})
		require.NoError(t, err)
	}

	count := th.CountAuditEntries(t, ctx, testPool, "setting", targetID)
	assert.Equal(t, 3, count, "audit.entries has no update path: repeated writes append, they don't overwrite")
}
