package registerRepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (r *repository) AddRegistration(ctx context.Context, reg *register.Registration, day, year int, month string) error {
	tablename := getTablename(day, year, month)
	query := fmt.Sprintf("INSERT INTO %s (userid, eventid, beginat) values (?,?,?);", tablename)
	result, err := r.DB.ExecContext(ctx, query, reg.UserID, reg.EventID, reg.BeginAt)
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
		return rp.NewNotCreatedErr(errors.New("no rows affected by insertion statement"), "registration")
	}
	return nil
}
