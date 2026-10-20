package testutils

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateServiceKey(t *testing.T) {
	// We can test key generation without Vault
	skm := &ServiceKeyManager{} // No Vault client needed for key generation

	key, err := skm.GenerateServiceKey()
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
		key, err := skm.GenerateServiceKey()
		require.NoError(t, err)
		assert.False(t, keys[key], "Generated keys should be unique")
		keys[key] = true
	}
}
