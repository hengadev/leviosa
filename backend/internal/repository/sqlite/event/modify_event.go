package eventRepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain/event"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/pkg/sqliteutil"
)

func (e *EventRepository) ModifyEvent(
	ctx context.Context,
	event *eventService.Event,
	whereMap map[string]any,
	prohibitedFields ...string,
) error {
	query, values, errs := sqliteutil.WriteUpdateQuery(*event, whereMap)
	if len(errs) > 0 {
		return rp.NewInternalErr(errs)
	}
	res, err := e.DB.ExecContext(ctx, query, values...)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotUpdatedErr(err, fmt.Sprintf("event with ID %s", event.ID))
	}
	return nil
}
