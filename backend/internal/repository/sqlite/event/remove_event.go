package eventRepository

import (
	"context"
	"errors"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) RemoveEvent(ctx context.Context, eventID string) error {
	query := "DELETE FROM events WHERE eventid=?"
	result, err := e.DB.ExecContext(ctx, query, eventID)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}

	if rowsAffected == 0 {
		return rp.NewNotDeletedErr(err, fmt.Sprintf("event with ID %s", eventID))
	}

	return nil
}
