package eventService

import (
	"errors"

	"github.com/hengadev/leviosa/internal/domain/event/models"
)

func ParseBeginAt(event *models.Event) (int, int, int, error) {
	if event.BeginAt.IsZero() {
		return 0, 0, 0, errors.New("BeginAt is zero")
	}
	day := event.BeginAt.Day()
	month := int(event.BeginAt.Month())
	year := event.BeginAt.Year()
	return day, month, year, nil
}
