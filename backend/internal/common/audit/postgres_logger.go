package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresLogger is the Postgres-backed adapter for the Logger port. It
// writes to the append-only audit.entries table (see the
// {timestamp}_audit_init_schema.sql migration) and exposes no update or
// delete operation.
type PostgresLogger struct {
	pool   *pgxpool.Pool
	schema string
}

// NewPostgresLogger creates a Postgres-backed audit Logger.
func NewPostgresLogger(pool *pgxpool.Pool) *PostgresLogger {
	return &PostgresLogger{pool: pool, schema: "audit"}
}

// Write inserts a new audit entry. It is the only write path exposed by
// this adapter — there is no corresponding Update or Delete.
func (l *PostgresLogger) Write(ctx context.Context, entry Entry) error {
	payload, err := json.Marshal(entry.Payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.entries (actor_id, actor_role, action, target_type, target_id, payload)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, l.schema)

	if _, err := l.pool.Exec(ctx, query,
		entry.ActorID,
		entry.ActorRole,
		entry.Action,
		entry.TargetType,
		entry.TargetID,
		payload,
	); err != nil {
		return errs.ClassifyPgError("write audit entry", err)
	}

	return nil
}
