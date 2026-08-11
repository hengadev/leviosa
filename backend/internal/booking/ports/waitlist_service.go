package ports

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
)

// WaitlistService defines the interface for waitlist business logic.
type WaitlistService interface {
	// JoinWaitlist creates a new waitlist entry in the `waiting` state for a
	// product/desired-date-range combination, on behalf of a registered client
	// or a guest (mirroring the guest-booking identity rule from ADR-0008).
	JoinWaitlist(ctx context.Context, request *domain.JoinWaitlistRequest) (*domain.WaitlistEntryResponse, error)
}
