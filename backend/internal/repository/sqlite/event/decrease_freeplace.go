package eventRepository

import (
	"context"
	"errors"
	"fmt"

	rp "github.com/GaryHY/leviosa/internal/repository"
)

// DecreaseFreePlace decrements the freeplace count for an event with the given eventID.
func (e *EventRepository) DecreaseFreePlace(ctx context.Context, eventID string) error {
	query := `
        UPDATE events
        SET freeplace = freeplace - 1 
        WHERE id = ? AND freeplace > 0;`
	res, err := e.DB.ExecContext(ctx, query, eventID)
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
