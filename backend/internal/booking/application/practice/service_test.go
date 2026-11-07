package practice

import (
	"context"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeBuildings struct {
	ports.BuildingService
	created []*domain.CreateBuildingRequest
}

func (f *fakeBuildings) CreateBuilding(_ context.Context, req *domain.CreateBuildingRequest) (*domain.BuildingResponse, error) {
	f.created = append(f.created, req)
	return &domain.BuildingResponse{ID: uuid.New(), Name: req.Name}, nil
}

type fakeRooms struct {
	ports.RoomService
	rooms   []*domain.RoomResponse
	created []*domain.CreateRoomRequest
}

func (f *fakeRooms) ListRooms(context.Context, ports.RoomFilter) ([]*domain.RoomResponse, error) {
	return f.rooms, nil
}

func (f *fakeRooms) CreateRoom(_ context.Context, req *domain.CreateRoomRequest) (*domain.RoomResponse, error) {
	f.created = append(f.created, req)
	room := &domain.RoomResponse{ID: uuid.New(), BuildingID: req.BuildingID, Name: req.Name}
	f.rooms = append(f.rooms, room)
	return room, nil
}

type fakeSchedules struct {
	ports.RoomScheduleRepository
	byRoom map[uuid.UUID][]*domain.RoomAvailabilitySchedule
}

func (f *fakeSchedules) ListByRoom(_ context.Context, roomID uuid.UUID) ([]*domain.RoomAvailabilitySchedule, error) {
	return f.byRoom[roomID], nil
}

func (f *fakeSchedules) Create(_ context.Context, s *domain.RoomAvailabilitySchedule) error {
	f.byRoom[s.RoomID] = append(f.byRoom[s.RoomID], s)
	return nil
}

type fakeAllocations struct {
	ports.RoomAllocationService
	existing []*domain.RoomAllocation
	byRoom   map[uuid.UUID][]*domain.RoomAllocation
	created  []*domain.CreateDedicatedAllocationRequest
}

func (f *fakeAllocations) GetRoomAllocations(_ context.Context, req *domain.GetRoomAllocationsRequest) ([]*domain.RoomAllocation, error) {
	return f.byRoom[req.RoomID], nil
}

func (f *fakeAllocations) GetPartnerAllocations(context.Context, *domain.GetPartnerAllocationsRequest) ([]*domain.RoomAllocation, error) {
	return f.existing, nil
}

func (f *fakeAllocations) CreateDedicatedAllocation(_ context.Context, req *domain.CreateDedicatedAllocationRequest) (*domain.RoomAllocation, error) {
	f.created = append(f.created, req)
	return &domain.RoomAllocation{RoomID: req.RoomID, IsActive: true}, nil
}

func newFakes() (*fakeBuildings, *fakeRooms, *fakeSchedules, *fakeAllocations) {
	return &fakeBuildings{}, &fakeRooms{}, &fakeSchedules{byRoom: map[uuid.UUID][]*domain.RoomAvailabilitySchedule{}}, &fakeAllocations{byRoom: map[uuid.UUID][]*domain.RoomAllocation{}}
}

func TestEnsurePracticeRoom_CreatesTheRoomOnAnEmptyPractice(t *testing.T) {
	buildings, rooms, schedules, allocations := newFakes()
	svc := New(buildings, rooms, schedules, allocations)
	userID := uuid.New()

	require.NoError(t, svc.EnsurePracticeRoom(context.Background(), userID))

	require.Len(t, buildings.created, 1)
	require.Len(t, rooms.created, 1)
	roomID := rooms.rooms[0].ID
	assert.Len(t, schedules.byRoom[roomID], 7, "open every day of the week")
	require.Len(t, allocations.created, 1)
	assert.Equal(t, roomID, allocations.created[0].RoomID)
	assert.Equal(t, userID, allocations.created[0].UserID)
	assert.NotNil(t, allocations.created[0].StartDate)
}

func TestEnsurePracticeRoom_ReusesTheExistingRoom(t *testing.T) {
	buildings, rooms, schedules, allocations := newFakes()
	existing := &domain.RoomResponse{ID: uuid.New()}
	rooms.rooms = []*domain.RoomResponse{existing}
	schedules.byRoom[existing.ID] = []*domain.RoomAvailabilitySchedule{{RoomID: existing.ID}}
	svc := New(buildings, rooms, schedules, allocations)

	require.NoError(t, svc.EnsurePracticeRoom(context.Background(), uuid.New()))

	assert.Empty(t, buildings.created)
	assert.Empty(t, rooms.created)
	assert.Len(t, schedules.byRoom[existing.ID], 1, "existing opening hours are left alone")
	require.Len(t, allocations.created, 1)
	assert.Equal(t, existing.ID, allocations.created[0].RoomID)
}

func TestEnsurePracticeRoom_AddsARoomWhenEveryRoomIsHeld(t *testing.T) {
	buildings, rooms, schedules, allocations := newFakes()
	held := &domain.RoomResponse{ID: uuid.New(), BuildingID: uuid.New()}
	rooms.rooms = []*domain.RoomResponse{held}
	allocations.byRoom[held.ID] = []*domain.RoomAllocation{{RoomID: held.ID, IsActive: true}}
	svc := New(buildings, rooms, schedules, allocations)

	require.NoError(t, svc.EnsurePracticeRoom(context.Background(), uuid.New()))

	assert.Empty(t, buildings.created, "the room goes in the existing building")
	require.Len(t, rooms.created, 1)
	assert.Equal(t, held.BuildingID, rooms.created[0].BuildingID)
	newRoom := rooms.rooms[1].ID
	assert.Len(t, schedules.byRoom[newRoom], 7)
	require.Len(t, allocations.created, 1)
	assert.Equal(t, newRoom, allocations.created[0].RoomID)
}

func TestEnsurePracticeRoom_ChangesNothingWhenAlreadyAllocated(t *testing.T) {
	buildings, rooms, schedules, allocations := newFakes()
	allocations.existing = []*domain.RoomAllocation{{RoomID: uuid.New(), IsActive: true}}
	svc := New(buildings, rooms, schedules, allocations)

	require.NoError(t, svc.EnsurePracticeRoom(context.Background(), uuid.New()))

	assert.Empty(t, buildings.created)
	assert.Empty(t, rooms.created)
	assert.Empty(t, allocations.created)
}
