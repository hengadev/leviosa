package price_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	priceHandler "github.com/Leviosa-care/leviosa/backend/internal/catalog/interface/price"
	ck "github.com/Leviosa-care/leviosa/backend/internal/common/auth/cookies"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	td "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestReplaceOneTimePrice TEST_PATH=test/integration/catalog/price/replace_one_time_price_test.go

// The admin edit page changes a product's price through
// PUT /admin/products/{id}/price: a new one-time price, the old one
// deactivated (a Stripe price cannot be edited).
func TestReplaceOneTimePrice(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	replace := func(t *testing.T, productID string, body any, accessToken string) *http.Response {
		t.Helper()
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		req, err := http.NewRequestWithContext(ctx, http.MethodPut,
			testServerURL+buildEndpointWithID(priceHandler.ReplaceOneTimePriceEndpoint, productID), bytes.NewReader(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		if accessToken != "" {
			req.AddCookie(&http.Cookie{Name: ck.AccessTokenCookieName, Value: accessToken})
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		return resp
	}

	activePrices := func(t *testing.T, productID string) []struct {
		Amount   int
		Currency string
		Interval string
	} {
		t.Helper()
		rows, err := testPool.Query(ctx, `SELECT amount, currency, interval::text FROM catalog.prices WHERE product_id = $1 AND is_active`, productID)
		require.NoError(t, err)
		defer rows.Close()
		var out []struct {
			Amount   int
			Currency string
			Interval string
		}
		for rows.Next() {
			var p struct {
				Amount   int
				Currency string
				Interval string
			}
			require.NoError(t, rows.Scan(&p.Amount, &p.Currency, &p.Interval))
			out = append(out, p)
		}
		return out
	}

	t.Run("should replace the active one-time price", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		productID := setupTestProduct(t, ctx)
		resp := createPriceViaAPIWithToken(t, productID, &domain.CreatePriceRequest{Amount: 7000, Currency: "EUR", Interval: "one_time"}, accessToken)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		res := replace(t, productID, domain.CreatePriceRequest{Amount: 8000, Currency: "eur", Interval: "one_time"}, accessToken)
		defer res.Body.Close()
		require.Equal(t, http.StatusOK, res.StatusCode)

		active := activePrices(t, productID)
		require.Len(t, active, 1, "only the new price is active")
		assert.Equal(t, 8000, active[0].Amount)
		assert.Equal(t, "EUR", active[0].Currency)
		assert.Equal(t, "one_time", active[0].Interval)

		var total int
		require.NoError(t, testPool.QueryRow(ctx, `SELECT count(*) FROM catalog.prices WHERE product_id = $1`, productID).Scan(&total))
		assert.Equal(t, 2, total, "the old price is kept, inactive")
	})

	t.Run("should refuse a price below the fixed deposit and keep the old one", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		td.ClearProductsTable(t, ctx, testPool)
		td.ClearPricesTable(t, ctx, testPool)
		td.ClearCategoriesTable(t, ctx, testPool)
		category := td.NewValidCategory("deposit category")
		td.InsertCategory(t, ctx, category, testPool)
		product := td.NewValidProduct("Deposit Product", category.ID)
		product.Deposit = &domain.Deposit{Type: domain.DepositFixed, Value: 3000}
		td.InsertProduct(t, ctx, testPool, product)
		productID := product.ID.String()
		resp := createPriceViaAPIWithToken(t, productID, &domain.CreatePriceRequest{Amount: 7000, Currency: "EUR", Interval: "one_time"}, accessToken)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		res := replace(t, productID, domain.CreatePriceRequest{Amount: 2000, Currency: "EUR", Interval: "one_time"}, accessToken)
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		active := activePrices(t, productID)
		require.Len(t, active, 1)
		assert.Equal(t, 7000, active[0].Amount)
	})

	t.Run("should refuse a recurring price", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		productID := setupTestProduct(t, ctx)

		res := replace(t, productID, domain.CreatePriceRequest{Amount: 8000, Currency: "EUR", Interval: "month"}, accessToken)
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("should refuse a request without a token", func(t *testing.T) {
		productID := setupTestProduct(t, ctx)

		res := replace(t, productID, domain.CreatePriceRequest{Amount: 8000, Currency: "EUR", Interval: "one_time"}, "")
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})
}

// createPriceViaAPIWithToken creates a price as an admin.
func createPriceViaAPIWithToken(t *testing.T, productID string, request *domain.CreatePriceRequest, accessToken string) *http.Response {
	t.Helper()
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(td.NewCreatePriceRequest(t, context.Background(), testServerURL, productID, request, accessToken))
	require.NoError(t, err)
	resp.Body.Close()
	return resp
}
