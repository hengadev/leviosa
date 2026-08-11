package ports

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/google/uuid"
)

// WaitlistRepository defines the interface for waitlist entry persistence.
type WaitlistRepository interface {
	// Create stores a new waitlist entry.
	Create(ctx context.Context, entry *domain.WaitlistEntryEncx) error

	// GetByID retrieves a waitlist entry by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WaitlistEntryEncx, error)
}
