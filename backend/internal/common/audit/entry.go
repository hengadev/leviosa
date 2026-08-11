// Package audit provides a shared, cross-cutting audit-log capability
// following the same shared-infrastructure pattern as errs/ctxutil: a small
// port (Logger) plus a concrete Postgres-backed adapter (PostgresLogger),
// meant to be injected directly into application-layer services that need
// to record sensitive actions (role changes, refunds, admin overrides,
// settings changes, ...).
package audit

import "github.com/google/uuid"

// Entry represents a single audit record: who did what, to which entity,
// and what changed. Entries are immutable once written — there is no
// update/delete path anywhere in this package, by design.
type Entry struct {
	// ActorID is the ID of the authenticated user who performed the action.
	// uuid.Nil if the actor could not be determined from context.
	ActorID uuid.UUID

	// ActorRole is the string form of the actor's role at the time of the
	// action (e.g. "administrator").
	ActorRole string

	// Action identifies the kind of action, e.g. "user.role_updated",
	// "booking.refunded", "booking.admin_override", "settings.field_updated".
	Action string

	// TargetType identifies the kind of entity affected, e.g. "user",
	// "booking", "setting".
	TargetType string

	// TargetID identifies the specific entity affected. Kept as a string
	// since not every target has a UUID identifier (e.g. settings are
	// keyed by a string key).
	TargetID string

	// Payload is a small JSON-serializable map describing what changed.
	Payload map[string]any
}
