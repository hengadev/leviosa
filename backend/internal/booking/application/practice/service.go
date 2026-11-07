package practice

import (
	"context"
	"fmt"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
)

// The practice's room, created the first time a practitioner is added. The
// building's address is a placeholder: the practice address shown to clients
// comes from the address setting (emails fall back to the building only when
// that setting is empty).
const (
	practiceName       = "Cabinet"
	placeholderAddress = "Adresse à préciser"
	placeholderField   = "—"
	practiceCountry    = "France"
)

// Opening hours of the practice room, every day of the week. Availabilities
// must fit inside them, so they are wide: the practitioner's own
// availabilities are what limits the slots.
var (
	openTime  = time.Date(0, 1, 1, 7, 0, 0, 0, time.UTC)
	closeTime = time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC)
)

type Service struct {
	buildings   ports.BuildingService
	rooms       ports.RoomService
	schedules   ports.RoomScheduleRepository
	allocations ports.RoomAllocationService
	now         func() time.Time
}

func New(buildings ports.BuildingService, rooms ports.RoomService, schedules ports.RoomScheduleRepository, allocations ports.RoomAllocationService) ports.PracticeRoomService {
	return &Service{
		buildings:   buildings,
		rooms:       rooms,
		schedules:   schedules,
		allocations: allocations,
		now:         time.Now,
	}
}

func (s *Service) EnsurePracticeRoom(ctx context.Context, userID uuid.UUID) error {
	allocations, err := s.allocations.GetPartnerAllocations(ctx, &domain.GetPartnerAllocationsRequest{UserID: userID, ActiveOnly: true})
	if err != nil {
		return fmt.Errorf("get practitioner allocations: %w", err)
	}
	for _, allocation := range allocations {
		if allocation.IsActive {
			return nil
		}
	}

	roomID, err := s.practiceRoom(ctx)
	if err != nil {
		return err
	}
	if err := s.ensureOpeningHours(ctx, roomID); err != nil {
		return err
	}

	start := s.now().UTC()
	if _, err := s.allocations.CreateDedicatedAllocation(ctx, &domain.CreateDedicatedAllocationRequest{
		RoomID:    roomID,
		UserID:    userID,
		StartDate: &start,
	}); err != nil {
		return fmt.Errorf("allocate the practice room: %w", err)
	}
	return nil
}

// practiceRoom returns an active room no practitioner holds yet: a
// dedicated allocation gives a room to one practitioner only. When every
// room is taken, it adds one to the practice building, creating the
// building first on an empty practice.
func (s *Service) practiceRoom(ctx context.Context) (uuid.UUID, error) {
	active := true
	rooms, err := s.rooms.ListRooms(ctx, ports.RoomFilter{IsActive: &active})
	if err != nil {
		return uuid.Nil, fmt.Errorf("list rooms: %w", err)
	}
	for _, room := range rooms {
		held, err := s.allocations.GetRoomAllocations(ctx, &domain.GetRoomAllocationsRequest{RoomID: room.ID, ActiveOnly: true})
		if err != nil {
			return uuid.Nil, fmt.Errorf("get room allocations: %w", err)
		}
		if len(held) == 0 {
			return room.ID, nil
		}
	}

	var buildingID uuid.UUID
	if len(rooms) > 0 {
		buildingID = rooms[0].BuildingID
	} else {
		building, err := s.buildings.CreateBuilding(ctx, &domain.CreateBuildingRequest{
			Name:       practiceName,
			Address:    placeholderAddress,
			City:       placeholderField,
			PostalCode: placeholderField,
			Country:    practiceCountry,
			IsActive:   true,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("create the practice building: %w", err)
		}
		buildingID = building.ID
	}

	name := practiceName
	if len(rooms) > 0 {
		name = fmt.Sprintf("%s %d", practiceName, len(rooms)+1)
	}
	room, err := s.rooms.CreateRoom(ctx, &domain.CreateRoomRequest{
		BuildingID: buildingID,
		Name:       name,
		Capacity:   1,
		IsActive:   true,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create the practice room: %w", err)
	}
	return room.ID, nil
}

// ensureOpeningHours gives the room opening hours on every day of the week
// when it has none.
func (s *Service) ensureOpeningHours(ctx context.Context, roomID uuid.UUID) error {
	existing, err := s.schedules.ListByRoom(ctx, roomID)
	if err != nil {
		return fmt.Errorf("list room opening hours: %w", err)
	}
	if len(existing) > 0 {
		return nil
	}
	for day := 0; day < 7; day++ {
		dayOfWeek := day
		if err := s.schedules.Create(ctx, &domain.RoomAvailabilitySchedule{
			RoomID:    roomID,
			DayOfWeek: &dayOfWeek,
			OpenTime:  openTime,
			CloseTime: closeTime,
			IsActive:  true,
		}); err != nil {
			return fmt.Errorf("create room opening hours: %w", err)
		}
	}
	return nil
}
