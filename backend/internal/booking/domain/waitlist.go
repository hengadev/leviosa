package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// WaitlistEntryStatus defines the lifecycle status of a waitlist entry.
//
// Full lifecycle: waiting -> notified -> claimed | expired
//
//	waiting | notified -> cancelled
//
// Issue 014 (this entity + the join-waitlist flow) only ever creates entries
// in the `waiting` state. The `notified`, `claimed`, and `expired` transitions
// are modeled here so the entity's invariants are complete and testable, but
// nothing in this codebase drives them yet — that's issue 015
// (waitlist-notify-and-claim), which is blocked and out of scope here.
type WaitlistEntryStatus string

const (
	WaitlistEntryStatusWaiting   WaitlistEntryStatus = "waiting"
	WaitlistEntryStatusNotified  WaitlistEntryStatus = "notified"
	WaitlistEntryStatusClaimed   WaitlistEntryStatus = "claimed"
	WaitlistEntryStatusExpired   WaitlistEntryStatus = "expired"
	WaitlistEntryStatusCancelled WaitlistEntryStatus = "cancelled"
)

// WaitlistEntry represents a client's (or guest's) request to be notified when
// a slot frees up for a product within a desired date range.
//
// Deliberately not a Booking: a waitlist entry has no slot, room allocation,
// or payment yet, so forcing it through the Booking FSM (ADR-0006) would
// violate that FSM's existing meaning.
type WaitlistEntry struct {
	ID uuid.UUID `json:"id"`

	// ProductID is the desired product/service. Encrypted for GDPR compliance,
	// mirroring Booking.ProductID — it reveals which health service a specific
	// client wants, at a specific time window.
	ProductID uuid.UUID `json:"product_id" encx:"encrypt"`

	// ClientID identifies a registered client. Nil for guest entries.
	ClientID *uuid.UUID `json:"client_id,omitempty"`

	// Desired date range the client would accept a slot within (encrypted, same
	// rationale as Booking.SlotStartTime/SlotEndTime).
	DesiredRangeStart time.Time `json:"desired_range_start" encx:"encrypt"`
	DesiredRangeEnd   time.Time `json:"desired_range_end" encx:"encrypt"`

	// Notes are optional free-form notes from the client (encrypted).
	Notes string `json:"notes,omitempty" encx:"encrypt"`

	// Guest contact fields (encrypted), used when ClientID is nil. Mirrors
	// Booking's guest-contact support (ADR-0008) — joining a waitlist must not
	// require account creation any more than booking does.
	GuestFirstName string `json:"guest_first_name,omitempty" encx:"encrypt"`
	GuestLastName  string `json:"guest_last_name,omitempty" encx:"encrypt"`
	GuestEmail     string `json:"guest_email,omitempty" encx:"encrypt"`
	GuestPhone     string `json:"guest_phone,omitempty" encx:"encrypt"`

	Status WaitlistEntryStatus `json:"status"`

	NotifiedAt  *time.Time `json:"notified_at,omitempty"`
	ClaimedAt   *time.Time `json:"claimed_at,omitempty"`
	ExpiredAt   *time.Time `json:"expired_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewWaitlistEntry creates a new waitlist entry in the `waiting` state.
// Either clientID must be non-nil, or guest contact fields must be provided —
// same identity rule as NewBooking (ADR-0008).
func NewWaitlistEntry(
	productID uuid.UUID,
	clientID *uuid.UUID,
	desiredRangeStart, desiredRangeEnd time.Time,
	notes string,
	guestFirstName, guestLastName, guestEmail, guestPhone string,
) (*WaitlistEntry, error) {
	if productID == uuid.Nil {
		return nil, ErrInvalidWaitlistProductID
	}

	if !desiredRangeEnd.After(desiredRangeStart) {
		return nil, ErrInvalidWaitlistDateRange
	}

	if desiredRangeEnd.Before(time.Now()) {
		return nil, ErrCannotJoinWaitlistInPast
	}

	isRegisteredClient := clientID != nil && *clientID != uuid.Nil
	isGuest := guestFirstName != "" && guestLastName != "" && (guestEmail != "" || guestPhone != "")

	if !isRegisteredClient && !isGuest {
		return nil, ErrInvalidWaitlistIdentity
	}
	if isRegisteredClient && isGuest {
		return nil, ErrAmbiguousWaitlistIdentity
	}

	now := time.Now()
	return &WaitlistEntry{
		ID:                uuid.New(),
		ProductID:         productID,
		ClientID:          clientID,
		DesiredRangeStart: desiredRangeStart,
		DesiredRangeEnd:   desiredRangeEnd,
		Notes:             notes,
		GuestFirstName:    guestFirstName,
		GuestLastName:     guestLastName,
		GuestEmail:        guestEmail,
		GuestPhone:        guestPhone,
		Status:            WaitlistEntryStatusWaiting,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// MarkNotified transitions a waiting entry to notified.
// Only valid from `waiting`.
func (w *WaitlistEntry) MarkNotified() error {
	if w.Status != WaitlistEntryStatusWaiting {
		return ErrWaitlistEntryNotWaiting
	}
	now := time.Now()
	w.Status = WaitlistEntryStatusNotified
	w.NotifiedAt = &now
	w.UpdatedAt = now
	return nil
}

// Claim transitions a notified entry to claimed.
// Only valid from `notified`.
func (w *WaitlistEntry) Claim() error {
	if w.Status != WaitlistEntryStatusNotified {
		return ErrWaitlistEntryNotNotified
	}
	now := time.Now()
	w.Status = WaitlistEntryStatusClaimed
	w.ClaimedAt = &now
	w.UpdatedAt = now
	return nil
}

// Expire transitions a notified entry to expired (the notification window
// elapsed without a claim). Only valid from `notified`.
func (w *WaitlistEntry) Expire() error {
	if w.Status != WaitlistEntryStatusNotified {
		return ErrWaitlistEntryNotNotified
	}
	now := time.Now()
	w.Status = WaitlistEntryStatusExpired
	w.ExpiredAt = &now
	w.UpdatedAt = now
	return nil
}

// Cancel transitions a waiting or notified entry to cancelled.
// Not valid once the entry has already reached a terminal state
// (claimed, expired, or cancelled).
func (w *WaitlistEntry) Cancel() error {
	switch w.Status {
	case WaitlistEntryStatusWaiting, WaitlistEntryStatusNotified:
		// allowed
	case WaitlistEntryStatusCancelled:
		return ErrWaitlistEntryAlreadyCancelled
	default:
		return ErrCannotCancelWaitlistEntry
	}
	now := time.Now()
	w.Status = WaitlistEntryStatusCancelled
	w.CancelledAt = &now
	w.UpdatedAt = now
	return nil
}

// IsGuestEntry returns true if this waitlist entry was joined without a
// registered account.
func (w *WaitlistEntry) IsGuestEntry() bool {
	return w.ClientID == nil
}

// GuestDisplayName returns the full name of the guest, or empty string if
// this is not a guest entry.
func (w *WaitlistEntry) GuestDisplayName() string {
	parts := make([]string, 0, 2)
	if w.GuestFirstName != "" {
		parts = append(parts, w.GuestFirstName)
	}
	if w.GuestLastName != "" {
		parts = append(parts, w.GuestLastName)
	}
	return strings.Join(parts, " ")
}

// IsActive returns true if the entry is still eligible to be matched against
// freed availability (i.e. hasn't reached a terminal state).
func (w *WaitlistEntry) IsActive() bool {
	return w.Status == WaitlistEntryStatusWaiting || w.Status == WaitlistEntryStatusNotified
}
