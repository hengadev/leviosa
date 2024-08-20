package eventRepository

import (
	"context"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) DecreaseFreeplace(ctx context.Context, ID string) (int, error) {
	res, err := e.DB.ExecContext(ctx, "UPDATE events SET freeplace = freeplace - 1 WHERE id=?;", ID)
	if err != nil {
		return 0, rp.NewRessourceUpdateErr(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, rp.NewRessourceUpdateErr(err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("rowsAffected = 0, ID not found")
	}
	return int(rowsAffected), nil
}
