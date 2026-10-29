package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// BookingNotificationService defines the interface for sending booking-related notifications.
// This interface is implemented by the notification module and injected into the booking service.
type BookingNotificationService interface {
	// SendBookingConfirmation sends a confirmation notification to the client after a booking is created.
	SendBookingConfirmation(ctx context.Context, data BookingNotificationData) error

	// SendBookingCancellation sends a cancellation notification to both client and partner.
	SendBookingCancellation(ctx context.Context, data BookingNotificationData) error

	// SendBookingReminder sends a reminder notification to the client before their appointment.
	SendBookingReminder(ctx context.Context, data BookingNotificationData) error

	// SendPaymentConfirmation sends a payment confirmation notification to the client.
	SendPaymentConfirmation(ctx context.Context, data BookingNotificationData) error

	// SendPaymentFailed sends a notification when payment processing fails.
	SendPaymentFailed(ctx context.Context, data BookingNotificationData) error

	// SendPaymentDisputeAlert notifies admin/staff that a client has disputed a charge
	// with their bank, linked to the specific booking. No automatic action (refund or
	// cancellation) is taken — this only makes the dispute visible for manual response.
	SendPaymentDisputeAlert(ctx context.Context, data BookingNotificationData) error
}

// LatePaymentNotificationService tells a client their payment, which arrived
// after the payment hold elapsed, was refunded in full because the slot was
// no longer available (ADR-0018). Implemented by the notification module next
// to BookingNotificationService; data.RefundedAmountCents is set.
type LatePaymentNotificationService interface {
	SendLatePaymentRefund(ctx context.Context, data BookingNotificationData) error
}

// BookingNotificationData contains information needed to send booking notifications.
// Required fields (IDs) are always populated by the booking service.
// Optional fields may be populated if available, otherwise the notification
// service implementation is responsible for fetching additional details.
type BookingNotificationData struct {
	// Required: Booking identifiers
	BookingID uuid.UUID
	ClientID  uuid.UUID
	PartnerID uuid.UUID
	RoomID    uuid.UUID
	ProductID uuid.UUID

	// Guest booking flag
	IsGuestBooking bool
	GuestEmail     string
	GuestPhone     string

	// Required: Appointment timing
	SlotStartTime time.Time
	SlotEndTime   time.Time

	// Required: Payment details
	TotalPriceCents int
	Currency        string

	// AmountPaidOnlineCents is what the client paid online for the booking
	// (the deposit, or the full price); set on booking-confirmed events. The
	// balance due at the practice is TotalPriceCents minus this amount.
	AmountPaidOnlineCents int

	// Locale the client booked in ("fr" or "en"); empty on older bookings,
	// read as French. Client emails are written in it.
	Locale string

	// Optional: Pre-populated details (if available)
	// The notification service implementation may fetch these if not provided
	ClientEmail string
	ClientName  string
	// ClientFirstName / ClientLastName, when known (guest bookings), avoid
	// splitting ClientName.
	ClientFirstName string
	ClientLastName  string
	ClientPhone     string
	PartnerEmail    string
	PartnerName     string
	ProductName     string
	RoomName        string
	BuildingName    string
	Address         string

	// Booking token for public guest access (empty for bookings created before issue 002)
	Token string

	// RefundedAmountCents is what was refunded to the client (late-payment
	// refund and cancellation notifications).
	RefundedAmountCents int

	// Cancellation details (only for cancellation notifications).
	// CancelledBy is "client" or "practitioner"; empty when not recorded.
	CancellationReason string
	CancelledAt        *time.Time
	CancelledBy        string

	// Dispute details (only for dispute alert notifications)
	DisputeReason string
}
