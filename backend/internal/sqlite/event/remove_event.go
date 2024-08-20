package eventRepository

import (
	"context"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) RemoveEvent(ctx context.Context, ID string) (string, error) {
	fail := func(err error) (string, error) {
		return "", rp.NewRessourceDeleteErr(err)
	}
	res, err := e.DB.ExecContext(ctx, "DELETE from events where id=?;", ID)
	if err != nil {
		return fail(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fail(err)
	}
	fmt.Println("the rows aff is:", rowsAffected)
	if rowsAffected == 0 {
		return fail(fmt.Errorf("no row deleted"))
	}
	return ID, nil
}
