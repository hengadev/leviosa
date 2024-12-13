package middleware

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

func ValidateRoleInContext(ctx context.Context, expectedRole userService.Role) error {
	role, ok := ctx.Value(RoleKey).(userService.Role)
	if !ok {
		return fmt.Errorf("extract role from context")
	} else if role != expectedRole {
		return fmt.Errorf("expected role %q, got %q", expectedRole, role)
	}
	return nil
}
