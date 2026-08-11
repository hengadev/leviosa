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

// make test-func TEST_NAME=TestGetCompanyRegistrationNumber TEST_PATH=test/integration/settings/company_registration_number_test.go

func TestGetCompanyRegistrationNumber(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should return 401 when not authenticated as admin", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)

		req := th.NewGetCompanyRegistrationNumberRequest(t, ctx, testServerURL)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("should return 404 when company registration number not set", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		req := th.NewGetCompanyRegistrationNumberRequest(t, ctx, testServerURL)
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

	t.Run("should successfully retrieve company registration number", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		registrationNumber := "12345678901234"
		th.InsertCompanyRegistrationNumber(t, ctx, registrationNumber, testPool)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		req := th.NewGetCompanyRegistrationNumberRequest(t, ctx, testServerURL)
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

		var respBody domain.GetCompanyRegistrationNumberResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.Equal(t, registrationNumber, respBody.RegistrationNumber)
	})
}

// make test-func TEST_NAME=TestSetCompanyRegistrationNumber TEST_PATH=test/integration/settings/company_registration_number_test.go

func TestSetCompanyRegistrationNumber(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should successfully set company registration number (SIRET)", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		registrationNumber := "12345678901234"
		request := domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: registrationNumber}
		req := th.NewSetCompanyRegistrationNumberRequest(t, ctx, testServerURL, request)

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

		var respBody domain.SetCompanyRegistrationNumberResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.True(t, respBody.Success)

		retrieved, err := th.GetCompanyRegistrationNumber(t, ctx, testPool)
		assert.NoError(t, err)
		assert.Equal(t, registrationNumber, retrieved)
	})

	t.Run("should successfully set company registration number (SIREN)", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		registrationNumber := "123456789"
		request := domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: registrationNumber}
		req := th.NewSetCompanyRegistrationNumberRequest(t, ctx, testServerURL, request)

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
	})

	t.Run("should return 400 for empty registration number", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		request := domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: ""}
		req := th.NewSetCompanyRegistrationNumberRequest(t, ctx, testServerURL, request)

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
		assert.Contains(t, respBody.Error, "registration_number_required")
	})

	t.Run("should return 400 for a registration number with non-digit characters", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		request := domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "1234ABCD9012"}
		req := th.NewSetCompanyRegistrationNumberRequest(t, ctx, testServerURL, request)

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
		assert.Contains(t, respBody.Error, "registration_number_format")
	})

	t.Run("should return 400 for a registration number of the wrong length", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		request := domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "12345"}
		req := th.NewSetCompanyRegistrationNumberRequest(t, ctx, testServerURL, request)

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
		assert.Contains(t, respBody.Error, "registration_number_length")
	})

	t.Run("should return 415 for incorrect content type", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		request := domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "123456789"}
		req := th.NewSetCompanyRegistrationNumberRequest(t, ctx, testServerURL, request)
		req.Header.Set("Content-Type", "text/plain")

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

		assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	})

	t.Run("should return 401 when not authenticated as admin", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)

		request := domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "123456789"}
		req := th.NewSetCompanyRegistrationNumberRequest(t, ctx, testServerURL, request)

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("should successfully update existing registration number", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		th.InsertCompanyRegistrationNumber(t, ctx, "111111111", testPool)

		newRegistrationNumber := "222222222"
		request := domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: newRegistrationNumber}
		req := th.NewSetCompanyRegistrationNumberRequest(t, ctx, testServerURL, request)

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

		retrieved, err := th.GetCompanyRegistrationNumber(t, ctx, testPool)
		assert.NoError(t, err)
		assert.Equal(t, newRegistrationNumber, retrieved)
	})

	t.Run("should trim whitespace before persisting", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		request := domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "  123456789  "}
		req := th.NewSetCompanyRegistrationNumberRequest(t, ctx, testServerURL, request)

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

		retrieved, err := th.GetCompanyRegistrationNumber(t, ctx, testPool)
		assert.NoError(t, err)
		assert.False(t, strings.Contains(retrieved, " "))
		assert.Equal(t, "123456789", retrieved)
	})
}
