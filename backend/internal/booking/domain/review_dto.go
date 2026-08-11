package domain

import (
	"time"

	"github.com/google/uuid"
)

// CreateReviewRequest is the client-submitted payload for POST /bookings/{id}/review.
type CreateReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment,omitempty"`
}

// ReviewResponse is the API representation of a single review.
type ReviewResponse struct {
	ID        uuid.UUID `json:"id"`
	BookingID uuid.UUID `json:"booking_id"`
	PartnerID uuid.UUID `json:"partner_id"`
	ProductID uuid.UUID `json:"product_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment,omitempty"`
	Hidden    bool      `json:"hidden"`
	CreatedAt time.Time `json:"created_at"`
}

// ReviewResponseFromReview converts a domain Review into its API DTO.
func ReviewResponseFromReview(r *Review) ReviewResponse {
	return ReviewResponse{
		ID:        r.ID,
		BookingID: r.BookingID,
		PartnerID: r.PartnerID,
		ProductID: r.ProductID,
		Rating:    r.Rating,
		Comment:   r.Comment,
		Hidden:    r.Hidden,
		CreatedAt: r.CreatedAt,
	}
}

// PublicReviewListResponse is returned by the public read endpoints: a page of
// non-hidden reviews plus an aggregate rating. The aggregate is computed on read
// from the full non-hidden set (not just the returned page) rather than maintained
// as a running counter — see docs/prd/public-trust-and-seo-foundations.md.
type PublicReviewListResponse struct {
	AggregateRating float64          `json:"aggregate_rating"`
	ReviewCount     int              `json:"review_count"`
	Reviews         []ReviewResponse `json:"reviews"`
}
