package eventRepository

import (
	"context"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) RemoveEvent(ctx context.Context, ID string) error {
	fail := func(err error) error {
		return rp.NewRessourceDeleteErr(err)
	}
	res, err := e.DB.ExecContext(ctx, "DELETE from events where id=?;", ID)
	if err != nil {
		return fail(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fail(err)
	}
	if rowsAffected == 0 {
		return fail(fmt.Errorf("no row deleted"))
	}
	return nil
}
