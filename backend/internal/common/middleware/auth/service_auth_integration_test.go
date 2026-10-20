package auth_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/services"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware/auth"
	"github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	"github.com/hengadev/encx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withLogger attaches a logger to the request context, as the app's logger middleware does;
// the auth middleware answers 500 without one.
func withLogger(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxutil.LoggerKey, slog.Default()))
}

// TestServiceAuthWithKeyStore tests the complete service authentication flow against a
// service key store holding the key hash
func TestServiceAuthWithKeyStore(t *testing.T) {
	ctx := context.Background()

	// Initialize encx crypto service for testing
	cryptoService, err := encx.NewTestCrypto(t)
	require.NoError(t, err, "Failed to create test crypto service")

	// Initialize the service key store
	keyStore := testutils.NewServiceKeyStore(cryptoService)

	// Generate and store a service API key
	testServiceName := services.Catalog
	apiKey, err := keyStore.GenerateAndStoreServiceKey(ctx, testServiceName)
	require.NoError(t, err, "Failed to generate service key")
	require.NotEmpty(t, apiKey, "Generated API key should not be empty")

	// Create auth middleware reading the key store
	middleware := auth.NewSessionAuthMiddleware(nil, cryptoService, keyStore)

	// Test successful authentication
	t.Run("Valid service authentication", func(t *testing.T) {
		// Create request with valid service headers
		req := httptest.NewRequest("GET", "/internal/test", nil)
		req.Header.Set(services.ServiceNameHeader, testServiceName)
		req.Header.Set(services.ServiceKeyHeader, apiKey)

		req = withLogger(req)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Track if next handler was called
		nextCalled := false
		var capturedServiceInfo *auth.ServiceInfo
		nextHandler := func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true

			// Extract service info from context
			serviceInfo, err := auth.GetServiceInfoFromContext(r.Context())
			if err == nil {
				capturedServiceInfo = serviceInfo
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		}

		// Execute middleware
		handler := middleware.RequireServiceAuth(nextHandler)
		handler(rr, req)

		// Assertions
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, nextCalled, "Next handler should have been called")
		assert.NotNil(t, capturedServiceInfo, "Service info should be available in context")
		assert.Equal(t, testServiceName, capturedServiceInfo.Name)
	})

	// Test authentication with wrong key
	t.Run("Invalid service key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/internal/test", nil)
		req.Header.Set(services.ServiceNameHeader, testServiceName)
		req.Header.Set(services.ServiceKeyHeader, "invalid-key-12345")

		req = withLogger(req)

		rr := httptest.NewRecorder()

		nextCalled := false
		nextHandler := func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}

		handler := middleware.RequireServiceAuth(nextHandler)
		handler(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.False(t, nextCalled, "Next handler should not have been called")
	})

	// Test authentication with non-existent service
	t.Run("Non-existent service", func(t *testing.T) {
		nonExistentService := "nonexistent"

		req := httptest.NewRequest("GET", "/internal/test", nil)
		req.Header.Set(services.ServiceNameHeader, nonExistentService)
		req.Header.Set(services.ServiceKeyHeader, apiKey)

		req = withLogger(req)

		rr := httptest.NewRecorder()

		nextCalled := false
		nextHandler := func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}

		handler := middleware.RequireServiceAuth(nextHandler)
		handler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.False(t, nextCalled, "Next handler should not have been called")
	})
}

// TestEndToEndServiceCommunication tests complete service-to-service communication
func TestEndToEndServiceCommunication(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cryptoService, err := encx.NewTestCrypto(t)
	require.NoError(t, err)

	// Generate service keys
	keyStore := testutils.NewServiceKeyStore(cryptoService)
	catalogKey, err := keyStore.GenerateAndStoreServiceKey(ctx, services.Catalog)
	require.NoError(t, err)

	// Create auth middleware
	middleware := auth.NewSessionAuthMiddleware(nil, cryptoService, keyStore)

	// Create a test server with protected endpoint
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		serviceInfo, err := auth.GetServiceInfoFromContext(r.Context())
		if err != nil {
			http.Error(w, "Service info not found", http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"message":  "Hello from protected endpoint",
			"service":  serviceInfo.Name,
			"endpoint": r.URL.Path,
			"method":   r.Method,
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"%s","service":"%s","endpoint":"%s","method":"%s"}`,
			response["message"], response["service"], response["endpoint"], response["method"])
	}

	// Wrap handler with service auth middleware
	protectedHandler := func(w http.ResponseWriter, r *http.Request) {
		// Apply service auth middleware
		middleware.RequireServiceAuth(testHandler)(w, withLogger(r))
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(protectedHandler))
	defer server.Close()

	// Test successful service call
	t.Run("Successful service-to-service call", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/internal/test", nil)
		require.NoError(t, err)

		// Add service authentication headers
		req.Header.Set(services.ServiceNameHeader, services.Catalog)
		req.Header.Set(services.ServiceKeyHeader, catalogKey)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Note: In a real test, you'd decode the JSON response here
		// For this example, we'll just verify the status code
	})

	// Test failed authentication
	t.Run("Failed authentication with invalid key", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/internal/test", nil)
		require.NoError(t, err)

		// Add invalid service authentication headers
		req.Header.Set(services.ServiceNameHeader, services.Catalog)
		req.Header.Set(services.ServiceKeyHeader, "invalid-key")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
