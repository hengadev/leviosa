package eventRepository

import (
	"context"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) DecreaseFreeplace(ctx context.Context, ID string) error {
	fail := func(err error) error {
		return rp.NewRessourceUpdateErr(err)
	}
	res, err := e.DB.ExecContext(ctx, "UPDATE events SET freeplace = freeplace - 1 WHERE id=?;", ID)
	if err != nil {
		return fail(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fail(err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}
	return nil
}
