package product_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	th "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME='^TestGetAdminProductByID$$' TEST_PATH=test/integration/catalog/product/get_admin_product_by_id_test.go

// The admin product pages (edit, duplicate) load a product through
// GET /admin/products/{id}, drafts included.
func TestGetAdminProductByID(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should return a draft product with its prices to an admin", func(t *testing.T) {
		clearTables(t, ctx)
		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		category := th.NewValidCategory("Ostéopathie")
		th.InsertCategory(t, ctx, category, testPool)
		product := th.NewValidProduct("Séance", category.ID)
		th.InsertProduct(t, ctx, testPool, product)
		price := th.NewValidPrice()
		price.ProductID = product.ID
		price.Interval = "one_time"
		th.InsertPrice(t, ctx, price, testPool)

		resp, err := client.Do(th.NewGetAdminProductByIDRequest(t, ctx, testServerURL, product.ID.String(), accessToken))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var got domain.ProductAggregator
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		require.NotNil(t, got.Product)
		assert.Equal(t, product.ID, got.Product.ID)
		assert.Equal(t, category.ID, got.Product.Category.ID)
		require.Len(t, got.Prices, 1)
		assert.Equal(t, price.Amount, got.Prices[0].Amount)
	})

	t.Run("should return 404 for a product that does not exist", func(t *testing.T) {
		clearTables(t, ctx)
		accessToken := tu.SetupAdminUser(t, ctx, authCtx)

		resp, err := client.Do(th.NewGetAdminProductByIDRequest(t, ctx, testServerURL, "8f0c5d9e-3b1a-4c2d-9e7f-1a2b3c4d5e6f", accessToken))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("should refuse a request without a token", func(t *testing.T) {
		clearTables(t, ctx)

		resp, err := client.Do(th.NewGetAdminProductByIDRequest(t, ctx, testServerURL, "8f0c5d9e-3b1a-4c2d-9e7f-1a2b3c4d5e6f", ""))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should refuse a user who is not an admin", func(t *testing.T) {
		clearTables(t, ctx)
		accessToken := tu.SetupStandardUser(t, ctx, authCtx)

		resp, err := client.Do(th.NewGetAdminProductByIDRequest(t, ctx, testServerURL, "8f0c5d9e-3b1a-4c2d-9e7f-1a2b3c4d5e6f", accessToken))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
