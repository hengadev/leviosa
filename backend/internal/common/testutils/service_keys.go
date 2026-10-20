package testutils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/hengadev/encx"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/services"
	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware/auth"
)

// ServiceKeyStore is an in-memory service key store for tests of the auth middleware's
// RequireServiceAuth: it generates service API keys and serves their hashes through
// auth.SecretReader. The app itself runs without a service key store (ADR-0015), so
// service-to-service endpoints only accept requests in tests.
type ServiceKeyStore struct {
	crypto encx.CryptoService

	mu     sync.RWMutex
	hashes map[string]string // services.ServiceAPIKeyPath(name) -> key hash
}

var _ auth.SecretReader = (*ServiceKeyStore)(nil)

// NewServiceKeyStore returns an empty store that hashes keys with crypto, as the
// middleware does.
func NewServiceKeyStore(crypto encx.CryptoService) *ServiceKeyStore {
	return &ServiceKeyStore{crypto: crypto, hashes: make(map[string]string)}
}

// GenerateServiceKey generates a cryptographically secure API key.
func (s *ServiceKeyStore) GenerateServiceKey() (string, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return base64.URLEncoding.EncodeToString(keyBytes), nil
}

// StoreServiceKey stores the hash of apiKey as serviceName's key.
func (s *ServiceKeyStore) StoreServiceKey(ctx context.Context, serviceName, apiKey string) error {
	if !services.IsValidService(serviceName) {
		return fmt.Errorf("invalid service name: %s", serviceName)
	}
	keyHash, err := s.hash(ctx, apiKey)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.hashes[services.ServiceAPIKeyPath(serviceName)] = keyHash
	return nil
}

// GenerateAndStoreServiceKey generates a new API key for serviceName, stores its hash and
// returns the key.
func (s *ServiceKeyStore) GenerateAndStoreServiceKey(ctx context.Context, serviceName string) (string, error) {
	apiKey, err := s.GenerateServiceKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate service key for %s: %w", serviceName, err)
	}
	if err := s.StoreServiceKey(ctx, serviceName, apiKey); err != nil {
		return "", fmt.Errorf("failed to store service key for %s: %w", serviceName, err)
	}
	return apiKey, nil
}

// GenerateAllServiceKeys generates and stores a key for every service and returns them
// by service name.
func (s *ServiceKeyStore) GenerateAllServiceKeys(ctx context.Context) (map[string]string, error) {
	serviceKeys := make(map[string]string)
	for _, serviceName := range services.AllServices() {
		apiKey, err := s.GenerateAndStoreServiceKey(ctx, serviceName)
		if err != nil {
			return nil, err
		}
		serviceKeys[serviceName] = apiKey
	}
	return serviceKeys, nil
}

// ValidateServiceKey reports whether providedKey is serviceName's stored key.
func (s *ServiceKeyStore) ValidateServiceKey(ctx context.Context, serviceName, providedKey string) (bool, error) {
	if !services.IsValidService(serviceName) {
		return false, fmt.Errorf("invalid service name: %s", serviceName)
	}

	s.mu.RLock()
	storedHash, ok := s.hashes[services.ServiceAPIKeyPath(serviceName)]
	s.mu.RUnlock()
	if !ok {
		return false, fmt.Errorf("no service key stored for service: %s", serviceName)
	}

	providedHash, err := s.hash(ctx, providedKey)
	if err != nil {
		return false, err
	}
	return storedHash == providedHash, nil
}

// Read implements auth.SecretReader: the key hash stored at path, in the layout the
// middleware reads (Data["data"]["key_hash"]), or nil when nothing is stored there.
func (s *ServiceKeyStore) Read(_ context.Context, path string) (*auth.SecretData, error) {
	s.mu.RLock()
	keyHash, ok := s.hashes[path]
	s.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return &auth.SecretData{Data: map[string]interface{}{
		"data": map[string]interface{}{"key_hash": keyHash},
	}}, nil
}

func (s *ServiceKeyStore) hash(ctx context.Context, apiKey string) (string, error) {
	keyBytes, err := encx.SerializeValue(apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to serialize API key for hashing: %w", err)
	}
	return s.crypto.HashBasic(ctx, keyBytes), nil
}
