package helpers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	th "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/stretchr/testify/assert"
)

// make test-func TEST_NAME=TestGetCompanyVatNumber TEST_PATH=test/integration/settings/company_vat_number_test.go

func TestGetCompanyVatNumber(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should return 404 when company VAT number not set", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		req := th.NewGetCompanyVatNumberRequest(t, ctx, testServerURL)
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

	t.Run("should successfully retrieve company VAT number", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		vatNumber := "FR12345678901"
		th.InsertCompanyVatNumber(t, ctx, vatNumber, testPool)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		req := th.NewGetCompanyVatNumberRequest(t, ctx, testServerURL)
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

		var respBody domain.GetCompanyVatNumberResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.Equal(t, vatNumber, respBody.VatNumber)
	})

	t.Run("should successfully retrieve a blank VAT number for franchise-regime businesses", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		th.InsertCompanyVatNumber(t, ctx, "", testPool)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		req := th.NewGetCompanyVatNumberRequest(t, ctx, testServerURL)
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

		var respBody domain.GetCompanyVatNumberResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.Equal(t, "", respBody.VatNumber)
	})
}

// make test-func TEST_NAME=TestSetCompanyVatNumber TEST_PATH=test/integration/settings/company_vat_number_test.go

func TestSetCompanyVatNumber(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should successfully set a valid VAT number", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		vatNumber := "FR12345678901"
		request := domain.SetCompanyVatNumberRequest{VatNumber: vatNumber}
		req := th.NewSetCompanyVatNumberRequest(t, ctx, testServerURL, request)

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

		var respBody domain.SetCompanyVatNumberResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.True(t, respBody.Success)

		retrieved, err := th.GetCompanyVatNumber(t, ctx, testPool)
		assert.NoError(t, err)
		assert.Equal(t, vatNumber, retrieved)
	})

	// This is the core acceptance criterion for this field: a French business
	// under the "franchise en base de TVA" regime has no VAT number, and the
	// endpoint must accept a blank value without error.
	t.Run("should successfully set a blank VAT number without error", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		request := domain.SetCompanyVatNumberRequest{VatNumber: ""}
		req := th.NewSetCompanyVatNumberRequest(t, ctx, testServerURL, request)

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

		var respBody domain.SetCompanyVatNumberResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.True(t, respBody.Success)

		retrieved, err := th.GetCompanyVatNumber(t, ctx, testPool)
		assert.NoError(t, err)
		assert.Equal(t, "", retrieved)
	})

	t.Run("should return 400 for a malformed non-empty VAT number", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		request := domain.SetCompanyVatNumberRequest{VatNumber: "1"}
		req := th.NewSetCompanyVatNumberRequest(t, ctx, testServerURL, request)

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
		assert.Contains(t, respBody.Error, "vat_number_format")
	})

	t.Run("should return 401 when not authenticated as admin", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)

		request := domain.SetCompanyVatNumberRequest{VatNumber: "FR12345678901"}
		req := th.NewSetCompanyVatNumberRequest(t, ctx, testServerURL, request)

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("should successfully update an existing VAT number to blank", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		th.InsertCompanyVatNumber(t, ctx, "FR12345678901", testPool)

		request := domain.SetCompanyVatNumberRequest{VatNumber: ""}
		req := th.NewSetCompanyVatNumberRequest(t, ctx, testServerURL, request)

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

		retrieved, err := th.GetCompanyVatNumber(t, ctx, testPool)
		assert.NoError(t, err)
		assert.Equal(t, "", retrieved)
	})
}
