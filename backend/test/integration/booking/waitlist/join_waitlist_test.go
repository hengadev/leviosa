package waitlist

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	twaitlist "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/waitlist"
	tcatalog "github.com/Leviosa-care/leviosa/backend/test/helpers/catalog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestJoinWaitlist TEST_PATH=test/integration/booking/waitlist/join_waitlist_test.go

func TestJoinWaitlist(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("registered client joins the waitlist and the entry persists in the waiting state", func(t *testing.T) {
		twaitlist.ClearWaitlistTable(t, ctx, testPool)
		tcatalog.ClearProductsTable(t, ctx, testPool)

		categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
		products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)
		productID := products[0].ID

		clientID := uuid.New()
		start := time.Now().Add(24 * time.Hour)
		end := start.Add(7 * 24 * time.Hour)

		requestBody := map[string]interface{}{
			"product_id":          productID.String(),
			"client_id":           clientID,
			"desired_range_start": start.Format(time.RFC3339),
			"desired_range_end":   end.Format(time.RFC3339),
			"notes":               "Any weekday morning works",
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/waitlist", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var entry domain.WaitlistEntryResponse
		err = json.NewDecoder(resp.Body).Decode(&entry)
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, entry.ID)
		assert.Equal(t, productID, entry.ProductID)
		require.NotNil(t, entry.ClientID)
		assert.Equal(t, clientID, *entry.ClientID)
		assert.Equal(t, domain.WaitlistEntryStatusWaiting, entry.Status)
		assert.WithinDuration(t, start, entry.DesiredRangeStart, time.Second)
		assert.WithinDuration(t, end, entry.DesiredRangeEnd, time.Second)

		// Verify persistence directly and that the entry is still `waiting`
		// until something (issue 015) acts on it.
		persisted, err := twaitlist.GetWaitlistEntryEncxByID(t, ctx, testPool, entry.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.WaitlistEntryStatusWaiting, persisted.Status)
		require.NotNil(t, persisted.ClientID)
		assert.Equal(t, clientID, *persisted.ClientID)

		decrypted, err := domain.DecryptWaitlistEntryEncx(ctx, crypto, persisted)
		require.NoError(t, err)
		assert.Equal(t, productID, decrypted.ProductID)
		assert.WithinDuration(t, start, decrypted.DesiredRangeStart, time.Second)
		assert.WithinDuration(t, end, decrypted.DesiredRangeEnd, time.Second)
		assert.Equal(t, "Any weekday morning works", decrypted.Notes)
	})

	t.Run("guest joins the waitlist without an account", func(t *testing.T) {
		twaitlist.ClearWaitlistTable(t, ctx, testPool)
		tcatalog.ClearProductsTable(t, ctx, testPool)

		categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
		products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)
		productID := products[0].ID

		start := time.Now().Add(48 * time.Hour)
		end := start.Add(3 * 24 * time.Hour)

		requestBody := map[string]interface{}{
			"product_id":          productID.String(),
			"desired_range_start": start.Format(time.RFC3339),
			"desired_range_end":   end.Format(time.RFC3339),
			"guest_first_name":    "Jane",
			"guest_last_name":     "Doe",
			"guest_email":         "jane.doe@example.com",
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/waitlist", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var entry domain.WaitlistEntryResponse
		err = json.NewDecoder(resp.Body).Decode(&entry)
		require.NoError(t, err)

		assert.Nil(t, entry.ClientID)
		assert.Equal(t, "Jane", entry.GuestFirstName)
		assert.Equal(t, "Doe", entry.GuestLastName)
		assert.Equal(t, domain.WaitlistEntryStatusWaiting, entry.Status)

		persisted, err := twaitlist.GetWaitlistEntryEncxByID(t, ctx, testPool, entry.ID)
		require.NoError(t, err)
		assert.Nil(t, persisted.ClientID)

		decrypted, err := domain.DecryptWaitlistEntryEncx(ctx, crypto, persisted)
		require.NoError(t, err)
		assert.Equal(t, "Jane", decrypted.GuestFirstName)
		assert.Equal(t, "jane.doe@example.com", decrypted.GuestEmail)
	})

	t.Run("returns 400 when neither client ID nor guest fields are provided", func(t *testing.T) {
		twaitlist.ClearWaitlistTable(t, ctx, testPool)
		tcatalog.ClearProductsTable(t, ctx, testPool)

		categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
		products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)
		productID := products[0].ID

		start := time.Now().Add(24 * time.Hour)
		end := start.Add(24 * time.Hour)

		requestBody := map[string]interface{}{
			"product_id":          productID.String(),
			"desired_range_start": start.Format(time.RFC3339),
			"desired_range_end":   end.Format(time.RFC3339),
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/waitlist", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		count, err := twaitlist.CountWaitlistEntries(t, ctx, testPool)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("returns 404 for a non-existent product", func(t *testing.T) {
		twaitlist.ClearWaitlistTable(t, ctx, testPool)

		clientID := uuid.New()
		start := time.Now().Add(24 * time.Hour)
		end := start.Add(24 * time.Hour)

		requestBody := map[string]interface{}{
			"product_id":          uuid.New().String(),
			"client_id":           clientID,
			"desired_range_start": start.Format(time.RFC3339),
			"desired_range_end":   end.Format(time.RFC3339),
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/waitlist", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("returns 400 when the desired range has already elapsed", func(t *testing.T) {
		twaitlist.ClearWaitlistTable(t, ctx, testPool)
		tcatalog.ClearProductsTable(t, ctx, testPool)

		categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
		products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)
		productID := products[0].ID

		clientID := uuid.New()
		start := time.Now().Add(-48 * time.Hour)
		end := time.Now().Add(-24 * time.Hour)

		requestBody := map[string]interface{}{
			"product_id":          productID.String(),
			"client_id":           clientID,
			"desired_range_start": start.Format(time.RFC3339),
			"desired_range_end":   end.Format(time.RFC3339),
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/waitlist", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 415 for missing content-type header", func(t *testing.T) {
		req, err := http.NewRequest("POST", testServerURL+"/waitlist", bytes.NewReader([]byte("{}")))
		require.NoError(t, err)
		// Intentionally not setting Content-Type

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	})
}
