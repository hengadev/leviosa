package eventRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) GetPriceID(ctx context.Context, eventID string) (string, error) {
	var priceID string
	query := `
        SELECT 
            price_id
        FROM events 
        WHERE id = ?;`
	err := e.DB.QueryRowContext(ctx, query, eventID).Scan(&priceID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", rp.NewNotFoundErr(err, fmt.Sprintf("price ID for event with ID %s", eventID))
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return "", rp.NewContextErr(err)
		default:
			return "", rp.NewDatabaseErr(err)
		}
	}
	return priceID, nil
}
