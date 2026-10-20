package auth

import (
	"sync"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"

	"github.com/hengadev/encx"
)

// DefaultServiceKeyCacheTTL is the default time-to-live for cached service key hashes.
const DefaultServiceKeyCacheTTL = 5 * time.Minute

// SessionAuthMiddleware implements AuthMiddleware using session repository
type SessionAuthMiddleware struct {
	sessionRepo session.SessionRepository
	crypto      encx.CryptoService

	// secretReader is the store of service API key hashes checked by RequireServiceAuth.
	// When nil, every service-authenticated request is rejected.
	secretReader SecretReader

	// serviceKeyCache stores key_hash values fetched from the secret reader per service.
	serviceKeyCache map[string]*cachedKeyEntry
	cacheTTL       time.Duration
	cacheMu        sync.RWMutex
}

// cachedKeyEntry is a single cached key-hash entry with its fetch timestamp.
type cachedKeyEntry struct {
	keyHash   string
	fetchedAt time.Time
}

// SessionAuthMiddlewareOption configures a SessionAuthMiddleware after creation.
type SessionAuthMiddlewareOption func(*SessionAuthMiddleware)

// WithServiceKeyCacheTTL sets the TTL for cached service key hashes.
func WithServiceKeyCacheTTL(ttl time.Duration) SessionAuthMiddlewareOption {
	return func(m *SessionAuthMiddleware) {
		m.cacheTTL = ttl
	}
}

// WithSecretReader sets a custom SecretReader (useful for testing).
func WithSecretReader(reader SecretReader) SessionAuthMiddlewareOption {
	return func(m *SessionAuthMiddleware) {
		m.secretReader = reader
	}
}

// NewSessionAuthMiddleware creates a new session-based auth middleware.
//
// serviceKeys is the store of service API key hashes that RequireServiceAuth checks. The
// app passes nil: its domains call each other in process, and the service keys lived in
// Vault, which is gone (ADR-0015), so service-to-service endpoints reject every request.
func NewSessionAuthMiddleware(sessionRepo session.SessionRepository, crypto encx.CryptoService, serviceKeys SecretReader, opts ...SessionAuthMiddlewareOption) AuthMiddleware {
	m := &SessionAuthMiddleware{
		sessionRepo:     sessionRepo,
		crypto:          crypto,
		secretReader:    serviceKeys,
		serviceKeyCache: make(map[string]*cachedKeyEntry),
		cacheTTL:        DefaultServiceKeyCacheTTL,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}
