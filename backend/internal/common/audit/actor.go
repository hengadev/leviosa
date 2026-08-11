package audit

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/google/uuid"
)

// ActorFromContext extracts the acting user's ID and role from the request
// context's session info (set by the session-auth middleware). It returns
// uuid.Nil and "unknown" if no session info is present in context — callers
// still get an audit entry, just with an unresolved actor, rather than
// silently skipping the write.
func ActorFromContext(ctx context.Context) (uuid.UUID, string) {
	info, ok := session.SessionInfoFromContext(ctx)
	if !ok || info == nil {
		return uuid.Nil, "unknown"
	}
	return info.UserID, info.Role.String()
}
