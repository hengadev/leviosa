package registerService

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) CreateRegistration(ctx context.Context, userID, spotStr string, event *eventService.Event) error {
	day, month, year := parseEventTimeForRegistration(event)
	// NOTE: why is the spot str a string and not an int ?
	spot, err := strconv.Atoi(spotStr)
	if err != nil {
		return fmt.Errorf("convert spot from string to int: %w", err)
	}
	offsetDuration := time.Duration(int(event.SessionDuration) * (spot - 1))
	registrationBeginAt := event.BeginAt.Add(offsetDuration)
	_ = registrationBeginAt
	registration := NewRegistration(
		userID,
		"",
		"event",
		time.Now(),
		time.Now().Add(24*time.Hour),
		nil,
		nil,
	)
	err = s.Repo.AddRegistration(ctx, registration, day, year, month)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrContext):
			return err
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrNotCreated):
			return domain.NewNotCreatedErr(err)
		default:
			return fmt.Errorf("add registration: %w", err)
		}
	}
	return nil
}

func parseEventTimeForRegistration(e *eventService.Event) (int, string, int) {
	month := strings.ToLower(time.Month(e.Month).String())
	return e.Day, month, e.Year
}
