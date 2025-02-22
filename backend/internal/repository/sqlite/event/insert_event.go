package eventRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/event/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (e *EventRepository) InsertEvent(
	ctx context.Context,
	tx *sql.Tx,
	event *models.Event,
) error {
	query := `INSERT INTO events (
                id,
                encrypted_title,
                encrypted_description,
				encrypted_postal_code,
				encrypted_city, 
				encrypted_address1, 
				encrypted_address2,
                placecount,
                freeplace,
                encrypted_begin_at,
                encrypted_end_at,
                encrypted_price_id,
                day,
                month,
                year
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.ExecContext(ctx, query,
		event.ID,
		event.Title,
		event.Description,
		event.PostalCode,
		event.City,
		event.Address1,
		event.Address2,
		event.PlaceCount,
		event.FreePlace,
		event.EncryptedBeginAt,
		event.EncryptedEndAt,
		event.PriceID,
		event.Day,
		event.Month,
		event.Year,
	)
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
		return rp.NewNotCreatedErr(errors.New("no rows affected by insertion statement"), "event")
	}
	return nil
}
