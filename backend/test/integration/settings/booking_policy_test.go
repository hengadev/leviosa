package helpers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	th "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Booking policy settings: minimum notice and booking horizon applied by the
// slot engine (docs/issues/010-slot-engine.md).
//
// make test-func TEST_NAME=TestBookingPolicySettings TEST_PATH=test/integration/settings/booking_policy_test.go

func TestBookingPolicySettings(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	withAuth := func(req *http.Request, accessToken string) *http.Request {
		for key, values := range tu.CreateAuthHeader(accessToken) {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))
		return req
	}

	getPolicy := func(t *testing.T) domain.GetBookingPolicyResponse {
		resp, err := client.Do(th.NewGetBookingPolicyRequest(t, ctx, testServerURL))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var body domain.GetBookingPolicyResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		return body
	}

	t.Run("should answer the defaults without authentication while unset", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)

		policy := getPolicy(t)
		assert.Equal(t, 12, policy.MinimumNoticeHours)
		assert.Equal(t, 8, policy.HorizonWeeks)
	})

	t.Run("should persist an administrator's values and serve them at once", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		req := withAuth(th.NewSetBookingPolicyRequest(t, ctx, testServerURL,
			domain.SetBookingPolicyRequest{MinimumNoticeHours: 2, HorizonWeeks: 4}), accessToken)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		assert.Equal(t, 2, th.GetSettingIntByKey(t, ctx, settings.BookingMinimumNoticeHours, testPool).Value)
		assert.Equal(t, 4, th.GetSettingIntByKey(t, ctx, settings.BookingHorizonWeeks, testPool).Value)

		policy := getPolicy(t)
		assert.Equal(t, 2, policy.MinimumNoticeHours)
		assert.Equal(t, 4, policy.HorizonWeeks)
	})

	t.Run("should return 400 for an out-of-range horizon", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		req := withAuth(th.NewSetBookingPolicyRequest(t, ctx, testServerURL,
			domain.SetBookingPolicyRequest{MinimumNoticeHours: 12, HorizonWeeks: 0}), accessToken)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should reject a write from a non-administrator", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupStandardUser(t, ctx, authCtx)
		req := withAuth(th.NewSetBookingPolicyRequest(t, ctx, testServerURL,
			domain.SetBookingPolicyRequest{MinimumNoticeHours: 0, HorizonWeeks: 52}), accessToken)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)

		policy := getPolicy(t)
		assert.Equal(t, 12, policy.MinimumNoticeHours, "the defaults still apply")
	})
}
