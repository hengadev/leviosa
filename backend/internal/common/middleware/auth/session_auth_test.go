package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/cookies"
	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
	"github.com/hengadev/encx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewSessionAuthMiddleware(t *testing.T) {
	mockRepo := &session.MockSessionRepository{}
	mockCrypto, err := NewTestCrypto(t)
	require.NoError(t, err, "Failed to create test crypto service")

	middleware := NewSessionAuthMiddleware(mockRepo, mockCrypto, nil)

	assert.NotNil(t, middleware)
	assert.IsType(t, &SessionAuthMiddleware{}, middleware)
}

func TestCookieExtraction(t *testing.T) {
	// Simple test to verify cookie extraction works
	req := newTestRequest("GET", "/test")
	req.AddCookie(&http.Cookie{
		Name:  cookies.AccessTokenCookieName,
		Value: "test_token",
	})

	cookie, err := req.Cookie(cookies.AccessTokenCookieName)
	require.NoError(t, err)
	assert.Equal(t, "test_token", cookie.Value)
}

func TestRequireAccessToken(t *testing.T) {
	crypto, err := NewTestCrypto(t)
	require.NoError(t, err, "Failed to create test crypto service")

	tests := []struct {
		name           string
		setupCookie    func(*http.Request)
		repoResponse   []byte
		repoError      error
		expectedStatus int
		expectedInCtx  bool
	}{
		{
			name:           "missing access token cookie",
			setupCookie:    func(req *http.Request) {},
			expectedStatus: http.StatusUnauthorized,
			expectedInCtx:  false,
		},
		{
			name: "empty access token cookie",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.AccessTokenCookieName,
					Value: "",
				})
			},
			expectedStatus: http.StatusUnauthorized,
			expectedInCtx:  false,
		},
		{
			name: "valid access token - session not found",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.AccessTokenCookieName,
					Value: "valid_token",
				})
			},
			repoError:      errs.ErrRepositoryNotFound,
			expectedStatus: http.StatusUnauthorized,
			expectedInCtx:  false,
		},
		{
			name: "valid access token - database error",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.AccessTokenCookieName,
					Value: "valid_token",
				})
			},
			repoError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedInCtx:  false,
		},
		{
			name: "valid access token - invalid session JSON",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.AccessTokenCookieName,
					Value: "valid_token",
				})
			},
			repoResponse:   []byte("invalid json"),
			expectedStatus: http.StatusInternalServerError,
			expectedInCtx:  false,
		},
		{
			name: "valid access token - pending session (should work)",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.AccessTokenCookieName,
					Value: "valid_token",
				})
			},
			repoResponse: func() []byte {
				session := &session.Session{
					ID:     uuid.New(),
					UserID: uuid.New(),
					Role:   identity.Standard,
					State:  session.SessionPending,
				}
				return createValidSessionJSON(t, crypto, session)
			}(),
			expectedStatus: http.StatusOK,
			expectedInCtx:  true,
		},
		// {
		// 	name: "valid access token - active session",
		// 	setupCookie: func(req *http.Request) {
		// 		req.AddCookie(&http.Cookie{
		// 			Name:  cookies.AccessTokenCookieName,
		// 			Value: "valid_token",
		// 		})
		// 	},
		// 	repoResponse: func() []byte {
		// 		session := &session.Session{
		// 			ID:     uuid.New(),
		// 			UserID: uuid.New(),
		// 			Role:   identity.Standard,
		// 			State:  session.SessionActive,
		// 		}
		// 		return createValidSessionJSON(t, crypto, session)
		// 	}(),
		// 	expectedStatus: http.StatusOK,
		// 	expectedInCtx:  true,
		// },
		{
			name: "valid access token - expired session (invalid state)",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.AccessTokenCookieName,
					Value: "valid_token",
				})
			},
			repoResponse: func() []byte {
				session := &session.Session{
					ID:     uuid.New(),
					UserID: uuid.New(),
					Role:   identity.Standard,
					State:  "expired", // Invalid state
				}
				return createValidSessionJSON(t, crypto, session)
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedInCtx:  false,
		},
		{
			name: "valid access token - expired session (time-based expiration)",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.AccessTokenCookieName,
					Value: "valid_token",
				})
			},
			repoResponse: func() []byte {
				session := &session.Session{
					ID:     uuid.New(),
					UserID: uuid.New(),
					Role:   identity.Standard,
					State:  session.SessionActive, // Valid state but expired
				}
				return createExpiredSessionJSON(t, crypto, session, time.Now().Add(-time.Hour))
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedInCtx:  false,
		},
		{
			name: "valid access token - pending session expired (time-based expiration)",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.AccessTokenCookieName,
					Value: "valid_token",
				})
			},
			repoResponse: func() []byte {
				session := &session.Session{
					ID:     uuid.New(),
					UserID: uuid.New(),
					Role:   identity.Standard,
					State:  session.SessionPending, // Valid state but expired
				}
				return createExpiredSessionJSON(t, crypto, session, time.Now().Add(-time.Minute))
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedInCtx:  false,
		},
		{
			name: "valid access token - active session not expired",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.AccessTokenCookieName,
					Value: "valid_token",
				})
			},
			repoResponse: func() []byte {
				session := &session.Session{
					ID:     uuid.New(),
					UserID: uuid.New(),
					Role:   identity.Standard,
					State:  session.SessionActive,
				}
				return createValidSessionJSON(t, crypto, session) // Uses future expiration
			}(),
			expectedStatus: http.StatusOK,
			expectedInCtx:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &session.MockSessionRepository{}
			middleware := NewSessionAuthMiddleware(mockRepo, crypto, nil)

			if tt.repoResponse != nil || tt.repoError != nil {
				mockRepo.On("FindSessionByAccessTokenHash", mock.Anything, mock.AnythingOfType("string")).Return(uuid.NewString(), tt.repoResponse, tt.repoError)
			}

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				session, found := session.SessionInfoFromContext(r.Context())
				if tt.expectedInCtx {
					assert.True(t, found, "Expected session in context")
					assert.NotNil(t, session, "Expected non-nil session")
				} else {
					assert.False(t, found, "Expected no session in context")
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			req := newTestRequest("GET", "/test")
			tt.setupCookie(req)
			rr := httptest.NewRecorder()

			// Execute
			handler := middleware.RequireAccessToken(testHandler)
			// handler.ServeHTTP(rr, req)
			handler(rr, req)

			// Verify
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRequireRefreshToken(t *testing.T) {
	crypto, err := NewTestCrypto(t)
	require.NoError(t, err, "Failed to create test crypto service")

	tests := []struct {
		name           string
		path           string
		setupCookie    func(*http.Request)
		repoResponse   []byte
		repoError      error
		expectedStatus int
		expectedInCtx  bool
	}{
		{
			name:           "non-refresh endpoint should be forbidden",
			path:           "/api/users",
			setupCookie:    func(req *http.Request) {},
			expectedStatus: http.StatusForbidden,
			expectedInCtx:  false,
		},
		{
			name:           "refresh endpoint - missing refresh token cookie",
			path:           cookies.RefreshEndpoint,
			setupCookie:    func(req *http.Request) {},
			expectedStatus: http.StatusUnauthorized,
			expectedInCtx:  false,
		},
		{
			name: "refresh endpoint - empty refresh token cookie",
			path: cookies.RefreshEndpoint,
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.RefreshTokenCookieName,
					Value: "",
				})
			},
			expectedStatus: http.StatusUnauthorized,
			expectedInCtx:  false,
		},
		{
			name: "refresh endpoint - valid refresh token - session not found",
			path: cookies.RefreshEndpoint,
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.RefreshTokenCookieName,
					Value: "valid_refresh_token",
				})
			},
			repoError:      errs.ErrRepositoryNotFound,
			expectedStatus: http.StatusUnauthorized,
			expectedInCtx:  false,
		},
		{
			name: "refresh endpoint - valid refresh token - database error",
			path: cookies.RefreshEndpoint,
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.RefreshTokenCookieName,
					Value: "valid_refresh_token",
				})
			},
			repoError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedInCtx:  false,
		},
		{
			name: "refresh endpoint - valid refresh token - active session",
			path: cookies.RefreshEndpoint,
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.RefreshTokenCookieName,
					Value: "valid_refresh_token",
				})
			},
			repoResponse: func() []byte {
				session := &session.Session{
					ID:     uuid.New(),
					UserID: uuid.New(),
					Role:   identity.Standard,
					State:  session.SessionActive,
				}
				return createValidSessionJSON(t, crypto, session)
			}(),
			expectedStatus: http.StatusOK,
			expectedInCtx:  true,
		},
		{
			name: "refresh endpoint - valid refresh token - pending session",
			path: cookies.RefreshEndpoint,
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  cookies.RefreshTokenCookieName,
					Value: "valid_refresh_token",
				})
			},
			repoResponse: func() []byte {
				session := &session.Session{
					ID:     uuid.New(),
					UserID: uuid.New(),
					Role:   identity.Standard,
					State:  session.SessionPending,
				}
				return createValidSessionJSON(t, crypto, session)
			}(),
			expectedStatus: http.StatusOK,
			expectedInCtx:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &session.MockSessionRepository{}
			middleware := NewSessionAuthMiddleware(mockRepo, crypto, nil)

			if tt.repoResponse != nil || tt.repoError != nil {
				mockRepo.On("FindSessionByRefreshTokenHash", mock.Anything, mock.AnythingOfType("string")).Return(uuid.NewString(), tt.repoResponse, tt.repoError)
			}

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				session, found := session.SessionInfoFromContext(r.Context())
				if tt.expectedInCtx {
					assert.True(t, found, "Expected session in context")
					assert.NotNil(t, session, "Expected non-nil session")
				} else {
					assert.False(t, found, "Expected no session in context")
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			req := newTestRequest("GET", tt.path)
			tt.setupCookie(req)
			rr := httptest.NewRecorder()

			// Execute
			handler := middleware.RequireRefreshToken(testHandler)
			// handler.ServeHTTP(rr, req)
			handler(rr, req)

			// Verify
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// newTestRequest builds a request whose context carries a logger, as the
// logging middleware provides in the app.
func newTestRequest(method, target string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	return req.WithContext(context.WithValue(req.Context(), ctxutil.LoggerKey, slog.Default()))
}

// createValidSessionJSON encrypts the session with crypto, as the session
// store does, and returns the JSON the repository holds. A zero ExpiresAt
// becomes one hour from now.
func createValidSessionJSON(t *testing.T, crypto encx.CryptoService, sessionStruct *session.Session) []byte {
	t.Helper()

	s := *sessionStruct
	if s.ExpiresAt.IsZero() {
		s.ExpiresAt = time.Now().Add(time.Hour)
	}
	sessionEncx, err := session.ProcessSessionEncx(context.Background(), crypto, &s)
	require.NoError(t, err)

	data, err := json.Marshal(sessionEncx)
	require.NoError(t, err)

	return data
}

// createExpiredSessionJSON is createValidSessionJSON with ExpiresAt set to expiresAt.
func createExpiredSessionJSON(t *testing.T, crypto encx.CryptoService, sessionStruct *session.Session, expiresAt time.Time) []byte {
	t.Helper()

	expired := *sessionStruct
	expired.ExpiresAt = expiresAt
	return createValidSessionJSON(t, crypto, &expired)
}
