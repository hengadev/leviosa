package testutils

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/hengadev/encx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/services"
)

func newTestServiceKeyStore(t *testing.T) *ServiceKeyStore {
	t.Helper()
	crypto, err := encx.NewTestCrypto(t)
	require.NoError(t, err)
	return NewServiceKeyStore(crypto)
}

func TestGenerateServiceKey(t *testing.T) {
	store := &ServiceKeyStore{} // key generation needs no crypto

	key, err := store.GenerateServiceKey()
	require.NoError(t, err)

	// Test key properties
	assert.NotEmpty(t, key, "Generated key should not be empty")
	assert.Greater(t, len(key), 40, "Generated key should be sufficiently long")

	// Test that key is valid base64
	_, err = base64.URLEncoding.DecodeString(key)
	assert.NoError(t, err, "Generated key should be valid base64")

	// Test uniqueness - generate multiple keys
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := store.GenerateServiceKey()
		require.NoError(t, err)
		assert.False(t, keys[key], "Generated keys should be unique")
		keys[key] = true
	}
}

func TestServiceKeyStore_GenerateAndValidateSingleServiceKey(t *testing.T) {
	ctx := context.Background()
	store := newTestServiceKeyStore(t)

	apiKey, err := store.GenerateAndStoreServiceKey(ctx, services.Settings)
	require.NoError(t, err)
	require.NotEmpty(t, apiKey)

	isValid, err := store.ValidateServiceKey(ctx, services.Settings, apiKey)
	require.NoError(t, err)
	assert.True(t, isValid, "Generated key should be valid")

	isValid, err = store.ValidateServiceKey(ctx, services.Settings, "wrong-key")
	require.NoError(t, err)
	assert.False(t, isValid, "Wrong key should not be valid")
}

func TestServiceKeyStore_GenerateAllServiceKeys(t *testing.T) {
	ctx := context.Background()
	store := newTestServiceKeyStore(t)

	serviceKeys, err := store.GenerateAllServiceKeys(ctx)
	require.NoError(t, err)

	expectedServices := services.AllServices()
	assert.Len(t, serviceKeys, len(expectedServices))
	for _, serviceName := range expectedServices {
		apiKey, exists := serviceKeys[serviceName]
		assert.True(t, exists, "Service %s should have an API key", serviceName)
		assert.NotEmpty(t, apiKey, "API key for %s should not be empty", serviceName)

		isValid, err := store.ValidateServiceKey(ctx, serviceName, apiKey)
		require.NoError(t, err)
		assert.True(t, isValid, "Generated key for %s should be valid", serviceName)
	}
}

func TestServiceKeyStore_RejectsUnknownService(t *testing.T) {
	ctx := context.Background()
	store := newTestServiceKeyStore(t)

	_, err := store.GenerateAndStoreServiceKey(ctx, "nonexistent")
	assert.Error(t, err)
	_, err = store.ValidateServiceKey(ctx, "nonexistent", "any-key")
	assert.Error(t, err)
}

// Read serves what the auth middleware looks up: the key hash at the service's key path.
func TestServiceKeyStore_ReadServesTheKeyHashAtTheServicePath(t *testing.T) {
	ctx := context.Background()
	crypto, err := encx.NewTestCrypto(t)
	require.NoError(t, err)
	store := NewServiceKeyStore(crypto)

	apiKey, err := store.GenerateAndStoreServiceKey(ctx, services.Catalog)
	require.NoError(t, err)

	secret, err := store.Read(ctx, services.ServiceAPIKeyPath(services.Catalog))
	require.NoError(t, err)
	require.NotNil(t, secret)
	data, ok := secret.Data["data"].(map[string]interface{})
	require.True(t, ok, "entries are nested under data")
	keyBytes, err := encx.SerializeValue(apiKey)
	require.NoError(t, err)
	assert.Equal(t, crypto.HashBasic(ctx, keyBytes), data["key_hash"])

	missing, err := store.Read(ctx, services.ServiceAPIKeyPath(services.Settings))
	require.NoError(t, err)
	assert.Nil(t, missing, "a service without a stored key reads as absent")
}
