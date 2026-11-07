package ports

import (
	"context"

	"github.com/google/uuid"
)

// PracticeRoomService gives a practitioner the room their availabilities
// need. A single practice has no use for buildings, rooms and allocations,
// so they stay invisible: each practitioner gets a room of their own in the
// practice building, both created on the spot when needed.
type PracticeRoomService interface {
	// EnsurePracticeRoom makes sure the practitioner (by user ID) holds an
	// active dedicated allocation of a room with opening hours. It changes
	// nothing for a practitioner who already holds one, so it is safe to
	// call again.
	EnsurePracticeRoom(ctx context.Context, userID uuid.UUID) error
}
