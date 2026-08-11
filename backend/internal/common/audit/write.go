package audit

import (
	"context"
	"log/slog"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
)

// WriteOrLog writes an audit entry through w and, if the write fails, makes
// sure the failure is surfaced via logging rather than silently swallowed.
// It never returns an error: a failed audit write must not fail the
// business action it's attached to (audit is a reliability-critical side
// effect, not a precondition for the action to succeed), but it also must
// not pass unnoticed, hence the error-level log.
//
// w may be nil (e.g. in tests that don't wire an audit logger), in which
// case this is a no-op.
func WriteOrLog(ctx context.Context, w Logger, entry Entry) {
	if w == nil {
		return
	}

	if err := w.Write(ctx, entry); err != nil {
		logger, lerr := ctxutil.GetLoggerFromContext(ctx)
		if lerr != nil {
			logger = slog.Default()
		}
		logger.ErrorContext(ctx, "audit: failed to write audit entry",
			"action", entry.Action,
			"target_type", entry.TargetType,
			"target_id", entry.TargetID,
			"actor_id", entry.ActorID,
			"error", err,
		)
	}
}
