package productPayment_test

import (
	"context"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("successful product update with all fields", func(t *testing.T) {
		// First create a product
		createReq := domain.CreateStripeProductRequest{
			Name:        "Original Product",
			Description: "Original description",
			Metadata: map[string]string{
				"version": "1.0",
			},
		}

		createdProduct, err := stripeService.CreateProduct(ctx, createReq)
		require.NoError(t, err)

		// Update the product
		newName := "Updated Product"
		newDescription := "Updated description"
		newMetadata := map[string]string{
			"version": "2.0",
			"updated": "true",
		}

		updateReq := &domain.UpdateStripeProductRequest{
			Name:        &newName,
			Description: &newDescription,
			Metadata:    newMetadata,
		}

		updatedProduct, err := stripeService.UpdateProduct(ctx, createdProduct.ID, updateReq)

		require.NoError(t, err)
		assert.NotNil(t, updatedProduct)
		assert.Equal(t, createdProduct.ID, updatedProduct.ID)
		assert.Equal(t, newName, updatedProduct.Name)
		assert.Equal(t, newDescription, updatedProduct.Description)
		assert.True(t, updatedProduct.Active)
	})

	t.Run("successful partial update - name only", func(t *testing.T) {
		// Create a product
		createReq := domain.CreateStripeProductRequest{
			Name:        "Original Name",
			Description: "Original description",
		}

		createdProduct, err := stripeService.CreateProduct(ctx, createReq)
		require.NoError(t, err)

		// Update only the name
		newName := "Updated Name Only"
		updateReq := &domain.UpdateStripeProductRequest{
			Name: &newName,
		}

		updatedProduct, err := stripeService.UpdateProduct(ctx, createdProduct.ID, updateReq)

		require.NoError(t, err)
		assert.Equal(t, newName, updatedProduct.Name)
	})

	t.Run("successful partial update - description only", func(t *testing.T) {
		// Create a product
		createReq := domain.CreateStripeProductRequest{
			Name:        "Product Name",
			Description: "Original description",
		}

		createdProduct, err := stripeService.CreateProduct(ctx, createReq)
		require.NoError(t, err)

		// Update only the description
		newDescription := "Updated description only"
		updateReq := &domain.UpdateStripeProductRequest{
			Description: &newDescription,
		}

		updatedProduct, err := stripeService.UpdateProduct(ctx, createdProduct.ID, updateReq)

		require.NoError(t, err)
		assert.Equal(t, newDescription, updatedProduct.Description)
	})

	t.Run("successful partial update - metadata only", func(t *testing.T) {
		// Create a product
		createReq := domain.CreateStripeProductRequest{
			Name:        "Product Name",
			Description: "Product description",
			Metadata: map[string]string{
				"original": "value",
			},
		}

		createdProduct, err := stripeService.CreateProduct(ctx, createReq)
		require.NoError(t, err)

		// Update only the metadata
		newMetadata := map[string]string{
			"updated":  "metadata",
			"version":  "2.0",
			"category": "new",
		}
		updateReq := &domain.UpdateStripeProductRequest{
			Metadata: newMetadata,
		}

		_, err = stripeService.UpdateProduct(ctx, createdProduct.ID, updateReq)

		require.NoError(t, err)
	})

	t.Run("update with empty request does nothing", func(t *testing.T) {
		// Create a product
		createReq := domain.CreateStripeProductRequest{
			Name:        "Original Product",
			Description: "Original description",
		}

		createdProduct, err := stripeService.CreateProduct(ctx, createReq)
		require.NoError(t, err)

		// Update with empty request
		updateReq := &domain.UpdateStripeProductRequest{}

		_, err = stripeService.UpdateProduct(ctx, createdProduct.ID, updateReq)

		require.NoError(t, err)
	})
}