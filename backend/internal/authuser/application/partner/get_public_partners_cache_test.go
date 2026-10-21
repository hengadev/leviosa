package partner_test

import (
	"context"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	td "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/google/uuid"
	"github.com/hengadev/encx"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestGetPublicPartners_DEKCacheSkipsKMSOnSecondLoad TEST_PATH=internal/authuser/application/partner/get_public_partners_cache_test.go

// countingKMS wraps a key provider and counts the KMS decrypt calls: one call
// per data key the cache cannot serve. Everything else (key resolution,
// encryption) is passed through untouched.
type countingKMS struct {
	encx.KeyManagementService
	decryptCalls int
}

func (k *countingKMS) DecryptDEK(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	k.decryptCalls++
	return k.KeyManagementService.DecryptDEK(ctx, keyID, ciphertext)
}

// newCachingCrypto builds a real encx service with the DEK cache enabled the
// way encryption.NewCrypto enables it by default (zero values select encx's
// defaults: 4096 entries, 10 minutes), on an in-memory KMS whose decrypt calls
// are counted.
func newCachingCrypto(t *testing.T) (*encx.Crypto, *countingKMS) {
	t.Helper()

	kms := &countingKMS{KeyManagementService: encx.NewSimpleTestKMS()}
	crypto, err := encx.NewCrypto(context.Background(), kms, encx.NewInMemorySecretStore(), encx.Config{
		KEKAlias:    "alias/test-kek",
		PepperAlias: "test-service",
		DBPath:      t.TempDir(),
	}, encx.WithDEKCache(0, 0))
	require.NoError(t, err)

	return crypto, kms
}

// encryptedRow builds a listing row the way the repositories store one: a data
// key wrapped by the KEK and first/last names encrypted under that data key.
func encryptedRow(t *testing.T, crypto *encx.Crypto, first, last string) *domain.PublicPartnerRow {
	t.Helper()
	ctx := context.Background()

	dek, err := crypto.GenerateDEK()
	require.NoError(t, err)
	wrapped, err := crypto.EncryptDEK(ctx, dek)
	require.NoError(t, err)

	encryptField := func(value string) []byte {
		serialized, err := encx.SerializeValue(value)
		require.NoError(t, err)
		encrypted, err := crypto.EncryptData(ctx, serialized, dek)
		require.NoError(t, err)
		return encrypted
	}

	return &domain.PublicPartnerRow{
		PartnerEncx:        td.NewTestPartnerEncxWithUserID(t, uuid.New()),
		FirstNameEncrypted: encryptField(first),
		LastNameEncrypted:  encryptField(last),
		UserDEKEncrypted:   wrapped,
		UserKeyVersion:     1,
	}
}

// Issue 031: the public partner listing decrypts one data key per partner, and
// the page is public — without the DEK cache every anonymous visit costs one
// sequential KMS round trip per partner. With the cache, the second load of the
// same listing makes no KMS decrypt call at all.
func TestGetPublicPartners_DEKCacheSkipsKMSOnSecondLoad(t *testing.T) {
	ctx := context.Background()

	crypto, kms := newCachingCrypto(t)

	rows := []*domain.PublicPartnerRow{
		encryptedRow(t, crypto, "Jeanne", "Moreau"),
		encryptedRow(t, crypto, "Paul", "Verlaine"),
		encryptedRow(t, crypto, "Camille", "Claudel"),
	}

	mockPartnerRepo := new(mockPartnerRepository)
	mockPartnerRepo.On("GetPublicPartners", ctx).Return(rows, nil)

	svc := setupPartnerService(
		mockPartnerRepo, new(mockUserRepository), new(mockProductService), new(mockCategoryService), crypto, new(mockStripeService),
	)

	first, err := svc.GetPublicPartners(ctx)
	require.NoError(t, err)
	require.Len(t, first, 3)
	for i, want := range []string{"Jeanne", "Paul", "Camille"} {
		require.Equal(t, want, first[i].FirstName, "partner %d's first name", i)
	}
	require.Equal(t, 3, kms.decryptCalls, "the first load unwraps one data key per partner")

	second, err := svc.GetPublicPartners(ctx)
	require.NoError(t, err)
	require.Len(t, second, 3)
	require.Equal(t, 3, kms.decryptCalls, "the second load must be served from the DEK cache: no KMS decrypt call")
}
