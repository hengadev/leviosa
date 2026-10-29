package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingResponse struct {
	ID                 uuid.UUID             `json:"id"`
	AvailabilityID     uuid.UUID             `json:"availability_id"`
	ClientID           *uuid.UUID            `json:"client_id,omitempty"`
	PartnerID          uuid.UUID             `json:"partner_id"`
	RoomID             uuid.UUID             `json:"room_id"`
	ProductID          uuid.UUID             `json:"product_id"`
	SlotStartTime      time.Time             `json:"slot_start_time"`
	SlotEndTime        time.Time             `json:"slot_end_time"`
	Status             BookingStatus         `json:"status"`
	TotalPriceCents    int                   `json:"total_price_cents"`
	Currency           string                `json:"currency"`
	PaymentStatus      PaymentStatus         `json:"payment_status"`
	PaymentIntentID    *string               `json:"payment_intent_id,omitempty"`
	ReceiptURL         string                `json:"receipt_url,omitempty"`
	// Notes (client_notes/partner_notes) are deliberately absent from every
	// response: Leviosa stores no health data (ADR-0019). The DB columns stay.
	CancellationReason *string               `json:"cancellation_reason,omitempty"`
	CancelledAt        *time.Time            `json:"cancelled_at,omitempty"`
	CompletedAt        *time.Time            `json:"completed_at,omitempty"`
	Availability       *AvailabilityResponse `json:"availability,omitempty"`

	// Who cancelled and what the cancellation refunded (ADR-0014).
	CancelledBy         CancellationActor `json:"cancelled_by,omitempty"`
	RefundedAmountCents *int              `json:"refunded_amount_cents,omitempty"`

	// Booking token for public guest access
	Token string `json:"token,omitempty"`

	// Guest contact fields (populated for guest bookings)
	GuestFirstName string `json:"guest_first_name,omitempty"`
	GuestLastName  string `json:"guest_last_name,omitempty"`
	GuestEmail     string `json:"guest_email,omitempty"`
	GuestPhone     string `json:"guest_phone,omitempty"`

	// Financial status: plain-language-ready cancellation/refund fields, computed
	// from the booking's payment status and the product's cancellation window.
	// See BookingFinancialStatus for field semantics.
	*BookingFinancialStatus
}

// BookingFinancialStatus holds the derived fields a client UI needs to render
// plain-language payment/refund status and the cancellation deadline, without
// the frontend re-implementing cancellation-window or refund-amount logic.
//
// RefundAmountCents/RefundedAt are intentionally generic (an amount + a
// timestamp) rather than assuming "full refund or nothing": today's refund
// flow (see BookingService.RefundBooking) only ever issues a full refund, but
// this shape is written so the graduated/partial refund policy planned in
// issue 017 can populate the same fields with a different amount without any
// rework on the reading side.
type BookingFinancialStatus struct {
	// CancellationDeadline is the last moment the booking can be cancelled
	// free of consequence, derived from the product's CancellationHours. Nil
	// when the product enforces no cancellation window, or the booking is no
	// longer in a cancellable state.
	CancellationDeadline *time.Time `json:"cancellation_deadline,omitempty"`
	// CancellationHours is the product's cancellation notice window, echoed
	// back so the UI can phrase copy without a second lookup.
	CancellationHours *int `json:"cancellation_hours,omitempty"`
	// RefundAmountCents is set once a refund exists for this booking.
	RefundAmountCents *int `json:"refund_amount_cents,omitempty"`
	// RefundedAt is the best-known timestamp the refund was processed.
	RefundedAt *time.Time `json:"refunded_at,omitempty"`
}

// Booking DTOs
type CreateBookingRequest struct {
	AvailabilityID uuid.UUID `json:"availability_id" validate:"required"`
	ClientID       *uuid.UUID `json:"client_id,omitempty"`
	ProductID      uuid.UUID `json:"product_id" validate:"required"`
	SlotStartTime  time.Time `json:"slot_start_time" validate:"required"`
	// No client_notes: Leviosa stores no health data (ADR-0019). The unknown-
	// fields decoder rejects the key if a client sends it anyway.

	// Guest fields (required when client_id is omitted)
	GuestFirstName string `json:"guest_first_name,omitempty"`
	GuestLastName  string `json:"guest_last_name,omitempty"`
	GuestEmail     string `json:"guest_email,omitempty"`
	GuestPhone     string `json:"guest_phone,omitempty"`

	// Locale is the language the client books in, "fr" (default) or "en":
	// notifications about the booking are sent in it.
	Locale string `json:"locale,omitempty"`
}

// CreateBookingResponse is the booking just created, plus what the page
// needs to collect the payment (ADR-0018): the PaymentIntent's client secret
// for the Stripe Payment Element, the amount it charges and the end of the
// payment hold.
type CreateBookingResponse struct {
	BookingResponse
	ClientSecret         string     `json:"client_secret,omitempty"`
	AmountDueOnlineCents int        `json:"amount_due_online_cents"`
	HoldExpiresAt        *time.Time `json:"hold_expires_at,omitempty"`
	Locale               string     `json:"locale,omitempty"`
}

// CancelBookingRequest is the body of the admin override cancellation. The
// client cancellation takes no body: no free-text field in the cancellation
// flow (ADR-0019).
type CancelBookingRequest struct {
	Reason string `json:"reason" validate:"required,min=1,max=500"`
}

type Transaction struct {
	ID              uuid.UUID     `json:"id"`
	SlotStartTime   string        `json:"slot_start_time"`
	ProductID       uuid.UUID     `json:"product_id"`
	ProductName     string        `json:"product_name"`
	AmountCents     int           `json:"amount_cents"`
	PaymentStatus   PaymentStatus `json:"payment_status"`
	BookingStatus   BookingStatus `json:"booking_status"`
}

type EarningsSummary struct {
	CurrentMonthCents int    `json:"current_month_cents"`
	LastMonthCents    int    `json:"last_month_cents"`
	PendingCents      int    `json:"pending_cents"`
	NextPayoutDate    string `json:"next_payout_date"`
	NextPayoutCents   int    `json:"next_payout_cents"`
	Transactions      []Transaction `json:"transactions"`
}

// PartnerBookingResponse is the enriched booking DTO returned by the partner bookings endpoint.
// It includes resolved names for display in the partner agenda UI.
type PartnerBookingResponse struct {
	ID              uuid.UUID     `json:"id"`
	ClientID        *uuid.UUID    `json:"client_id,omitempty"`
	ClientName      string        `json:"client_name"`
	ClientEmail     string        `json:"client_email,omitempty"`
	ClientPhone     string        `json:"client_phone,omitempty"`
	ProductName     string        `json:"product_name"`
	RoomName        string        `json:"room_name"`
	SlotStartTime   time.Time     `json:"slot_start_time"`
	SlotEndTime     time.Time     `json:"slot_end_time"`
	Status          BookingStatus `json:"status"`
	PaymentStatus   PaymentStatus `json:"payment_status"`
	TotalPriceCents int           `json:"total_price_cents"`
	Currency        string        `json:"currency"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty"`
}

// AdminBookingResponse is the enriched booking DTO returned by the admin list endpoint.
type AdminBookingResponse struct {
	ID              uuid.UUID     `json:"id"`
	ClientName      string        `json:"client_name"`
	PartnerName     string        `json:"partner_name"`
	ProductName     string        `json:"product_name"`
	RoomName        string        `json:"room_name"`
	SlotStartTime   time.Time     `json:"slot_start_time"`
	SlotEndTime     time.Time     `json:"slot_end_time"`
	Status          BookingStatus `json:"status"`
	PaymentStatus   PaymentStatus `json:"payment_status"`
	TotalPriceCents int           `json:"total_price_cents"`
	Currency        string        `json:"currency"`
	CreatedAt       time.Time     `json:"created_at"`
}

// PublicBookingLookupResponse is the DTO returned by the unauthenticated public
// booking lookup endpoint. It contains only the fields safe to show to guests:
// no ClientID, no internal partner/room IDs, and no decrypted guest contact fields.
type PublicBookingLookupResponse struct {
	ID              uuid.UUID     `json:"id"`
	ProductName     string        `json:"product_name"`
	PartnerName     string        `json:"partner_name"`
	SlotStartTime   time.Time     `json:"slot_start_time"`
	SlotEndTime     time.Time     `json:"slot_end_time"`
	Status          BookingStatus `json:"status"`
	TotalPriceCents int           `json:"total_price_cents"`
	Currency        string        `json:"currency"`
	PaymentStatus   PaymentStatus `json:"payment_status"`

	// Locale the client booked in ("fr" or "en"); the booking page is shown
	// in it.
	Locale BookingLocale `json:"locale"`

	// AmountPaidOnlineCents is what the client paid online (the deposit, or
	// the full price); the balance is paid at the practice.
	AmountPaidOnlineCents int `json:"amount_paid_online_cents"`

	// RefundPreview is what the client would get back if they cancelled now;
	// set only while the booking can be cancelled. The cancellation applies
	// the same rule.
	RefundPreview *RefundPreview `json:"refund_preview,omitempty"`

	// Who cancelled and what was refunded, once cancelled.
	CancelledBy         CancellationActor `json:"cancelled_by,omitempty"`
	CancelledAt         *time.Time        `json:"cancelled_at,omitempty"`
	RefundedAmountCents *int              `json:"refunded_amount_cents,omitempty"`
}

// RefundPreview is the refund a cancellation made now would give.
type RefundPreview struct {
	AmountCents int `json:"amount_cents"`
	Percentage  int `json:"percentage"`
	// MinHoursBefore is the threshold of the cancellation tier that applies;
	// nil when the product has no tiers (the whole amount is refunded).
	MinHoursBefore *int `json:"min_hours_before,omitempty"`
}

// AdminBookingsListResponse is the paginated response for the admin bookings list.
type AdminBookingsListResponse struct {
	Bookings []AdminBookingResponse `json:"bookings"`
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	Limit    int                    `json:"limit"`
}
