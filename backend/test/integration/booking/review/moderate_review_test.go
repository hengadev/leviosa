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

// make test-func TEST_NAME=TestModerateReview TEST_PATH=test/integration/booking/review/moderate_review_test.go

func createReviewViaAPI(t *testing.T, ctx context.Context, client *http.Client) (*bookingFixture, domain.ReviewResponse) {
	t.Helper()

	fixture := newCompletedBooking(t, ctx, client)

	body, err := json.Marshal(domain.CreateReviewRequest{Rating: 4, Comment: "Solid"})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", testServerURL+"/bookings/"+fixture.BookingID.String()+"/review", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: fixture.ClientToken})

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var review domain.ReviewResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&review))

	return fixture, review
}

func moderateRequest(t *testing.T, client *http.Client, action string, reviewID uuid.UUID, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest("POST", testServerURL+"/reviews/"+reviewID.String()+"/"+action, nil)
	require.NoError(t, err)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: token})
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func TestModerateReview(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("admin can hide a review, removing it from public reads", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		fixture, review := createReviewViaAPI(t, ctx, client)
		adminToken := newAdminToken(t, ctx)

		hideResp := moderateRequest(t, client, "hide", review.ID, adminToken)
		defer hideResp.Body.Close()
		require.Equal(t, http.StatusOK, hideResp.StatusCode)

		var hidden domain.ReviewResponse
		require.NoError(t, json.NewDecoder(hideResp.Body).Decode(&hidden))
		assert.True(t, hidden.Hidden)

		// Public read for the partner should now exclude it.
		listResp, err := client.Get(testServerURL + "/partners/reviews/" + fixture.PartnerID.String())
		require.NoError(t, err)
		defer listResp.Body.Close()
		var list domain.PublicReviewListResponse
		require.NoError(t, json.NewDecoder(listResp.Body).Decode(&list))
		assert.Equal(t, 0, list.ReviewCount)
		assert.Empty(t, list.Reviews)
	})

	t.Run("admin can unhide a previously hidden review", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		fixture, review := createReviewViaAPI(t, ctx, client)
		adminToken := newAdminToken(t, ctx)

		hideResp := moderateRequest(t, client, "hide", review.ID, adminToken)
		hideResp.Body.Close()
		require.Equal(t, http.StatusOK, hideResp.StatusCode)

		unhideResp := moderateRequest(t, client, "unhide", review.ID, adminToken)
		defer unhideResp.Body.Close()
		require.Equal(t, http.StatusOK, unhideResp.StatusCode)

		var unhidden domain.ReviewResponse
		require.NoError(t, json.NewDecoder(unhideResp.Body).Decode(&unhidden))
		assert.False(t, unhidden.Hidden)

		listResp, err := client.Get(testServerURL + "/partners/reviews/" + fixture.PartnerID.String())
		require.NoError(t, err)
		defer listResp.Body.Close()
		var list domain.PublicReviewListResponse
		require.NoError(t, json.NewDecoder(listResp.Body).Decode(&list))
		assert.Equal(t, 1, list.ReviewCount)
	})

	t.Run("hiding an already-hidden review returns 422", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		_, review := createReviewViaAPI(t, ctx, client)
		adminToken := newAdminToken(t, ctx)

		first := moderateRequest(t, client, "hide", review.ID, adminToken)
		first.Body.Close()
		require.Equal(t, http.StatusOK, first.StatusCode)

		second := moderateRequest(t, client, "hide", review.ID, adminToken)
		defer second.Body.Close()
		assert.Equal(t, http.StatusUnprocessableEntity, second.StatusCode)
	})

	t.Run("non-admin cannot hide a review", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		fixture, review := createReviewViaAPI(t, ctx, client)

		// The booking's own client (a Standard user) is not staff/admin.
		resp := moderateRequest(t, client, "hide", review.ID, fixture.ClientToken)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("hiding a non-existent review returns 404", func(t *testing.T) {
		clearReviewTestTables(t, ctx)
		adminToken := newAdminToken(t, ctx)

		resp := moderateRequest(t, client, "hide", uuid.New(), adminToken)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
