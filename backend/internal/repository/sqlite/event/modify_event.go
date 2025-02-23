package eventRepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain/event/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/pkg/sqliteutil"
)

func (e *EventRepository) ModifyEvent(
	ctx context.Context,
	event *models.Event,
	whereMap map[string]any,
) error {
	if event == nil {
		return rp.NewValidationErr(errors.New("nil event"), "event")
	}
	query, values, errs := sqliteutil.WriteUpdateQuery(*event, whereMap)
	if len(errs) > 0 {
		return rp.NewInternalErr(errs)
	}
	fmt.Println("the query is:", query)
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
