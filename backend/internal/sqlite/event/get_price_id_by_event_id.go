package eventRepository

import (
	"context"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) GetPriceIDByEventID(ctx context.Context, ID string) (string, error) {
	var priceID string
	err := e.DB.QueryRowContext(ctx, "SELECT priceid from events where id = ?;", ID).Scan(&priceID)
	if err != nil {
		return "", rp.NewBadQueryErr(err)
	}
	return priceID, nil
}
