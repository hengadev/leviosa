package userRepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"
)

func (u *Repository) ModifyAccount(
	ctx context.Context,
	user *userService.User,
	whereMap map[string]any,
	prohibitedFields ...string,
) error {
	query, values, err := sqliteutil.WriteUpdateQuery(*user, whereMap, prohibitedFields...)
	if err != nil {
		return rp.NewInternalError(err)
	}
	result, err := u.DB.ExecContext(ctx, query, values...)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextError(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotUpdatedErr(err, fmt.Sprintf("user with ID %s", user.ID))
	}
	return nil
}
