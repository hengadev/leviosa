package price_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	priceHandler "github.com/Leviosa-care/leviosa/backend/internal/catalog/interface/price"
	td "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME='^TestGetPublicProductPrices$$' TEST_PATH=test/integration/catalog/price/get_public_product_prices_test.go

func TestGetPublicProductPrices(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	// insertProduct adds a product with the given status and one active price.
	insertProduct := func(t *testing.T, status domain.PublishedStatus) *domain.Product {
		category := td.NewValidCategory("Prices Category " + string(status))
		td.InsertCategory(t, ctx, category, testPool)
		product := td.NewValidProduct("Prices Product "+string(status), category.ID)
		product.Status = status
		td.InsertProduct(t, ctx, testPool, product)
		price := td.NewValidPrice()
		price.ProductID = product.ID
		price.IsActive = true
		td.InsertPrice(t, ctx, price, testPool)
		return product
	}
	get := func(t *testing.T, productID string) *http.Response {
		res, err := client.Get(testServerURL + buildEndpointWithID(priceHandler.GetPublicProductPricesEndpoint, productID))
		require.NoError(t, err)
		return res
	}
	clear := func(t *testing.T) {
		td.ClearPricesTable(t, ctx, testPool)
		td.ClearProductsTable(t, ctx, testPool)
		td.ClearCategoriesTable(t, ctx, testPool)
	}

	t.Run("returns the active prices of a published product", func(t *testing.T) {
		clear(t)
		product := insertProduct(t, domain.Published)

		res := get(t, product.ID.String())
		defer res.Body.Close()
		require.Equal(t, http.StatusOK, res.StatusCode)

		var prices []priceHandler.PublicPriceResponse
		require.NoError(t, json.NewDecoder(res.Body).Decode(&prices))
		assert.Len(t, prices, 1)
	})

	t.Run("returns 404 for a draft or archived product", func(t *testing.T) {
		clear(t)
		for _, status := range []domain.PublishedStatus{domain.Draft, domain.Archived} {
			product := insertProduct(t, status)

			res := get(t, product.ID.String())
			res.Body.Close()
			assert.Equal(t, http.StatusNotFound, res.StatusCode, "a %s product's prices are not public", status)
		}
	})
}
