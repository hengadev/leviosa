package ports

import (
	"context"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/google/uuid"
)

// BookingRepository defines the interface for booking data persistence
type BookingRepository interface {
	// Create stores a new booking
	Create(ctx context.Context, booking *domain.BookingEncx) error

	// CreateGuarded stores a new booking under a lock on its partner's
	// agenda (and on its availability). verify receives the partner's
	// bookings that may occupy time (every status but cancelled and expired,
	// in any of the partner's availabilities) as they stand once the lock is
	// held, and refuses the insert by returning an error (returned as is).
	// Two clients booking the same partner at the same moment are serialised
	// here, so the second one sees the first one's booking.
	CreateGuarded(ctx context.Context, booking *domain.BookingEncx, verify func(existing []*domain.BookingEncx) error) error

	// ConfirmAwaitingPayment moves a booking from awaiting_payment to
	// confirmed with its payment paid, only if it is still awaiting payment.
	// It reports whether this call made the transition, so a payment webhook
	// delivered twice (even concurrently) confirms the booking once.
	ConfirmAwaitingPayment(ctx context.Context, id uuid.UUID) (bool, error)

	// ConfirmGuarded confirms a booking awaiting payment or expired (a
	// payment that succeeded after its hold elapsed, ADR-0018) under the same
	// lock on its partner's agenda as CreateGuarded, so no booking can take
	// the slot while it is checked. verify receives the booking as it stands
	// once locked and the partner's other bookings that may occupy time; it
	// refuses the confirmation by returning an error (returned as is). It
	// reports false, without calling verify, when the booking is no longer
	// awaiting payment or expired (a concurrent delivery confirmed it).
	ConfirmGuarded(ctx context.Context, id, partnerID uuid.UUID, verify func(current *domain.BookingEncx, others []*domain.BookingEncx) error) (bool, error)

	// ExpireElapsedHolds moves every booking still awaiting payment whose
	// hold expired at or before now to expired, in one conditional update
	// (a booking confirmed meanwhile is left alone), and returns those it
	// moved. Two callers at once never both move the same booking.
	ExpireElapsedHolds(ctx context.Context, now time.Time) ([]ExpiredHold, error)

	// ExpireHold moves one booking awaiting payment whose hold expired at or
	// before now to expired, and reports whether this call moved it.
	ExpireHold(ctx context.Context, id uuid.UUID, now time.Time) (bool, error)

	// ClaimLatePaymentRefund marks the payment of an expired booking as
	// refunded, only if it is not already: of two deliveries of the same late
	// payment, only the one that gets true refunds it. ReleaseLatePaymentRefund
	// undoes the claim (back to paid) when the refund could not be made.
	ClaimLatePaymentRefund(ctx context.Context, id uuid.UUID) (bool, error)
	ReleaseLatePaymentRefund(ctx context.Context, id uuid.UUID) error

	// GetByID retrieves a booking by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingEncx, error)

	// Update modifies an existing booking
	Update(ctx context.Context, booking *domain.BookingEncx) error

	// Delete removes a booking (hard delete for GDPR compliance)
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves bookings with optional filtering
	List(ctx context.Context, filter BookingFilter) ([]*domain.BookingEncx, error)

	// GetByClientID retrieves all bookings for a specific client
	GetByClientID(ctx context.Context, clientID uuid.UUID, filter BookingFilter) ([]*domain.BookingEncx, error)

	// GetByPartnerID retrieves all bookings for a specific partner
	GetByPartnerID(ctx context.Context, partnerID uuid.UUID, filter BookingFilter) ([]*domain.BookingEncx, error)

	// GetByAvailabilityID retrieves booking for a specific availability (should be max 1)
	// DEPRECATED: Use GetBookingsByAvailability for slot-based booking system
	GetByAvailabilityID(ctx context.Context, availabilityID uuid.UUID) (*domain.BookingEncx, error)

	// GetBookingsByAvailability retrieves all bookings for a specific availability.
	// Used for slot overlap detection in the slot-based booking system.
	// Returns bookings sorted by slot_start_time ASC.
	GetBookingsByAvailability(ctx context.Context, availabilityID uuid.UUID) ([]*domain.BookingEncx, error)

	// GetUpcoming retrieves upcoming bookings (status = confirmed, start time in future)
	GetUpcoming(ctx context.Context, filter BookingFilter) ([]*domain.BookingEncx, error)

	// GetByPaymentIntentID retrieves booking by Stripe payment intent ID
	GetByPaymentIntentID(ctx context.Context, paymentIntentID string) (*domain.BookingEncx, error)

	// GetGuestBookings retrieves all bookings that have no client_id (guest bookings).
	GetGuestBookings(ctx context.Context) ([]*domain.BookingEncx, error)

	// SetBookingClientID atomically sets client_id on a single booking.
	// Only succeeds if the booking's client_id is currently NULL.
	// Returns true if the row was updated, false if it was already owned.
	SetBookingClientID(ctx context.Context, bookingID, clientID uuid.UUID) (bool, error)

	// FindBookingsDueForReminder returns all confirmed bookings where
	// reminded_at IS NULL. Time filtering (slot_start_time within the reminder
	// window) is performed at the service layer after decryption because
	// slot_start_time is stored encrypted for GDPR compliance.
	FindBookingsDueForReminder(ctx context.Context) ([]*domain.BookingEncx, error)

	// MarkReminderSent sets reminded_at to now() for a given booking.
	MarkReminderSent(ctx context.Context, bookingID uuid.UUID) error

	// ClaimReminder sets reminded_at to now() only when the booking is still
	// confirmed and not yet reminded, and reports whether this call did it.
	// The reminder is sent only after a successful claim, so each booking is
	// reminded once even with several scheduler ticks or instances.
	ClaimReminder(ctx context.Context, bookingID uuid.UUID) (bool, error)
}

// ExpiredHold is a booking the expiry job moved to expired, with the
// PaymentIntent left to cancel.
type ExpiredHold struct {
	BookingID       uuid.UUID
	PaymentIntentID *string
}

// BookingFilter defines filtering options for booking queries
type BookingFilter struct {
	// Entity filters
	ClientID       *uuid.UUID
	PartnerID      *uuid.UUID
	RoomID         *uuid.UUID
	AvailabilityID *uuid.UUID

	// Status filters. ExcludeStatus drops bookings in those statuses even
	// when Status names them: views of real appointments pass
	// domain.NonAppointmentStatuses() (ADR-0018).
	Status        []domain.BookingStatus
	ExcludeStatus []domain.BookingStatus
	PaymentStatus []domain.PaymentStatus

	// Time-based filters
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	TimeRange     *TimeRange // Filter by availability time range

	// Price filters
	MinPrice *int // In cents
	MaxPrice *int // In cents

	// Pagination
	Limit  int
	Offset int

	// Sorting
	OrderBy        string // "created_at", "total_price_cents"
	OrderDirection string // "asc", "desc"
}
