// Package review implements the review sub-module's use cases (issue 009). It lives
// inside the booking domain — rather than as a standalone top-level domain — because
// it depends entirely on booking-relationship checks and the booking status FSM
// (ADR-0006), which the booking domain already owns.
package review

import (
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
)

type service struct {
	reviewRepo ports.ReviewRepository
	bookingSvc ports.BookingService
}

// New creates a new ReviewService. bookingSvc is used to look up the referenced
// booking's status and ownership (client ID, partner ID, product ID) — reviews are
// denormalized against those fields at creation time so public reads don't need to
// decrypt booking.bookings rows.
func New(reviewRepo ports.ReviewRepository, bookingSvc ports.BookingService) ports.ReviewService {
	return &service{reviewRepo: reviewRepo, bookingSvc: bookingSvc}
}
