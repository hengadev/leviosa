package settings

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"
)

// auditFieldUpdate records an audit entry for a successful settings
// field-update use case. field is the settings key (e.g.
// settings.CompanyName); newValue is recorded as-is in the audit payload.
// A failed audit write is logged (see audit.WriteOrLog) but never fails the
// settings update itself.
func (s *SettingsService) auditFieldUpdate(ctx context.Context, field string, newValue any) {
	actorID, actorRole := audit.ActorFromContext(ctx)
	audit.WriteOrLog(ctx, s.audit, audit.Entry{
		ActorID:    actorID,
		ActorRole:  actorRole,
		Action:     "settings.field_updated",
		TargetType: "setting",
		TargetID:   field,
		Payload: map[string]any{
			"field":     field,
			"new_value": newValue,
		},
	})
}
