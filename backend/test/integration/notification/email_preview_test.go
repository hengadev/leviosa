package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Admin email previews (docs/issues/027-email-preview-catalogue-and-admin-endpoints.md):
// administrator-only, rendered with sample data; the contact form stays
// public.
//
// make test-func TEST_NAME=TestEmailPreviewEndpoints TEST_PATH=test/integration/notification/email_preview_test.go

type previewSummary struct {
	ID       string            `json:"id"`
	Label    string            `json:"label"`
	Audience string            `json:"audience"`
	Locales  []string          `json:"locales"`
	Subjects map[string]string `json:"subjects"`
}

type previewBody struct {
	Subject     string   `json:"subject"`
	HTML        string   `json:"html"`
	Attachments []string `json:"attachments"`
}

func TestEmailPreviewEndpoints(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	get := func(t *testing.T, path, accessToken string) *http.Response {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServerURL+path, nil)
		require.NoError(t, err)
		if accessToken != "" {
			for key, values := range tu.CreateAuthHeader(accessToken) {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
			req.AddCookie(tu.CreateAuthCookie(accessToken))
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { resp.Body.Close() })
		return resp
	}

	paths := []string{
		"/admin/emails/previews",
		"/admin/emails/previews/booking_reminder?locale=fr",
	}

	t.Run("should refuse anonymous requests", func(t *testing.T) {
		for _, path := range paths {
			resp := get(t, path, "")
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, path)
		}
	})

	t.Run("should refuse a non-administrator", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		token := tu.SetupStandardUser(t, ctx, authCtx)
		for _, path := range paths {
			resp := get(t, path, token)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode, path)
		}
	})

	t.Run("should list every email to an administrator", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		resp := get(t, "/admin/emails/previews", tu.SetupAdminUser(t, ctx, authCtx))
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var list []previewSummary
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
		require.Len(t, list, 11)
		for _, entry := range list {
			assert.NotEmpty(t, entry.ID)
			assert.NotEmpty(t, entry.Label)
			assert.Contains(t, []string{"client", "partner", "practice"}, entry.Audience)
			require.NotEmpty(t, entry.Locales)
			for _, locale := range entry.Locales {
				assert.NotEmpty(t, entry.Subjects[locale], "%s %s subject", entry.ID, locale)
			}
		}
	})

	t.Run("should render a client email in French and English", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		token := tu.SetupAdminUser(t, ctx, authCtx)

		for _, query := range []string{"", "?locale=fr", "?locale=en"} {
			resp := get(t, "/admin/emails/previews/booking_confirmation_full"+query, token)
			require.Equal(t, http.StatusOK, resp.StatusCode, query)

			var body previewBody
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			assert.NotEmpty(t, body.Subject, query)
			assert.Contains(t, body.HTML, "4821B", query)
			require.Len(t, body.Attachments, 1, query)
			assert.True(t, strings.HasSuffix(body.Attachments[0], ".ics"), query)
		}

		fr := get(t, "/admin/emails/previews/booking_reminder", token)
		en := get(t, "/admin/emails/previews/booking_reminder?locale=en", token)
		var frBody, enBody previewBody
		require.NoError(t, json.NewDecoder(fr.Body).Decode(&frBody))
		require.NoError(t, json.NewDecoder(en.Body).Decode(&enBody))
		assert.NotEqual(t, frBody.Subject, enBody.Subject, "the locale defaults to French and en switches language")
	})

	t.Run("should answer 404 for an unknown id or an unsupported locale", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		token := tu.SetupAdminUser(t, ctx, authCtx)

		for _, path := range []string{
			"/admin/emails/previews/no_such_email",
			"/admin/emails/previews/booking_reminder?locale=de",
			"/admin/emails/previews/booking_partner_notice?locale=en",
		} {
			resp := get(t, path, token)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode, path)
		}
	})
}

func TestContactFormStaysPublic(t *testing.T) {
	body := `{"name":"Alice Visiteuse","contact_email":"alice@example.com","message":"Bonjour"}`
	resp, err := http.Post(testServerURL+"/contact", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Less(t, resp.StatusCode, 300)
	assert.Contains(t, notifications.contactNames(), "Alice Visiteuse")
}
