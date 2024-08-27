package userRepository

import (
	"context"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) DeleteUser(ctx context.Context, userID int) error {
	res, err := u.DB.ExecContext(ctx, "DELETE FROM users WHERE id = ?;", userID)
	if err != nil {
		return rp.NewRessourceDeleteErr(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return rp.NewRessourceDeleteErr(err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
