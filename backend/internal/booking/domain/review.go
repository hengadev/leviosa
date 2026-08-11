package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	// MinReviewRating and MaxReviewRating bound the star rating (inclusive).
	MinReviewRating = 1
	MaxReviewRating = 5

	// MaxReviewCommentLength caps the optional free-text comment.
	MaxReviewCommentLength = 2000
)

// Review represents a client's rating and optional comment for a completed booking.
// At most one review may exist per booking (enforced by a unique constraint on
// booking_id at the persistence layer). Reviews are meant for public display —
// see docs/prd/public-trust-and-seo-foundations.md — so, unlike most booking domain
// entities, they are not encx-encrypted.
type Review struct {
	ID        uuid.UUID `json:"id"`
	BookingID uuid.UUID `json:"booking_id"`
	ClientID  uuid.UUID `json:"client_id"`
	PartnerID uuid.UUID `json:"partner_id"`
	ProductID uuid.UUID `json:"product_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment,omitempty"`
	Hidden    bool      `json:"hidden"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewReview constructs a new Review. It validates only the review's own invariants
// (rating bounds, comment length) — the booking-relationship authorization rules
// (booking must be `completed`, requester must be the booking's client, at most one
// review per booking) are enforced by the application layer before this constructor
// is called, mirroring the pattern ADR-0013 establishes for messaging.
func NewReview(bookingID, clientID, partnerID, productID uuid.UUID, rating int, comment string) (*Review, error) {
	if bookingID == uuid.Nil {
		return nil, ErrInvalidReviewBookingID
	}
	if clientID == uuid.Nil {
		return nil, ErrInvalidClientID
	}
	if partnerID == uuid.Nil {
		return nil, ErrInvalidPartnerID
	}
	if rating < MinReviewRating || rating > MaxReviewRating {
		return nil, ErrInvalidReviewRating
	}
	if len(comment) > MaxReviewCommentLength {
		return nil, ErrReviewCommentTooLong
	}

	now := time.Now()
	return &Review{
		ID:        uuid.New(),
		BookingID: bookingID,
		ClientID:  clientID,
		PartnerID: partnerID,
		ProductID: productID,
		Rating:    rating,
		Comment:   comment,
		Hidden:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Hide removes the review from public display while retaining it in storage,
// preserving an audit trail for staff/admin moderation.
func (r *Review) Hide() error {
	if r.Hidden {
		return ErrReviewAlreadyHidden
	}
	r.Hidden = true
	r.UpdatedAt = time.Now()
	return nil
}

// Unhide restores a previously hidden review to public display.
func (r *Review) Unhide() error {
	if !r.Hidden {
		return ErrReviewNotHidden
	}
	r.Hidden = false
	r.UpdatedAt = time.Now()
	return nil
}

// IsVisible reports whether the review should appear in public reads.
func (r *Review) IsVisible() bool {
	return !r.Hidden
}
