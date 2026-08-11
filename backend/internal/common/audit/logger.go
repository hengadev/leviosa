package audit

import "context"

// Logger is the audit-log port. It intentionally exposes only Write — an
// audit log that can be edited or removed after the fact isn't trustworthy
// as one, so no Update/Delete operation is exposed here by design.
type Logger interface {
	// Write appends a new audit entry. Implementations should treat this as
	// an insert-only operation.
	Write(ctx context.Context, entry Entry) error
}
