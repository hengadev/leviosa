package register

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
)

func (s *Service) CreateRegistration(ctx context.Context, userID, spotStr string, event *event.Event) error {
	day, month, year := parseEventTimeForRegistration(event)
	// check if there is a registration for the user for that specific event
	ok, err := s.Repo.HasRegistration(ctx, day, year, month, userID)
	if err != nil {
		return fmt.Errorf("check registration: %w", err)
	}
	// there is a registration, remove the previous one and continue
	if ok {
		if err := s.Repo.RemoveRegistration(ctx, day, year, month); err != nil {
			return fmt.Errorf("remove previous registration: %w", err)
		}
	}
	spot, err := strconv.Atoi(spotStr)
	if err != nil {
		return fmt.Errorf("convert spot from string to int: %w", err)
	}
	offsetDuration := time.Duration(int(event.SessionDuration) * (spot - 1))
	registrationBeginAt := event.BeginAt.Add(offsetDuration)
	registration := NewRegistration(userID, event.ID, registrationBeginAt)
	if err = s.Repo.AddRegistration(ctx, registration, day, year, month); err != nil {
		return fmt.Errorf("add registration: %w", err)
	}
	return nil
}

func parseEventTimeForRegistration(e *event.Event) (int, string, int) {
	month := strings.ToLower(time.Month(e.Month).String())
	return e.Day, month, e.Year
}
