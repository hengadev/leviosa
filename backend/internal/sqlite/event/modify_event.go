package eventRepository

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"
)

func (e *EventRepository) ModifyEvent(
	ctx context.Context,
	event *eventService.Event,
	whereMap map[string]any,
	prohibitedFields ...string,
) error {
	fail := func(err error) error {
		return rp.NewRessourceUpdateErr(err)
	}
	if event == nil {
		return fail(fmt.Errorf("nil user"))
	}
	query, values, err := sqliteutil.WriteUpdateQuery(*event, whereMap, prohibitedFields...)
	if err != nil {
		return fail(err)
	}
	res, err := e.DB.ExecContext(ctx, query, values...)
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
