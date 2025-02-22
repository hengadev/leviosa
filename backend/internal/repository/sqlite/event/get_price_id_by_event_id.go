package eventRepository

import (
	"context"
	"database/sql"
	"errors"

	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (e *EventRepository) GetPriceID(ctx context.Context, eventID string) (string, error) {
	var priceID string
	query := `
        SELECT 
            encrypted_price_id
        FROM events 
        WHERE id = ?;`
	err := e.DB.QueryRowContext(ctx, query, eventID).Scan(&priceID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", rp.NewNotFoundErr(err, "price ID")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return "", rp.NewContextErr(err)
		default:
			return "", rp.NewDatabaseErr(err)
		}
	}
	return priceID, nil
}
