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

// make test-func TEST_NAME='^TestProductDeposit$' TEST_PATH=test/integration/catalog/product/deposit_test.go

// TestProductDeposit covers issue 013: a product's optional deposit
// round-trips through the API and the database, and a deposit that breaks
// the domain rules (fixed amount above the price, percentage outside
// (0, 100]) is rejected.
func TestProductDeposit(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { clearTables(t, ctx) })
	client := &http.Client{Timeout: 10 * time.Second}

	newRequest := func(categoryID string, deposit *domain.Deposit, amount int) domain.CreateProductWithPriceRequest {
		return domain.CreateProductWithPriceRequest{
			Product: domain.CreateProductRequest{
				CategoryID:        categoryID,
				Name:              "Séance deposit",
				Description:       "Product with a deposit",
				Duration:          60,
				Availability:      domain.Online,
				BufferTime:        10,
				CancellationHours: 24,
				Deposit:           deposit,
			},
			Price: domain.CreatePriceRequest{Amount: amount, Currency: "EUR", Interval: "month"},
		}
	}

	createProduct := func(t *testing.T, body domain.CreateProductWithPriceRequest, accessToken string) (*http.Response, string) {
		t.Helper()
		res, err := client.Do(th.NewCreateProductWithPriceRequest(t, ctx, testServerURL, body, accessToken))
		require.NoError(t, err)
		t.Cleanup(func() { res.Body.Close() })
		var out struct {
			ProductID string `json:"product_id"`
		}
		if res.StatusCode == http.StatusCreated {
			require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
		}
		return res, out.ProductID
	}

	countProducts := func(t *testing.T) int {
		t.Helper()
		var n int
		require.NoError(t, testPool.QueryRow(ctx, "SELECT count(*) FROM catalog.products").Scan(&n))
		return n
	}

	for _, tc := range []struct {
		name    string
		deposit *domain.Deposit
	}{
		{name: "fixed", deposit: &domain.Deposit{Type: domain.DepositFixed, Value: 2000}},
		{name: "percentage", deposit: &domain.Deposit{Type: domain.DepositPercent, Value: 30}},
		{name: "none", deposit: nil},
	} {
		t.Run("a product with a "+tc.name+" deposit round-trips", func(t *testing.T) {
			clearTables(t, ctx)
			accessToken := tu.SetupAdminUser(t, ctx, authCtx)
			category := th.NewValidCategory("Deposit Category")
			th.InsertCategory(t, ctx, category, testPool)

			res, productID := createProduct(t, newRequest(category.ID.String(), tc.deposit, 7000), accessToken)
			require.Equal(t, http.StatusCreated, res.StatusCode)

			// A new product is a draft: only the admin route returns it.
			getRes, err := client.Do(th.NewGetAdminProductByIDRequest(t, ctx, testServerURL, productID, accessToken))
			require.NoError(t, err)
			defer getRes.Body.Close()
			require.Equal(t, http.StatusOK, getRes.StatusCode)

			var got domain.ProductAggregator
			require.NoError(t, json.NewDecoder(getRes.Body).Decode(&got))
			assert.Equal(t, tc.deposit, got.Product.Deposit)
		})
	}

	for _, tc := range []struct {
		name    string
		deposit *domain.Deposit
	}{
		{name: "a fixed deposit above the price", deposit: &domain.Deposit{Type: domain.DepositFixed, Value: 7001}},
		{name: "a negative fixed deposit", deposit: &domain.Deposit{Type: domain.DepositFixed, Value: -1}},
		{name: "a percentage of zero", deposit: &domain.Deposit{Type: domain.DepositPercent, Value: 0}},
		{name: "a percentage above 100", deposit: &domain.Deposit{Type: domain.DepositPercent, Value: 101}},
		{name: "an unknown deposit type", deposit: &domain.Deposit{Type: "arrhes", Value: 10}},
	} {
		t.Run("creating a product with "+tc.name+" is rejected", func(t *testing.T) {
			clearTables(t, ctx)
			accessToken := tu.SetupAdminUser(t, ctx, authCtx)
			category := th.NewValidCategory("Deposit Category")
			th.InsertCategory(t, ctx, category, testPool)

			res, _ := createProduct(t, newRequest(category.ID.String(), tc.deposit, 7000), accessToken)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Zero(t, countProducts(t), "no product is created")
		})
	}

	setupProductWithPrice := func(t *testing.T, amount int) *domain.Product {
		t.Helper()
		category := th.NewValidCategory("Deposit Category")
		th.InsertCategory(t, ctx, category, testPool)
		product := th.NewValidProduct("Deposit Product", category.ID)
		th.InsertProduct(t, ctx, testPool, product)
		price := th.NewValidPrice()
		price.ProductID = product.ID
		price.Amount = amount
		price.Interval = "one_time"
		th.InsertPrice(t, ctx, price, testPool)
		return product
	}

	modify := func(t *testing.T, productID string, body any, accessToken string) int {
		t.Helper()
		res, err := client.Do(th.NewModifyProductRequest(t, ctx, testServerURL, productID, body, accessToken))
		require.NoError(t, err)
		defer res.Body.Close()
		return res.StatusCode
	}

	t.Run("modifying sets, keeps and removes the deposit", func(t *testing.T) {
		clearTables(t, ctx)
		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		product := setupProductWithPrice(t, 7000)

		deposit := &domain.Deposit{Type: domain.DepositFixed, Value: 2000}
		require.Equal(t, http.StatusNoContent, modify(t, product.ID.String(),
			domain.UpdateProductRequest{Deposit: domain.DepositPatch{Present: true, Value: deposit}}, accessToken))
		stored, err := th.GetProductByID(t, ctx, product.ID, testPool)
		require.NoError(t, err)
		assert.Equal(t, deposit, stored.Deposit)

		// An update that does not mention the deposit leaves it alone.
		newName := "Renamed"
		require.Equal(t, http.StatusNoContent, modify(t, product.ID.String(),
			domain.UpdateProductRequest{Name: &newName}, accessToken))
		stored, err = th.GetProductByID(t, ctx, product.ID, testPool)
		require.NoError(t, err)
		assert.Equal(t, deposit, stored.Deposit)

		// JSON null removes it: the full price is due online again.
		require.Equal(t, http.StatusNoContent, modify(t, product.ID.String(),
			map[string]any{"deposit": nil}, accessToken))
		stored, err = th.GetProductByID(t, ctx, product.ID, testPool)
		require.NoError(t, err)
		assert.Nil(t, stored.Deposit)
	})

	t.Run("modifying to a fixed deposit above an active price is rejected", func(t *testing.T) {
		clearTables(t, ctx)
		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		product := setupProductWithPrice(t, 7000)

		status := modify(t, product.ID.String(), map[string]any{
			"deposit": map[string]any{"type": "fixed", "value": 7001},
		}, accessToken)
		assert.Equal(t, http.StatusBadRequest, status)
		stored, err := th.GetProductByID(t, ctx, product.ID, testPool)
		require.NoError(t, err)
		assert.Nil(t, stored.Deposit)
	})

	t.Run("modifying to an invalid percentage is rejected", func(t *testing.T) {
		clearTables(t, ctx)
		accessToken := tu.SetupAdminUser(t, ctx, authCtx)
		product := setupProductWithPrice(t, 7000)

		status := modify(t, product.ID.String(), map[string]any{
			"deposit": map[string]any{"type": "percent", "value": 150},
		}, accessToken)
		assert.Equal(t, http.StatusBadRequest, status)
	})
}
