package eventService_test

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain/event/models"
)

type MockRepo struct {
	GetEventByIDFunc            func(ctx context.Context, eventID string) (*models.Event, error)
	GetEventForUserFunc         func(ctx context.Context, userID string) (*models.EventUser, error)
	GetPriceIDFunc              func(ctx context.Context, eventID string) (string, error)
	EventHasAvailablePlacesFunc func(ctx context.Context, eventID string) (bool, error)
	IsDateAvailableFunc         func(ctx context.Context, day, month, year int) error
	AddEventFunc                func(ctx context.Context, event *models.Event) (string, error)
	RemoveEventFunc             func(ctx context.Context, eventID string) error
	ModifyEventFunc             func(ctx context.Context, event *models.Event, whereMap map[string]any) error
	DecreaseFreePlaceFunc       func(ctx context.Context, eventID string) error
}

func (m *MockRepo) GetEventByID(ctx context.Context, eventID string) (*models.Event, error) {
	if m.GetEventByIDFunc != nil {
		return m.GetEventByIDFunc(ctx, eventID)
	}
	return nil, nil
}

func (m *MockRepo) GetEventForUser(ctx context.Context, eventID string) (*models.EventUser, error) {
	if m.GetEventByIDFunc != nil {
		return m.GetEventForUserFunc(ctx, eventID)
	}
	return nil, nil
}

func (m *MockRepo) GetPriceID(ctx context.Context, eventID string) (string, error) {
	if m.GetPriceIDFunc != nil {
		return m.GetPriceIDFunc(ctx, eventID)
	}
	return "", nil
}

func (m *MockRepo) EventHasAvailablePlaces(ctx context.Context, eventID string) (bool, error) {
	if m.EventHasAvailablePlacesFunc != nil {
		return m.EventHasAvailablePlacesFunc(ctx, eventID)
	}
	return false, nil
}
func (m *MockRepo) IsDateAvailable(ctx context.Context, day, month, year int) error {
	if m.IsDateAvailableFunc != nil {
		return m.IsDateAvailableFunc(ctx, day, month, year)
	}
	return nil
}

func (m *MockRepo) AddEvent(ctx context.Context, event *models.Event) (string, error) {
	if m.AddEventFunc != nil {
		return m.AddEventFunc(ctx, event)
	}
	return "", nil
}

func (m *MockRepo) RemoveEvent(ctx context.Context, eventID string) error {
	if m.RemoveEventFunc != nil {
		return m.RemoveEventFunc(ctx, eventID)
	}
	return nil
}

func (m *MockRepo) ModifyEvent(ctx context.Context, event *models.Event, whereMap map[string]any) error {
	if m.ModifyEventFunc != nil {
		return m.ModifyEventFunc(ctx, event, whereMap)
	}
	return nil
}

func (m *MockRepo) DecreaseFreePlace(ctx context.Context, eventID string) error {
	if m.DecreaseFreePlaceFunc != nil {
		return m.DecreaseFreePlaceFunc(ctx, eventID)
	}
	return nil
}
