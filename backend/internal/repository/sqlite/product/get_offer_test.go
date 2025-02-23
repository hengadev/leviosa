package productRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/product"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/product"
	"github.com/GaryHY/leviosa/tests/utils"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	"github.com/GaryHY/test-assert"
)

func TestGetOffer(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", test.GetSQLiteMigrationPath())
	offer := factories.NewBasicOffer(nil)
	tests := []struct {
		name          string
		offerID       string
		version       int64
		expectedErr   error
		expectedOffer *productService.Offer
	}{
		{
			name:          "offer not in database",
			offerID:       test.GenerateRandomString(16),
			version:       20250217155108,
			expectedErr:   rp.ErrNotFound,
			expectedOffer: nil,
		},
		{
			name:          "nominal case",
			offerID:       offer.ID,
			version:       20250217165531,
			expectedErr:   nil,
			expectedOffer: offer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, productRepository.New)
			defer teardown()
			offer, err := repo.GetOffer(ctx, tt.offerID)
			assert.EqualError(t, err, tt.expectedErr)
			assert.ReflectEqual(t, offer, tt.expectedOffer)
		})
	}
}
