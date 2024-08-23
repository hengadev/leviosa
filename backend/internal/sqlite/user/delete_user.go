package userRepository

import (
	"context"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) DeleteUser(ctx context.Context, userID int) (int, error) {
	res, err := u.DB.ExecContext(ctx, "DELETE FROM users WHERE id = ?;", userID)
	if err != nil {
		return 0, rp.NewRessourceDeleteErr(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, rp.NewRessourceDeleteErr(err)
	}
	return int(rowsAffected), nil
}
