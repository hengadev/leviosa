package settings

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	bookingPorts "github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	settingsPorts "github.com/Leviosa-care/leviosa/backend/internal/settings/ports"
	"github.com/google/uuid"
)

// InProcessPracticeAccessProvider reads the practice access from the settings
// service on every call, with no cache: a door code changed by an
// administrator shows on the next read. It does not go through the
// notification module's cached settings provider.
type InProcessPracticeAccessProvider struct {
	settings  settingsPorts.SettingsService
	rooms     bookingPorts.RoomService
	buildings bookingPorts.BuildingService
}

// NewInProcessPracticeAccessProvider creates the provider. rooms and buildings
// resolve the building address used when the company address setting is
// empty; with either nil there is no fallback.
func NewInProcessPracticeAccessProvider(settings settingsPorts.SettingsService, rooms bookingPorts.RoomService, buildings bookingPorts.BuildingService) bookingPorts.PracticeAccessProvider {
	return &InProcessPracticeAccessProvider{settings: settings, rooms: rooms, buildings: buildings}
}

// GetPracticeAccess fails only when the door code or the instructions cannot
// be read; an unreadable address leaves it empty. Errors and log lines never
// carry a value.
func (p *InProcessPracticeAccessProvider) GetPracticeAccess(ctx context.Context, roomID uuid.UUID) (domain.PracticeAccess, error) {
	access, err := p.settings.GetPracticeAccess(ctx)
	if err != nil {
		return domain.PracticeAccess{}, fmt.Errorf("get practice access from settings: %w", err)
	}
	return domain.PracticeAccess{
		Address:      p.address(ctx, roomID),
		DoorCode:     access.DoorCode,
		Instructions: access.AccessInstructions,
	}, nil
}

// address is the company address setting, or the address of the room's
// building when the setting is empty, as the booking emails do.
func (p *InProcessPracticeAccessProvider) address(ctx context.Context, roomID uuid.UUID) string {
	company, err := p.settings.GetCompanyLegalAddress(ctx)
	if err != nil {
		if !errors.Is(err, errs.ErrRepositoryNotFound) {
			slog.WarnContext(ctx, "failed to read the company address for the practice access", "error", err)
		}
	} else if address := strings.TrimSpace(company.Address); address != "" {
		return address
	}

	if p.rooms == nil || p.buildings == nil || roomID == uuid.Nil {
		return ""
	}
	room, err := p.rooms.GetRoom(ctx, roomID)
	if err != nil {
		slog.WarnContext(ctx, "failed to read the room for the practice access", "room_id", roomID, "error", err)
		return ""
	}
	building, err := p.buildings.GetBuildingByID(ctx, room.BuildingID)
	if err != nil {
		slog.WarnContext(ctx, "failed to read the building for the practice access", "building_id", room.BuildingID, "error", err)
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%s, %s %s", building.Address, building.PostalCode, building.City))
}
