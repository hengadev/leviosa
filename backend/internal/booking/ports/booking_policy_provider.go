package ports

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
)

// BookingPolicyProvider supplies the practice's booking policy (minimum
// notice, booking horizon) read by the slot engine. In the modular monolith
// it reads the settings domain in-process, so a change made in the admin
// settings applies to the next slot query.
type BookingPolicyProvider interface {
	GetBookingPolicy(ctx context.Context) (domain.BookingPolicy, error)
}
