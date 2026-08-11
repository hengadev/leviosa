package helpers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	th "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/stretchr/testify/assert"
)

// make test-func TEST_NAME=TestGetCompanyHostingProvider TEST_PATH=test/integration/settings/company_hosting_provider_test.go

func TestGetCompanyHostingProvider(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should return 404 when hosting provider not set", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		req := th.NewGetCompanyHostingProviderRequest(t, ctx, testServerURL)
		authHeader := tu.CreateAuthHeader(accessToken)
		for key, values := range authHeader {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("should successfully retrieve hosting provider identity", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		hostingProvider := "OVH SAS, 2 rue Kellermann, 59100 Roubaix, France"
		th.InsertCompanyHostingProvider(t, ctx, hostingProvider, testPool)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		req := th.NewGetCompanyHostingProviderRequest(t, ctx, testServerURL)
		authHeader := tu.CreateAuthHeader(accessToken)
		for key, values := range authHeader {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respBody domain.GetCompanyHostingProviderResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.Equal(t, hostingProvider, respBody.HostingProvider)
	})
}

// make test-func TEST_NAME=TestSetCompanyHostingProvider TEST_PATH=test/integration/settings/company_hosting_provider_test.go

func TestSetCompanyHostingProvider(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should successfully set hosting provider identity", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		hostingProvider := "OVH SAS, 2 rue Kellermann, 59100 Roubaix, France"
		request := domain.SetCompanyHostingProviderRequest{HostingProvider: hostingProvider}
		req := th.NewSetCompanyHostingProviderRequest(t, ctx, testServerURL, request)

		authHeader := tu.CreateAuthHeader(accessToken)
		for key, values := range authHeader {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respBody domain.SetCompanyHostingProviderResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.True(t, respBody.Success)

		retrieved, err := th.GetCompanyHostingProvider(t, ctx, testPool)
		assert.NoError(t, err)
		assert.Equal(t, hostingProvider, retrieved)
	})

	t.Run("should return 400 for empty hosting provider identity", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		request := domain.SetCompanyHostingProviderRequest{HostingProvider: ""}
		req := th.NewSetCompanyHostingProviderRequest(t, ctx, testServerURL, request)

		authHeader := tu.CreateAuthHeader(accessToken)
		for key, values := range authHeader {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var respBody struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.Contains(t, respBody.Error, "hosting_provider_required")
	})

	t.Run("should return 400 for whitespace-only hosting provider identity", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		request := domain.SetCompanyHostingProviderRequest{HostingProvider: "   \n\t   "}
		req := th.NewSetCompanyHostingProviderRequest(t, ctx, testServerURL, request)

		authHeader := tu.CreateAuthHeader(accessToken)
		for key, values := range authHeader {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var respBody struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.Contains(t, respBody.Error, "hosting_provider_empty")
	})

	t.Run("should return 400 for hosting provider identity exceeding 500 characters", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		request := domain.SetCompanyHostingProviderRequest{HostingProvider: strings.Repeat("A", 501)}
		req := th.NewSetCompanyHostingProviderRequest(t, ctx, testServerURL, request)

		authHeader := tu.CreateAuthHeader(accessToken)
		for key, values := range authHeader {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var respBody struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.Contains(t, respBody.Error, "hosting_provider_length")
	})

	t.Run("should return 401 when not authenticated as admin", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)

		request := domain.SetCompanyHostingProviderRequest{HostingProvider: "OVH SAS, Roubaix, France"}
		req := th.NewSetCompanyHostingProviderRequest(t, ctx, testServerURL, request)

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("should successfully update existing hosting provider identity", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		th.InsertCompanyHostingProvider(t, ctx, "Old Host, Old Address", testPool)

		newHostingProvider := "New Host SAS, New Address, France"
		request := domain.SetCompanyHostingProviderRequest{HostingProvider: newHostingProvider}
		req := th.NewSetCompanyHostingProviderRequest(t, ctx, testServerURL, request)

		authHeader := tu.CreateAuthHeader(accessToken)
		for key, values := range authHeader {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		retrieved, err := th.GetCompanyHostingProvider(t, ctx, testPool)
		assert.NoError(t, err)
		assert.Equal(t, newHostingProvider, retrieved)
	})
}
