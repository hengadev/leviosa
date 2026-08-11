package review

import (
	"bytes"
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

// make test-func TEST_NAME=TestCreateReview TEST_PATH=test/integration/booking/review/create_review_test.go

func TestCreateReview(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	postReview := func(t *testing.T, bookingID uuid.UUID, token string, rating int, comment string) *http.Response {
		t.Helper()
		body, err := json.Marshal(domain.CreateReviewRequest{Rating: rating, Comment: comment})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/review", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: token})
		}

		resp, err := client.Do(req)
		require.NoError(t, err)
		return resp
	}

	t.Run("client can review their own completed booking", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		fixture := newCompletedBooking(t, ctx, client)

		resp := postReview(t, fixture.BookingID, fixture.ClientToken, 5, "Excellent session")
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var review domain.ReviewResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&review))

		assert.Equal(t, fixture.BookingID, review.BookingID)
		assert.Equal(t, fixture.PartnerID, review.PartnerID)
		assert.Equal(t, fixture.ProductID, review.ProductID)
		assert.Equal(t, 5, review.Rating)
		assert.Equal(t, "Excellent session", review.Comment)
		assert.False(t, review.Hidden)
	})

	t.Run("rejects review for a booking that is not completed", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		fixture := newBooking(t, ctx, client) // left `confirmed`, not completed

		resp := postReview(t, fixture.BookingID, fixture.ClientToken, 4, "")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("rejects review from someone other than the booking's client", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		fixture := newCompletedBooking(t, ctx, client)

		// The partner is authenticated but is not the booking's client.
		resp := postReview(t, fixture.BookingID, fixture.PartnerToken, 4, "")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("rejects a second review for the same booking", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		fixture := newCompletedBooking(t, ctx, client)

		first := postReview(t, fixture.BookingID, fixture.ClientToken, 5, "First")
		first.Body.Close()
		require.Equal(t, http.StatusCreated, first.StatusCode)

		second := postReview(t, fixture.BookingID, fixture.ClientToken, 3, "Second attempt")
		defer second.Body.Close()

		assert.Equal(t, http.StatusConflict, second.StatusCode)
	})

	t.Run("rejects an out-of-range rating", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		fixture := newCompletedBooking(t, ctx, client)

		resp := postReview(t, fixture.BookingID, fixture.ClientToken, 6, "")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 404 for a non-existent booking", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		fixture := newCompletedBooking(t, ctx, client)

		resp := postReview(t, uuid.New(), fixture.ClientToken, 5, "")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("requires authentication", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		fixture := newCompletedBooking(t, ctx, client)

		resp := postReview(t, fixture.BookingID, "", 5, "")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
