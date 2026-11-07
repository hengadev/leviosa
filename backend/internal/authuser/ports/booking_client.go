package ports

import (
	"context"

	"github.com/google/uuid"
)

// BookingClient defines the interface for communicating with the booking service.
// Used by authuser to claim guest bookings after account creation.
type BookingClient interface {
	// ClaimBookings links all guest bookings matching the given email to the
	// specified client. Fire-and-forget semantics: errors are logged but do not
	// propagate to the caller.
	ClaimBookings(ctx context.Context, clientID string, email string)

	// EnsurePracticeRoom gives a practitioner (by user ID) the practice room
	// their availabilities need, creating it when the practice has none.
	EnsurePracticeRoom(ctx context.Context, userID uuid.UUID) error
}
