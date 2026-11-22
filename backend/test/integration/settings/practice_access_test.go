package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

// Practice access: the door code and access instructions of the practice
// (docs/issues/023-practice-access-settings.md). Private: encrypted at rest,
// admin-only read and write.
//
// make test-func TEST_NAME=TestPracticeAccessSettings TEST_PATH=test/integration/settings/practice_access_test.go

func TestPracticeAccessSettings(t *testing.T) {
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

	getAccess := func(t *testing.T, accessToken string) domain.GetPracticeAccessResponse {
		resp, err := client.Do(withAuth(th.NewGetPracticeAccessRequest(t, ctx, testServerURL), accessToken))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var body domain.GetPracticeAccessResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		return body
	}

	setAccess := func(t *testing.T, accessToken string, request domain.SetPracticeAccessRequest) (int, domain.SetPracticeAccessResponse, string) {
		resp, err := client.Do(withAuth(th.NewSetPracticeAccessRequest(t, ctx, testServerURL, request), accessToken))
		require.NoError(t, err)
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		var body domain.SetPracticeAccessResponse
		if resp.StatusCode == http.StatusOK {
			require.NoError(t, json.Unmarshal(raw, &body))
		}
		return resp.StatusCode, body, string(raw)
	}

	t.Run("should answer empty values to an administrator while unset", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		access := getAccess(t, tu.SetupAdminUser(t, ctx, authCtx))
		assert.Empty(t, access.DoorCode)
		assert.Empty(t, access.AccessInstructions)
	})

	t.Run("should store encrypted values, serve them back and clear them", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		th.ClearAuditEntriesTable(t, ctx, testPool)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		status, body, _ := setAccess(t, accessToken, domain.SetPracticeAccessRequest{
			DoorCode:           " 4521B ",
			AccessInstructions: "Bâtiment B, 2e étage.\nSonner à Leviosa.",
		})
		require.Equal(t, http.StatusOK, status)
		assert.True(t, body.CodeChanged)

		for key, plain := range map[string]string{
			settings.PracticeAccessDoorCode:     "4521B",
			settings.PracticeAccessInstructions: "Bâtiment B",
		} {
			stored := th.GetEncryptedSettingByKey(t, ctx, key, testPool)
			assert.NotEmpty(t, stored.ValueEncrypted, key)
			assert.False(t, bytes.Contains(stored.ValueEncrypted, []byte(plain)), "%s holds the plain text", key)

			entry, ok := th.GetLatestAuditEntry(t, ctx, testPool, "setting", key)
			require.True(t, ok, "expected an audit entry for %s", key)
			assert.Equal(t, "settings.field_updated", entry.Action)
			assert.Equal(t, "[redacted]", entry.Payload["new_value"])
		}

		access := getAccess(t, accessToken)
		assert.Equal(t, "4521B", access.DoorCode)
		assert.Equal(t, "Bâtiment B, 2e étage.\nSonner à Leviosa.", access.AccessInstructions)

		status, body, _ = setAccess(t, accessToken, domain.SetPracticeAccessRequest{
			DoorCode:           "4521B",
			AccessInstructions: "Autre porte",
		})
		require.Equal(t, http.StatusOK, status)
		assert.False(t, body.CodeChanged, "the same door code is not a change")

		status, body, _ = setAccess(t, accessToken, domain.SetPracticeAccessRequest{})
		require.Equal(t, http.StatusOK, status)
		assert.True(t, body.CodeChanged)

		access = getAccess(t, accessToken)
		assert.Empty(t, access.DoorCode)
		assert.Empty(t, access.AccessInstructions)
	})

	t.Run("should refuse a door code over the limit without echoing it", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		longCode := "SECRET-" + string(bytes.Repeat([]byte("9"), domain.MaxPracticeAccessDoorCodeLength))

		status, _, raw := setAccess(t, accessToken, domain.SetPracticeAccessRequest{DoorCode: longCode})
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, raw, "door code")
		assert.NotContains(t, raw, "SECRET")
	})

	t.Run("should refuse anonymous and non-administrator requests", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		adminToken := tu.SetupAdminUser(t, ctx, authCtx)
		status, _, _ := setAccess(t, adminToken, domain.SetPracticeAccessRequest{DoorCode: "4521B"})
		require.Equal(t, http.StatusOK, status)

		standardToken := tu.SetupStandardUser(t, ctx, authCtx)
		for name, token := range map[string]string{"anonymous": "", "standard user": standardToken} {
			t.Run(name, func(t *testing.T) {
				get := th.NewGetPracticeAccessRequest(t, ctx, testServerURL)
				set := th.NewSetPracticeAccessRequest(t, ctx, testServerURL, domain.SetPracticeAccessRequest{DoorCode: "0000"})
				if token != "" {
					get, set = withAuth(get, token), withAuth(set, token)
				}
				for _, req := range []*http.Request{get, set} {
					resp, err := client.Do(req)
					require.NoError(t, err)
					raw, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, resp.StatusCode, req.Method)
					assert.NotContains(t, string(raw), "4521B")
				}
			})
		}

		assert.Equal(t, "4521B", getAccess(t, adminToken).DoorCode, "the refused write changed nothing")
	})

	t.Run("should serve the keys on no public, internal or bulk route", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		adminToken := tu.SetupAdminUser(t, ctx, authCtx)
		status, _, _ := setAccess(t, adminToken, domain.SetPracticeAccessRequest{DoorCode: "4521B", AccessInstructions: "Bâtiment B"})
		require.Equal(t, http.StatusOK, status)

		requests := []*http.Request{}
		for _, path := range []string{"/settings/practice-access", "/internal/settings/practice-access"} {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServerURL+path, nil)
			require.NoError(t, err)
			requests = append(requests, req)
		}
		requests = append(requests, withAuth(th.NewBulkSettingsRequest(t, ctx, testServerURL,
			[]string{settings.PracticeAccessDoorCode, settings.PracticeAccessInstructions}), adminToken))

		for _, req := range requests {
			resp, err := client.Do(req)
			require.NoError(t, err)
			raw, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			assert.NotEqual(t, http.StatusOK, resp.StatusCode, req.URL.Path)
			assert.NotContains(t, string(raw), "4521B", req.URL.Path)
			assert.NotContains(t, string(raw), "Bâtiment", req.URL.Path)
		}
	})
}
