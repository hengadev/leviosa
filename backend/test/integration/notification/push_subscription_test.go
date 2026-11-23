package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Practitioner notifications (docs/issues/044-iphone-notifications-new-booking.md):
// the Cabinet switch subscribes and unsubscribes the current device through
// the real handler, auth middleware, encx and Postgres. Endpoints and keys
// are encrypted at rest; a user may have several devices.
//
// make test-func TEST_NAME=TestPushSubscriptionEndpoints TEST_PATH=test/integration/notification/push_subscription_test.go

const (
	iphoneEndpoint = "https://web.push.apple.com/QGuQyavXutnMH-iphone-capability"
	ipadEndpoint   = "https://web.push.apple.com/QGuQyavXutnMH-ipad-capability"
)

func pushRequest(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, testServerURL+path, reader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		for key, values := range tu.CreateAuthHeader(token) {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(token))
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func deviceBody(endpoint, userAgent string) map[string]any {
	return map[string]any{
		"endpoint":       endpoint,
		"expirationTime": nil,
		"keys":           map[string]string{"p256dh": "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtUbVlUls0VJXg7A8u-Ts1XbjhazAkj7I99e8QcYP7DkM", "auth": "tBHItJI5svbpez7KI4CCXg"},
		"user_agent":     userAgent,
	}
}

type storedSubscription struct {
	userID    uuid.UUID
	hash      string
	endpoint  []byte
	p256dh    []byte
	auth      []byte
	userAgent string
}

func storedSubscriptions(t *testing.T) []storedSubscription {
	t.Helper()
	rows, err := authCtx.Pool.Query(context.Background(), `
		SELECT user_id, endpoint_hash, endpoint_encrypted, p256dh_encrypted, auth_encrypted, user_agent
		FROM notification.push_subscriptions ORDER BY created_at, user_agent`)
	require.NoError(t, err)
	defer rows.Close()
	var out []storedSubscription
	for rows.Next() {
		var s storedSubscription
		require.NoError(t, rows.Scan(&s.userID, &s.hash, &s.endpoint, &s.p256dh, &s.auth, &s.userAgent))
		out = append(out, s)
	}
	require.NoError(t, rows.Err())
	return out
}

func TestPushSubscriptionEndpoints(t *testing.T) {
	ctx := context.Background()

	t.Run("should refuse anonymous requests and clients", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		client := tu.SetupStandardUser(t, ctx, authCtx)

		assert.Equal(t, http.StatusUnauthorized, pushRequest(t, http.MethodGet, "/notifications/push", "", nil).StatusCode)
		assert.Equal(t, http.StatusUnauthorized, pushRequest(t, http.MethodPut, "/notifications/push/subscription", "", deviceBody(iphoneEndpoint, "iPhone")).StatusCode)
		assert.Equal(t, http.StatusForbidden, pushRequest(t, http.MethodGet, "/notifications/push", client, nil).StatusCode)
		assert.Equal(t, http.StatusForbidden, pushRequest(t, http.MethodPut, "/notifications/push/subscription", client, deviceBody(iphoneEndpoint, "iPhone")).StatusCode)
		assert.Empty(t, storedSubscriptions(t))
	})

	t.Run("should say notifications are available, with the public key", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		resp := pushRequest(t, http.MethodGet, "/notifications/push", tu.SetupAdminUser(t, ctx, authCtx), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var status struct {
			Enabled   bool   `json:"enabled"`
			PublicKey string `json:"public_key"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&status))
		assert.True(t, status.Enabled)
		assert.Equal(t, testVAPIDPublicKey, status.PublicKey)
	})

	t.Run("should store several devices encrypted and unsubscribe one", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		admin := tu.SetupAdminUser(t, ctx, authCtx)

		require.Equal(t, http.StatusNoContent, pushRequest(t, http.MethodPut, "/notifications/push/subscription", admin, deviceBody(iphoneEndpoint, "iPhone")).StatusCode)
		require.Equal(t, http.StatusNoContent, pushRequest(t, http.MethodPut, "/notifications/push/subscription", admin, deviceBody(ipadEndpoint, "iPad")).StatusCode)
		// Switching on again on the same device keeps one row for it.
		require.Equal(t, http.StatusNoContent, pushRequest(t, http.MethodPut, "/notifications/push/subscription", admin, deviceBody(iphoneEndpoint, "iPhone")).StatusCode)

		stored := storedSubscriptions(t)
		require.Len(t, stored, 2, "one row per device")
		assert.Equal(t, stored[0].userID, stored[1].userID)
		for _, s := range stored {
			assert.Len(t, s.hash, 64)
			for _, plain := range []string{"web.push.apple.com", "capability", "tBHItJI5svbpez7KI4CCXg", "BNcRdreALRFX"} {
				for _, column := range [][]byte{s.endpoint, s.p256dh, s.auth} {
					assert.False(t, bytes.Contains(column, []byte(plain)), "%q stored in clear", plain)
				}
			}
		}

		// Both devices get a notification.
		pushSender.reset()
		require.NoError(t, pushService.NotifyUser(ctx, stored[0].userID, notificationDomain.PushMessage{Title: "Nouvelle réservation"}))
		assert.ElementsMatch(t, []string{iphoneEndpoint, ipadEndpoint}, pushSender.sentTo())

		// Unsubscribing the iPad leaves the iPhone.
		require.Equal(t, http.StatusNoContent, pushRequest(t, http.MethodDelete, "/notifications/push/subscription", admin, map[string]string{"endpoint": ipadEndpoint}).StatusCode)
		stored = storedSubscriptions(t)
		require.Len(t, stored, 1)
		assert.Equal(t, "iPhone", stored[0].userAgent)

		// Unsubscribing a device that is not stored is not an error.
		assert.Equal(t, http.StatusNoContent, pushRequest(t, http.MethodDelete, "/notifications/push/subscription", admin, map[string]string{"endpoint": ipadEndpoint}).StatusCode)
	})

	t.Run("should not let another practitioner unsubscribe a device", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		admin := tu.SetupAdminUser(t, ctx, authCtx)
		partner := tu.SetupPartnerUser(t, ctx, authCtx)

		require.Equal(t, http.StatusNoContent, pushRequest(t, http.MethodPut, "/notifications/push/subscription", admin, deviceBody(iphoneEndpoint, "iPhone")).StatusCode)
		require.Equal(t, http.StatusNoContent, pushRequest(t, http.MethodDelete, "/notifications/push/subscription", partner, map[string]string{"endpoint": iphoneEndpoint}).StatusCode)
		assert.Len(t, storedSubscriptions(t), 1)
	})

	t.Run("should delete a subscription the push service reports gone", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		admin := tu.SetupAdminUser(t, ctx, authCtx)
		require.Equal(t, http.StatusNoContent, pushRequest(t, http.MethodPut, "/notifications/push/subscription", admin, deviceBody(iphoneEndpoint, "iPhone")).StatusCode)
		require.Equal(t, http.StatusNoContent, pushRequest(t, http.MethodPut, "/notifications/push/subscription", admin, deviceBody(ipadEndpoint, "iPad")).StatusCode)
		userID := storedSubscriptions(t)[0].userID

		pushSender.reset(ipadEndpoint)
		require.NoError(t, pushService.NotifyUser(ctx, userID, notificationDomain.PushMessage{Title: "Nouvelle réservation"}))

		stored := storedSubscriptions(t)
		require.Len(t, stored, 1)
		assert.Equal(t, "iPhone", stored[0].userAgent)
	})

	t.Run("should refuse a malformed subscription", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		admin := tu.SetupAdminUser(t, ctx, authCtx)
		for _, body := range []any{
			deviceBody("http://insecure.example/push", "iPhone"),
			map[string]any{"endpoint": iphoneEndpoint},
			map[string]any{"endpoint": iphoneEndpoint, "keys": map[string]string{"p256dh": "k", "auth": "a"}, "extra": true},
		} {
			resp := pushRequest(t, http.MethodPut, "/notifications/push/subscription", admin, body)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "%v", body)
		}
		assert.Empty(t, storedSubscriptions(t))
	})
}
