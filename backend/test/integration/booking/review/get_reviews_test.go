package review

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestGetPartnerReviews TEST_PATH=test/integration/booking/review/get_reviews_test.go

func createReviewDirect(t *testing.T, ctx context.Context, bookingID, clientID, partnerID, productID uuid.UUID, rating int, comment string, hidden bool) {
	t.Helper()
	review, err := domain.NewReview(bookingID, clientID, partnerID, productID, rating, comment)
	require.NoError(t, err)
	if hidden {
		require.NoError(t, review.Hide())
	}
	require.NoError(t, reviewRepo.Create(ctx, review))
}

func TestGetPartnerReviews(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	getPartnerReviews := func(t *testing.T, partnerID uuid.UUID) *domain.PublicReviewListResponse {
		t.Helper()
		resp, err := client.Get(testServerURL + "/partners/reviews/" + partnerID.String())
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body domain.PublicReviewListResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		return &body
	}

	t.Run("returns empty aggregate for a partner with no reviews", func(t *testing.T) {
		clearReviewTestTables(t, ctx)

		body := getPartnerReviews(t, uuid.New())

		assert.Equal(t, 0, body.ReviewCount)
		assert.Equal(t, float64(0), body.AggregateRating)
		assert.Empty(t, body.Reviews)
	})

	t.Run("returns non-hidden reviews and aggregate rating, excluding hidden ones", func(t *testing.T) {
		clearReviewTestTables(t, ctx)

		// booking_id has a foreign key to booking.bookings, so each review needs a
		// real, distinct booking — create three completed bookings against the same
		// partner.
		partner := newPartner(t, ctx)
		bookingA := newBookingForPartner(t, ctx, client, partner)
		completeBooking(t, client, bookingA.BookingID, partner.PartnerToken)
		bookingB := newBookingForPartner(t, ctx, client, partner)
		completeBooking(t, client, bookingB.BookingID, partner.PartnerToken)
		bookingC := newBookingForPartner(t, ctx, client, partner)
		completeBooking(t, client, bookingC.BookingID, partner.PartnerToken)

		createReviewDirect(t, ctx, bookingA.BookingID, bookingA.ClientID, partner.PartnerID, partner.ProductID, 5, "Loved it", false)
		createReviewDirect(t, ctx, bookingB.BookingID, bookingB.ClientID, partner.PartnerID, partner.ProductID, 3, "It was okay", false)
		createReviewDirect(t, ctx, bookingC.BookingID, bookingC.ClientID, partner.PartnerID, partner.ProductID, 1, "Hidden abusive review", true)

		body := getPartnerReviews(t, partner.PartnerID)

		assert.Equal(t, 2, body.ReviewCount)
		assert.InDelta(t, 4.0, body.AggregateRating, 0.001) // (5+3)/2
		assert.Len(t, body.Reviews, 2)
		for _, r := range body.Reviews {
			assert.False(t, r.Hidden)
		}
	})

	t.Run("does not mix reviews across partners", func(t *testing.T) {
		clearReviewTestTables(t, ctx)

		fixtureA := newCompletedBooking(t, ctx, client)
		createReviewDirect(t, ctx, fixtureA.BookingID, fixtureA.ClientID, fixtureA.PartnerID, fixtureA.ProductID, 5, "", false)

		otherPartnerID := uuid.New()
		body := getPartnerReviews(t, otherPartnerID)

		assert.Equal(t, 0, body.ReviewCount)
		assert.Empty(t, body.Reviews)
	})
}

func TestGetProductReviews(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("returns non-hidden reviews and aggregate rating for a product", func(t *testing.T) {
		clearReviewTestTables(t, ctx)

		fixture := newCompletedBooking(t, ctx, client)
		createReviewDirect(t, ctx, fixture.BookingID, fixture.ClientID, fixture.PartnerID, fixture.ProductID, 4, "Good", false)

		resp, err := client.Get(testServerURL + "/products/reviews/" + fixture.ProductID.String())
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body domain.PublicReviewListResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

		assert.Equal(t, 1, body.ReviewCount)
		assert.InDelta(t, 4.0, body.AggregateRating, 0.001)
		require.Len(t, body.Reviews, 1)
		assert.Equal(t, "Good", body.Reviews[0].Comment)
	})
}
