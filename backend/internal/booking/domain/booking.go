package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// BookingStatus defines the lifecycle status of a booking.
//
// A booking starts awaiting payment and holds its slot during the payment
// hold (ADR-0018):
//
//	awaiting_payment → confirmed | expired
//	confirmed → completed | cancelled | no_show
type BookingStatus string

const (
	BookingStatusAwaitingPayment BookingStatus = "awaiting_payment"
	BookingStatusExpired         BookingStatus = "expired"
	BookingStatusConfirmed       BookingStatus = "confirmed"
	BookingStatusCancelled       BookingStatus = "cancelled"
	BookingStatusCompleted       BookingStatus = "completed"
	BookingStatusNoShow          BookingStatus = "no_show"
)

// NonAppointmentStatuses are the statuses of bookings that are not real
// appointments: a booking still awaiting payment, or one whose hold expired
// unpaid. Agenda, earnings, reminders and financial views exclude them
// (ADR-0018).
func NonAppointmentStatuses() []BookingStatus {
	return []BookingStatus{BookingStatusAwaitingPayment, BookingStatusExpired}
}

// IsAppointment reports whether the booking is a real appointment, i.e. not
// awaiting payment and not expired.
func (b *Booking) IsAppointment() bool {
	return b.Status != BookingStatusAwaitingPayment && b.Status != BookingStatusExpired
}

// BookingLocale is the language the client booked in; notifications about
// the booking are sent in it.
type BookingLocale string

const (
	BookingLocaleFR BookingLocale = "fr"
	BookingLocaleEN BookingLocale = "en"
)

// ParseBookingLocale accepts "fr" or "en"; an empty value is French, the
// practice's default language.
func ParseBookingLocale(value string) (BookingLocale, error) {
	switch BookingLocale(value) {
	case "", BookingLocaleFR:
		return BookingLocaleFR, nil
	case BookingLocaleEN:
		return BookingLocaleEN, nil
	default:
		return "", ErrInvalidBookingLocale
	}
}

// PaymentStatus defines the payment state of a booking
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
	PaymentStatusDisputed PaymentStatus = "disputed"
)

// Booking represents a client reservation of partner availability
type Booking struct {
	ID             uuid.UUID  `json:"id"`
	AvailabilityID uuid.UUID  `json:"availability_id"`
	ClientID       *uuid.UUID `json:"client_id,omitempty"`
	PartnerID      uuid.UUID  `json:"partner_id"`
	RoomID         uuid.UUID  `json:"room_id"`

	// Slot information (product and time boundaries)
	ProductID     uuid.UUID `json:"product_id" encx:"encrypt"`      // Link to catalog product
	SlotStartTime time.Time `json:"slot_start_time" encx:"encrypt"` // Actual slot start (a slot the slot engine offered)
	SlotEndTime   time.Time `json:"slot_end_time" encx:"encrypt"`   // Calculated: SlotStartTime + Product.Duration

	// Legacy notes columns (kept in the schema, never collected or shown:
	// Leviosa stores no health data — ADR-0019). PartnerNotes is written only by
	// internal bookkeeping (refund tracking); ClientNotes is never written.
	ClientNotes  string `json:"-" encx:"encrypt"`
	PartnerNotes string `json:"-" encx:"encrypt"`

	// Guest contact fields (encrypted for GDPR compliance, used when ClientID is nil)
	GuestFirstName string `json:"guest_first_name,omitempty" encx:"encrypt"`
	GuestLastName  string `json:"guest_last_name,omitempty" encx:"encrypt"`
	GuestEmail     string `json:"guest_email,omitempty" encx:"encrypt"`
	GuestPhone     string `json:"guest_phone,omitempty" encx:"encrypt"`

	// Pricing information
	TotalPriceCents int    `json:"total_price_cents"`
	Currency        string `json:"currency"`

	// Payment tracking
	PaymentStatus   PaymentStatus `json:"payment_status"`
	PaymentIntentID *string       `json:"payment_intent_id,omitempty"`

	// Booking lifecycle
	Status             BookingStatus `json:"status"`
	CancelledAt        *time.Time    `json:"cancelled_at,omitempty"`
	CancellationReason string        `json:"cancellation_reason,omitempty" encx:"encrypt"`
	CompletedAt        *time.Time    `json:"completed_at,omitempty"`

	// Booking token for public guest access
	Token string `json:"token,omitempty"`

	// Reminder tracking
	RemindedAt *time.Time `json:"reminded_at,omitempty"`

	// Payment hold (ADR-0018): while awaiting payment, the booking keeps its
	// slot until HoldExpiresAt.
	HoldExpiresAt *time.Time `json:"hold_expires_at,omitempty"`

	// Locale the client booked in (fr or en); empty on bookings made before
	// it was recorded, read as French.
	Locale BookingLocale `json:"locale,omitempty"`

	// PaymentClientSecret is the PaymentIntent's client secret, returned to
	// the client that just created the booking so the page can mount the
	// Stripe Payment Element. Never persisted.
	PaymentClientSecret string `json:"-"`

	// AmountDueOnlineCents is what the PaymentIntent charges (the deposit, or
	// the full price), set on the booking just created. Never persisted.
	AmountDueOnlineCents int `json:"-"`

	// Administrative fields
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewBooking creates a new booking, awaiting payment (ADR-0018): it becomes
// confirmed only once its payment succeeds (Confirm), or expired when its
// hold elapses unpaid (Expire).
// Either clientID must be non-nil, or guest contact fields must be provided.
// For guest bookings: guestFirstName, guestLastName, and at least one of guestEmail or guestPhone are required.
func NewBooking(availabilityID uuid.UUID, clientID *uuid.UUID, partnerID, roomID uuid.UUID, totalPriceCents int, currency string, guestFirstName, guestLastName, guestEmail, guestPhone string) (*Booking, error) {
	if availabilityID == uuid.Nil {
		return nil, ErrInvalidAvailabilityID
	}
	if partnerID == uuid.Nil {
		return nil, ErrInvalidPartnerID
	}
	if roomID == uuid.Nil {
		return nil, ErrInvalidRoomID
	}
	if totalPriceCents < 0 {
		return nil, ErrInvalidBookingPrice
	}

	// Validate identity: either a registered client or guest contact info
	isRegisteredClient := clientID != nil && *clientID != uuid.Nil
	isGuest := guestFirstName != "" && guestLastName != "" && (guestEmail != "" || guestPhone != "")

	if !isRegisteredClient && !isGuest {
		return nil, ErrInvalidClientID
	}
	if isRegisteredClient && isGuest {
		return nil, ErrAmbiguousBookingIdentity
	}

	if currency == "" {
		currency = "EUR" // Default currency
	}

	return &Booking{
		ID:              uuid.New(),
		AvailabilityID:  availabilityID,
		ClientID:        clientID,
		PartnerID:       partnerID,
		RoomID:          roomID,
		TotalPriceCents: totalPriceCents,
		Currency:        currency,
		PaymentStatus:   PaymentStatusPending,
		Status:          BookingStatusAwaitingPayment,
		GuestFirstName:  guestFirstName,
		GuestLastName:   guestLastName,
		GuestEmail:      guestEmail,
		GuestPhone:      guestPhone,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

// SetPartnerNotes records internal bookkeeping in the legacy partner-notes
// column (e.g. refund tracking). It is never exposed through the API
// (ADR-0019: no health data).
func (b *Booking) SetPartnerNotes(notes string) {
	b.PartnerNotes = notes
	b.UpdatedAt = time.Now()
}

// SetPaymentIntentID sets the Stripe payment intent ID
func (b *Booking) SetPaymentIntentID(paymentIntentID string) {
	b.PaymentIntentID = &paymentIntentID
	b.UpdatedAt = time.Now()
}

// MarkPaymentPaid marks the booking payment as paid
func (b *Booking) MarkPaymentPaid() error {
	if b.PaymentStatus == PaymentStatusRefunded {
		return ErrCannotMarkRefundedAsPaid
	}
	b.PaymentStatus = PaymentStatusPaid
	b.UpdatedAt = time.Now()
	return nil
}

// MarkPaymentFailed marks the booking payment as failed
func (b *Booking) MarkPaymentFailed() {
	b.PaymentStatus = PaymentStatusFailed
	b.UpdatedAt = time.Now()
}

// RefundPayment marks the booking payment as refunded
func (b *Booking) RefundPayment() error {
	if b.PaymentStatus != PaymentStatusPaid {
		return ErrCannotRefundUnpaidBooking
	}
	b.PaymentStatus = PaymentStatusRefunded
	b.UpdatedAt = time.Now()
	return nil
}

// MarkPaymentDisputed marks the booking payment as disputed after a Stripe
// charge.dispute.created webhook. Only a paid booking can be disputed. No
// automatic refund or cancellation happens here — dispute response is a
// manual, evidence-gathering process (ADR-0007, issue 018).
func (b *Booking) MarkPaymentDisputed() error {
	if b.PaymentStatus != PaymentStatusPaid {
		return ErrCannotDisputeUnpaidBooking
	}
	b.PaymentStatus = PaymentStatusDisputed
	b.UpdatedAt = time.Now()
	return nil
}

// ResolveDispute transitions a disputed booking's payment status once Stripe
// closes the dispute (charge.dispute.closed). stripeDisputeStatus is the
// Stripe dispute's terminal status: "lost" means the cardholder's bank kept
// the chargeback, so the booking reverts to Refunded (the money is gone,
// mirroring what a manual refund would leave); any other outcome (won,
// warning_closed, etc.) reverts to Paid. This only records the outcome — it
// never triggers an actual Stripe refund call.
func (b *Booking) ResolveDispute(stripeDisputeStatus string) error {
	if b.PaymentStatus != PaymentStatusDisputed {
		return ErrCannotResolveNonDisputedBooking
	}
	if stripeDisputeStatus == "lost" {
		b.PaymentStatus = PaymentStatusRefunded
	} else {
		b.PaymentStatus = PaymentStatusPaid
	}
	b.UpdatedAt = time.Now()
	return nil
}

// HoldUntil starts the payment hold: the booking keeps its slot until
// expiresAt while it awaits payment.
func (b *Booking) HoldUntil(expiresAt time.Time) error {
	if b.Status != BookingStatusAwaitingPayment {
		return ErrBookingNotAwaitingPayment
	}
	b.HoldExpiresAt = &expiresAt
	b.UpdatedAt = time.Now()
	return nil
}

// Confirm moves a booking awaiting payment to confirmed, once its payment
// succeeded (or when nothing is due online). Any other status is refused, so
// a second confirmation of the same booking is detectable.
func (b *Booking) Confirm() error {
	if b.Status != BookingStatusAwaitingPayment {
		return ErrBookingNotAwaitingPayment
	}
	b.Status = BookingStatusConfirmed
	b.UpdatedAt = time.Now()
	return nil
}

// Expire moves a booking awaiting payment to expired: its hold elapsed
// unpaid and its slot is offered again.
func (b *Booking) Expire() error {
	if b.Status != BookingStatusAwaitingPayment {
		return ErrBookingNotAwaitingPayment
	}
	b.Status = BookingStatusExpired
	b.UpdatedAt = time.Now()
	return nil
}

// ConfirmLatePayment confirms a booking whose payment succeeded after its
// hold elapsed (ADR-0018), awaiting payment or already expired, once the
// caller made sure its slot is still free. Any other status is refused.
func (b *Booking) ConfirmLatePayment() error {
	if b.Status != BookingStatusAwaitingPayment && b.Status != BookingStatusExpired {
		return ErrBookingNotAwaitingPayment
	}
	b.Status = BookingStatusConfirmed
	b.UpdatedAt = time.Now()
	return nil
}

// Cancel cancels the booking with a reason
func (b *Booking) Cancel(reason string) error {
	if !b.IsAppointment() {
		return ErrBookingNotConfirmed
	}
	if b.Status == BookingStatusCompleted {
		return ErrCannotCancelCompletedBooking
	}
	if b.Status == BookingStatusCancelled {
		return ErrBookingAlreadyCancelled
	}

	now := time.Now()
	b.Status = BookingStatusCancelled
	b.CancelledAt = &now
	b.CancellationReason = reason
	b.UpdatedAt = now
	return nil
}

// Complete marks the booking as completed
func (b *Booking) Complete() error {
	if !b.IsAppointment() {
		return ErrBookingNotConfirmed
	}
	if b.Status == BookingStatusCancelled {
		return ErrCannotCompleteBooking
	}
	if b.Status == BookingStatusCompleted {
		return ErrBookingAlreadyCompleted
	}

	now := time.Now()
	b.Status = BookingStatusCompleted
	b.CompletedAt = &now
	b.UpdatedAt = now
	return nil
}

// MarkNoShow marks the booking as a no-show
func (b *Booking) MarkNoShow() error {
	if !b.IsAppointment() {
		return ErrBookingNotConfirmed
	}
	if b.Status == BookingStatusCancelled {
		return ErrCannotMarkCancelledAsNoShow
	}
	if b.Status == BookingStatusCompleted {
		return ErrCannotMarkCompletedAsNoShow
	}

	b.Status = BookingStatusNoShow
	b.UpdatedAt = time.Now()
	return nil
}

// IsGuestBooking returns true if this booking was made without a registered account.
func (b *Booking) IsGuestBooking() bool {
	return b.ClientID == nil
}

// GuestDisplayName returns the full name of the guest, or empty string if not a guest booking.
func (b *Booking) GuestDisplayName() string {
	parts := make([]string, 0, 2)
	if b.GuestFirstName != "" {
		parts = append(parts, b.GuestFirstName)
	}
	if b.GuestLastName != "" {
		parts = append(parts, b.GuestLastName)
	}
	return strings.Join(parts, " ")
}

// IsActive checks if the booking is in an active state
func (b *Booking) IsActive() bool {
	return b.Status == BookingStatusConfirmed
}

// IsCancellable checks if the booking can be cancelled
func (b *Booking) IsCancellable() bool {
	return b.Status == BookingStatusConfirmed
}

// RequiresPayment checks if the booking still requires payment
func (b *Booking) RequiresPayment() bool {
	return b.PaymentStatus == PaymentStatusPending || b.PaymentStatus == PaymentStatusFailed
}

// SetToken sets the signed booking token for public guest access.
func (b *Booking) SetToken(token string) {
	b.Token = token
	b.UpdatedAt = time.Now()
}
