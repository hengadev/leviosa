package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/services"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	mw "github.com/Leviosa-care/leviosa/backend/internal/common/middleware"

	"github.com/hengadev/encx"
)

// ServiceInfo contains service authentication details for request context
type ServiceInfo struct {
	Name string `json:"name"`
}

type serviceContextKeyType struct{}

// serviceContextKey is the unexported key used to store service info in request context.
var serviceContextKey = serviceContextKeyType{}

// RequireServiceAuth validates service authentication headers and makes service info available in context
func (m *SessionAuthMiddleware) RequireServiceAuth(next mw.Handler) mw.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := ctxutil.GetLoggerFromContext(ctx)
		if err != nil {
			httpx.RespondWithError(w, err, http.StatusInternalServerError)
			return
		}

		// Extract service authentication headers
		serviceName := r.Header.Get(services.ServiceNameHeader)
		serviceKey := r.Header.Get(services.ServiceKeyHeader)

		if serviceName == "" {
			logger.WarnContext(ctx, "Service auth middleware: Missing service name header",
				"operation", "require_service_auth",
				"method", r.Method,
				"path", r.URL.Path,
				"header", services.ServiceNameHeader)
			httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
			return
		}

		if serviceKey == "" {
			logger.WarnContext(ctx, "Service auth middleware: Missing service key header",
				"operation", "require_service_auth",
				"method", r.Method,
				"path", r.URL.Path,
				"service_name", serviceName,
				"header", services.ServiceKeyHeader)
			httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
			return
		}

		// Validate service name
		if !services.IsValidService(serviceName) {
			logger.WarnContext(ctx, "Service auth middleware: Invalid service name",
				"operation", "require_service_auth",
				"method", r.Method,
				"path", r.URL.Path,
				"service_name", serviceName,
				"valid_services", services.AllServices())
			httpx.RespondWithError(w, errs.ErrInvalidValue, http.StatusBadRequest)
			return
		}

		// Validate service key against the service key store (with caching)
		if !m.validateServiceKey(ctx, serviceName, serviceKey) {
			logger.WarnContext(ctx, "Service auth middleware: Invalid service key",
				"operation", "require_service_auth",
				"method", r.Method,
				"path", r.URL.Path,
				"service_name", serviceName)
			httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
			return
		}

		// Create service info for context
		serviceInfo := &ServiceInfo{
			Name: serviceName,
		}

		// Add service info to context
		ctx = context.WithValue(ctx, serviceContextKey, serviceInfo)
		r = r.WithContext(ctx)

		logger.InfoContext(ctx, "Service auth middleware: Service authentication successful",
			"operation", "require_service_auth",
			"method", r.Method,
			"path", r.URL.Path,
			"service_name", serviceName)

		// Continue to next handler
		next(w, r)
	}
}

// validateServiceKey validates the service key against the stored service key hash.
// The key hash is cached in memory with a configurable TTL to avoid a store round-trip on every request.
// Without a store, every key is rejected.
func (m *SessionAuthMiddleware) validateServiceKey(ctx context.Context, serviceName, serviceKey string) bool {
	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		return false
	}

	if m.secretReader == nil {
		logger.WarnContext(ctx, "Service auth middleware: No service key store configured, rejecting",
			"operation", "validate_service_key",
			"service_name", serviceName)
		return false
	}

	// Hash the provided key once for comparison.
	providedKeyBytes, err := encx.SerializeValue(serviceKey)
	if err != nil {
		logger.ErrorContext(ctx, "Service auth middleware: Failed to serialize service key",
			"operation", "validate_service_key",
			"service_name", serviceName,
			"error", err)
		return false
	}
	providedKeyHash := m.crypto.HashBasic(ctx, providedKeyBytes)

	// Check the cache first.
	storedKeyHash, ok := m.getCachedKeyHash(serviceName)
	if ok {
		// Cache hit (within TTL) – compare directly, no store round-trip.
		isValid := storedKeyHash == providedKeyHash
		if isValid {
			logger.InfoContext(ctx, "Service auth middleware: Service key validation successful (cached)",
				"operation", "validate_service_key",
				"service_name", serviceName)
		} else {
			logger.WarnContext(ctx, "Service auth middleware: Service key validation failed (cached)",
				"operation", "validate_service_key",
				"service_name", serviceName)
		}
		return isValid
	}

	// Cache miss or expired – fetch from the store.
	keyPath := services.ServiceAPIKeyPath(serviceName)

	logger.InfoContext(ctx, "Service auth middleware: Fetching service key from store",
		"operation", "validate_service_key",
		"service_name", serviceName,
		"key_path", keyPath)

	secretData, err := m.secretReader.Read(ctx, keyPath)
	if err != nil {
		logger.ErrorContext(ctx, "Service auth middleware: Failed to read service key from store",
			"operation", "validate_service_key",
			"service_name", serviceName,
			"key_path", keyPath,
			"error", err)
		return false
	}

	if secretData == nil || secretData.Data == nil {
		logger.WarnContext(ctx, "Service auth middleware: Service key not found in store",
			"operation", "validate_service_key",
			"service_name", serviceName,
			"key_path", keyPath)
		return false
	}

	// Entries follow the KV v2 layout: data nested under the "data" key.
	data, ok := secretData.Data["data"].(map[string]interface{})
	if !ok {
		logger.ErrorContext(ctx, "Service auth middleware: Invalid service key entry format",
			"operation", "validate_service_key",
			"service_name", serviceName,
			"key_path", keyPath)
		return false
	}

	fetchedKeyHash, ok := data["key_hash"].(string)
	if !ok || fetchedKeyHash == "" {
		logger.WarnContext(ctx, "Service auth middleware: Missing or invalid key_hash in store",
			"operation", "validate_service_key",
			"service_name", serviceName,
			"key_path", keyPath)
		return false
	}

	// Update the cache.
	m.setCachedKeyHash(serviceName, fetchedKeyHash)

	isValid := fetchedKeyHash == providedKeyHash

	if isValid {
		logger.InfoContext(ctx, "Service auth middleware: Service key validation successful",
			"operation", "validate_service_key",
			"service_name", serviceName)
	} else {
		logger.WarnContext(ctx, "Service auth middleware: Service key validation failed",
			"operation", "validate_service_key",
			"service_name", serviceName)
	}

	return isValid
}

// getCachedKeyHash returns the cached key hash for the given service if it exists and is not expired.
func (m *SessionAuthMiddleware) getCachedKeyHash(serviceName string) (string, bool) {
	m.cacheMu.RLock()
	defer m.cacheMu.RUnlock()

	entry, exists := m.serviceKeyCache[serviceName]
	if !exists {
		return "", false
	}

	if time.Since(entry.fetchedAt) > m.cacheTTL {
		return "", false
	}

	return entry.keyHash, true
}

// setCachedKeyHash stores the key hash for the given service in the cache.
func (m *SessionAuthMiddleware) setCachedKeyHash(serviceName, keyHash string) {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()

	m.serviceKeyCache[serviceName] = &cachedKeyEntry{
		keyHash:   keyHash,
		fetchedAt: time.Now(),
	}
}

// GetServiceInfoFromContext extracts service info from request context
func GetServiceInfoFromContext(ctx context.Context) (*ServiceInfo, error) {
	serviceInfo, ok := ctx.Value(serviceContextKey).(*ServiceInfo)
	if !ok {
		return nil, errs.NewInvalidValueErr("service info not found in context")
	}
	return serviceInfo, nil
}
