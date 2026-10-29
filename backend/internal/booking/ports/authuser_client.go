package ports

import (
	"context"

	"github.com/google/uuid"
)

// AuthUserClient defines the interface for communicating with the authuser service.
// This interface provides access to partner information needed by the booking module.
//
// Implementation Strategy:
// - Modular Monolith: In-process implementation delegates to authuser.PartnerService
// - Microservices: HTTP-based implementation uses ServiceClient for authenticated calls
type AuthUserClient interface {
	// GetPartnerVerificationStatus checks if a partner is verified by partner ID.
	// Returns false if partner doesn't exist (no error for not found).
	GetPartnerVerificationStatus(ctx context.Context, partnerID uuid.UUID) (bool, error)

	// GetPartnerByUserID retrieves a partner by their user ID.
	// This is needed because allocations are created using user IDs, not partner IDs.
	// Returns error if partner not found or if there's a system error.
	GetPartnerByUserID(ctx context.Context, userID uuid.UUID) (*PartnerInfo, error)

	// GetPartnerUserID resolves a partner ID (the public partner profile, as
	// the practice page settings hold it) to the partner's user ID, which
	// availabilities and bookings reference.
	GetPartnerUserID(ctx context.Context, partnerID uuid.UUID) (uuid.UUID, error)

	// GetUserName returns the display name ("FirstName LastName") for a user.
	// Returns an empty string if the user is not found.
	GetUserName(ctx context.Context, userID uuid.UUID) (string, error)

	// GetUserEmail returns the email address for a user by their ID.
	// Used to set receipt_email on Stripe payment intents for authenticated
	// clients so Stripe emails them a hosted receipt automatically.
	// Returns an empty string if the user is not found.
	GetUserEmail(ctx context.Context, userID uuid.UUID) (string, error)

	// GetUserContact returns the display name, email and phone for a user by
	// their ID, resolved with a single lookup. Used by the partner agenda so a
	// partner can contact a registered client (name, email, phone only — no
	// notes or health data, ADR-0019). Empty strings when a value is unset.
	GetUserContact(ctx context.Context, userID uuid.UUID) (name, email, phone string, err error)
}

// PartnerInfo represents basic partner information needed by booking module
type PartnerInfo struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	IsVerified bool      `json:"is_verified"`
}
