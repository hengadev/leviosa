package eventRepository

import (
	"context"
	"errors"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) DecreaseFreeplace(ctx context.Context, eventID string) error {
	res, err := e.DB.ExecContext(ctx, "UPDATE events SET freeplace = freeplace - 1 WHERE id=?;", eventID)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotUpdatedErr(err, fmt.Sprintf("decrease freeplace count for event with ID %s", eventID))
	}
	return nil
}
