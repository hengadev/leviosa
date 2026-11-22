package ports

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/google/uuid"
)

// PracticeAccessProvider supplies the practice access (address, door code,
// access instructions) of a booking's room. In the modular monolith it reads
// the settings domain in-process on every call, so a door code changed in the
// admin settings shows on the next read. The address is the company address
// setting, falling back to the address of the room's building. The values are
// private: never log them.
type PracticeAccessProvider interface {
	GetPracticeAccess(ctx context.Context, roomID uuid.UUID) (domain.PracticeAccess, error)
}
