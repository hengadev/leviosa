package userRepository

import (
	"context"
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
	fail := func(err error) error {
		return rp.NewRessourceUpdateErr(err)
	}
	if user == nil {
		return fail(fmt.Errorf("nil user"))
	}
	query, values, err := sqliteutil.WriteUpdateQuery(*user, whereMap, prohibitedFields...)
	if err != nil {
		return fail(err)
	}
	res, err := u.DB.ExecContext(ctx, query, values...)
	if err != nil {
		return fail(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fail(err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user to modified not found")
	}
	return nil
}
