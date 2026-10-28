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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Practice page settings: which partner and product the public practice page
// shows (docs/issues/009-practice-page.md).
//
// make test-func TEST_NAME=TestSetPracticePagePartner TEST_PATH=test/integration/settings/practice_page_test.go

func TestGetPracticePageSettings(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should return 404 when the practice page partner is not set", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)

		req := th.NewGetPracticePagePartnerRequest(t, ctx, testServerURL)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("should return the partner and product set by an administrator, without authentication", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		partnerID := uuid.NewString()
		productID := uuid.NewString()
		th.InsertSettingString(t, ctx, &domain.Setting[string]{Key: settings.PracticePagePartnerID, Value: partnerID}, testPool)
		th.InsertSettingString(t, ctx, &domain.Setting[string]{Key: settings.PracticePageProductID, Value: productID}, testPool)

		partnerResp, err := client.Do(th.NewGetPracticePagePartnerRequest(t, ctx, testServerURL))
		assert.NoError(t, err)
		defer partnerResp.Body.Close()
		assert.Equal(t, http.StatusOK, partnerResp.StatusCode)
		var partnerBody domain.GetPracticePagePartnerResponse
		assert.NoError(t, json.NewDecoder(partnerResp.Body).Decode(&partnerBody))
		assert.Equal(t, partnerID, partnerBody.PartnerID)

		productResp, err := client.Do(th.NewGetPracticePageProductRequest(t, ctx, testServerURL))
		assert.NoError(t, err)
		defer productResp.Body.Close()
		assert.Equal(t, http.StatusOK, productResp.StatusCode)
		var productBody domain.GetPracticePageProductResponse
		assert.NoError(t, json.NewDecoder(productResp.Body).Decode(&productBody))
		assert.Equal(t, productID, productBody.ProductID)
	})
}

func TestSetPracticePagePartner(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should set the practice page partner and persist it", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		partnerID := uuid.NewString()

		req := th.NewSetPracticePagePartnerRequest(t, ctx, testServerURL, domain.SetPracticePagePartnerRequest{PartnerID: partnerID})
		for key, values := range tu.CreateAuthHeader(accessToken) {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respBody domain.SetPracticePagePartnerResponse
		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody))
		assert.True(t, respBody.Success)

		stored := th.GetSettingStringByKey(t, ctx, settings.PracticePagePartnerID, testPool)
		assert.Equal(t, partnerID, stored.Value)
	})

	t.Run("should return 400 for a non-UUID partner ID", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		req := th.NewSetPracticePagePartnerRequest(t, ctx, testServerURL, domain.SetPracticePagePartnerRequest{PartnerID: "not-a-uuid"})
		for key, values := range tu.CreateAuthHeader(accessToken) {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should reject an unauthenticated write", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)

		req := th.NewSetPracticePagePartnerRequest(t, ctx, testServerURL, domain.SetPracticePagePartnerRequest{PartnerID: uuid.NewString()})

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})
}

func TestSetPracticePageProduct(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should set the practice page product and persist it", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		productID := uuid.NewString()

		req := th.NewSetPracticePageProductRequest(t, ctx, testServerURL, domain.SetPracticePageProductRequest{ProductID: productID})
		for key, values := range tu.CreateAuthHeader(accessToken) {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		stored := th.GetSettingStringByKey(t, ctx, settings.PracticePageProductID, testPool)
		assert.Equal(t, productID, stored.Value)
	})

	t.Run("should return 400 for a non-UUID product ID", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		req := th.NewSetPracticePageProductRequest(t, ctx, testServerURL, domain.SetPracticePageProductRequest{ProductID: "42"})
		for key, values := range tu.CreateAuthHeader(accessToken) {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestSetPracticePageVideo(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	adminRequest := func(t *testing.T, request domain.SetPracticePageVideoRequest) *http.Response {
		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		req := th.NewSetPracticePageVideoRequest(t, ctx, testServerURL, request)
		for key, values := range tu.CreateAuthHeader(accessToken) {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		req.AddCookie(tu.CreateAuthCookie(accessToken))
		resp, err := client.Do(req)
		assert.NoError(t, err)
		return resp
	}

	t.Run("should set a video URL and expose it publicly", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		videoURL := "https://media.leviosa.care/practice.mp4"
		resp := adminRequest(t, domain.SetPracticePageVideoRequest{VideoURL: videoURL})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		getResp, err := client.Do(th.NewGetPracticePageVideoRequest(t, ctx, testServerURL))
		assert.NoError(t, err)
		defer getResp.Body.Close()
		assert.Equal(t, http.StatusOK, getResp.StatusCode)
		var body domain.GetPracticePageVideoResponse
		assert.NoError(t, json.NewDecoder(getResp.Body).Decode(&body))
		assert.Equal(t, videoURL, body.VideoURL)
	})

	t.Run("should accept an empty URL to clear the video", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		resp := adminRequest(t, domain.SetPracticePageVideoRequest{VideoURL: ""})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		stored := th.GetSettingStringByKey(t, ctx, settings.PracticePageVideoURL, testPool)
		assert.Empty(t, stored.Value)
	})

	t.Run("should return 400 for a non-http URL", func(t *testing.T) {
		th.ClearSettingTestData(t, ctx, testPool, s3Client)
		defer tu.ClearAuthData(t, ctx, authCtx)

		resp := adminRequest(t, domain.SetPracticePageVideoRequest{VideoURL: "ftp://example.com/video.mp4"})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
