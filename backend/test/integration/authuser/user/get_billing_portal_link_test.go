package user_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	td "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestGetBillingPortalLink TEST_PATH=test/integration/authuser/user/get_billing_portal_link_test.go

// createAuthenticatedTestUser inserts a test user (optionally with a Stripe customer ID) and
// returns an access token for an active session belonging to that user.
func createAuthenticatedTestUser(t *testing.T, ctx context.Context, stripeCustomerID string) string {
	t.Helper()

	testUser := td.NewTestUser(t, "billingportal-"+uuid.New().String()+"@example.com", "Billing", "Portal")
	testUser.State = domain.Active
	testUser.StripeCustomerID = stripeCustomerID

	testUserEncx, err := domain.ProcessUserEncx(ctx, crypto, testUser)
	require.NoError(t, err)
	err = td.InsertUserEncx(t, ctx, testUserEncx, testPool)
	require.NoError(t, err)

	now := time.Now()
	accessToken, err := session.GenerateToken()
	require.NoError(t, err)
	refreshToken, err := session.GenerateToken()
	require.NoError(t, err)

	standardSession := &session.Session{
		ID:           uuid.New(),
		UserID:       testUser.ID,
		Role:         identity.Standard,
		State:        session.SessionActive,
		CreatedAt:    now,
		ExpiresAt:    now.Add(24 * time.Hour),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	standardSessionEncx, err := session.ProcessSessionEncx(ctx, crypto, standardSession)
	require.NoError(t, err)
	td.InsertSessionEncx(t, ctx, redisClient, standardSessionEncx, time.Hour)

	return accessToken
}

func TestGetBillingPortalLink(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	returnURL := "https://example.com/client/profile?stripe=return"

	t.Run("should return 401 without authentication", func(t *testing.T) {
		req := td.NewGetBillingPortalLinkRequest(t, ctx, testServerURL, returnURL, "")
		resp, err := client.Do(req)

		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should return a billing portal session URL for an authenticated user with a Stripe customer", func(t *testing.T) {
		td.ClearUsersTable(t, ctx, testPool)
		td.ClearSessionsRedis(t, ctx, redisClient)

		// The stripe-mock testcontainer generates fixture responses for any documented Stripe
		// endpoint, including /v1/billing_portal/sessions, regardless of whether the given
		// customer ID "really" exists — mirroring how other Stripe integration tests in this
		// repo (e.g. catalog price creation) verify against stripe-mock rather than live Stripe.
		accessToken := createAuthenticatedTestUser(t, ctx, "cus_billing_portal_test")

		req := td.NewGetBillingPortalLinkRequest(t, ctx, testServerURL, returnURL, accessToken)
		resp, err := client.Do(req)

		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		url := td.ParseGetBillingPortalLinkResponse(t, resp)
		assert.NotEmpty(t, url, "expected a non-empty billing portal session URL")
	})

	t.Run("should return an error when the user has no Stripe customer", func(t *testing.T) {
		td.ClearUsersTable(t, ctx, testPool)
		td.ClearSessionsRedis(t, ctx, redisClient)

		accessToken := createAuthenticatedTestUser(t, ctx, "")

		req := td.NewGetBillingPortalLinkRequest(t, ctx, testServerURL, returnURL, accessToken)
		resp, err := client.Do(req)

		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("should return 400 when return_url is missing", func(t *testing.T) {
		td.ClearUsersTable(t, ctx, testPool)
		td.ClearSessionsRedis(t, ctx, redisClient)

		accessToken := createAuthenticatedTestUser(t, ctx, "cus_billing_portal_test")

		req := td.NewGetBillingPortalLinkRequest(t, ctx, testServerURL, "", accessToken)
		resp, err := client.Do(req)

		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
