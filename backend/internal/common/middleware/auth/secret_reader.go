package auth

import "context"

// SecretData represents the data returned from a secret store.
type SecretData struct {
	Data map[string]interface{}
}

// SecretReader abstracts reading service API key hashes from a secrets store. Entries
// follow the KV v2 layout: Data["data"]["key_hash"].
type SecretReader interface {
	// Read reads the secret at the given path and returns its top-level data map.
	// Returns nil SecretData when the path does not exist.
	Read(ctx context.Context, path string) (*SecretData, error)
}
