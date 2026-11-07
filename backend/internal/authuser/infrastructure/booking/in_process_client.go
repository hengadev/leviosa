package booking

import (
	"context"
	"log/slog"

	authuserPorts "github.com/Leviosa-care/leviosa/backend/internal/authuser/ports"
	bookingPorts "github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
)

// InProcessClient is an in-process implementation of BookingClient that directly
// delegates to the booking services.
//
// This is used in the modular monolith architecture for efficient in-process
// communication without HTTP overhead.
type InProcessClient struct {
	bookingService      bookingPorts.BookingService
	practiceRoomService bookingPorts.PracticeRoomService
}

// NewInProcessClient creates a new in-process BookingClient.
func NewInProcessClient(bookingService bookingPorts.BookingService, practiceRoomService bookingPorts.PracticeRoomService) authuserPorts.BookingClient {
	return &InProcessClient{bookingService: bookingService, practiceRoomService: practiceRoomService}
}

// EnsurePracticeRoom gives the practitioner the practice room. Unlike
// ClaimBookings its error reaches the caller: a practitioner without a room
// cannot publish availabilities.
func (c *InProcessClient) EnsurePracticeRoom(ctx context.Context, userID uuid.UUID) error {
	return c.practiceRoomService.EnsurePracticeRoom(ctx, userID)
}

// ClaimBookings links guest bookings to the newly created client account.
// Fire-and-forget: errors are logged but never propagated to the caller.
func (c *InProcessClient) ClaimBookings(ctx context.Context, clientID string, email string) {
	if clientID == "" || email == "" {
		slog.WarnContext(ctx, "booking claim: skipped, missing clientID or email",
			"client_id", clientID, "email", email)
		return
	}

	id, err := uuid.Parse(clientID)
	if err != nil {
		slog.ErrorContext(ctx, "booking claim: invalid clientID",
			"client_id", clientID, "err", err)
		return
	}

	claimed, err := c.bookingService.ClaimBookings(ctx, id, email)
	if err != nil {
		slog.ErrorContext(ctx, "booking claim: failed (non-blocking)",
			"client_id", clientID, "email", email, "err", err)
		return
	}

	if claimed > 0 {
		slog.InfoContext(ctx, "booking claim: succeeded",
			"client_id", clientID, "email", email, "claimed", claimed)
	}
}
