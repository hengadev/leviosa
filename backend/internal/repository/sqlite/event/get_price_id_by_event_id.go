package eventRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) GetPriceIDByEventID(ctx context.Context, eventID string) (string, error) {
	var priceID string
	err := e.DB.QueryRowContext(ctx, "SELECT priceid from events where id = ?;", eventID).Scan(&priceID)
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
