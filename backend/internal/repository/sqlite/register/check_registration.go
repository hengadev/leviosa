package registerRepository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// Function that return if there is a registration for a certain user for a certain event at a certain time.
func (r *repository) CheckRegistration(ctx context.Context, registration *register.Registration) error {
	var value int
	query := "SELECT 1 FROM ? WHERE beginAt=?;"
	err := r.DB.QueryRowContext(ctx, query, registration.EventID, registration.BeginAt.Format(time.RFC3339)).Scan(&value)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return rp.NewNotFoundError(err, "user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextError(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	return nil
}
