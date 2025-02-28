package sessionService_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/hengadev/leviosa/internal/domain"
	"github.com/hengadev/leviosa/internal/domain/session"
	rp "github.com/hengadev/leviosa/internal/repository"

	"github.com/hengadev/test-assert"
)

func TestGetSession(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		mockRepo      func() *MockRepo
		expectedError error
		expectedValue *sessionService.Session
	}{
		{name: "empty string ID", sessionID: "", mockRepo: func() *MockRepo { return &MockRepo{} }, expectedValue: nil, expectedError: domain.ErrNotFound},
		{name: "session ID not found", sessionID: "nonexistent", mockRepo: func() *MockRepo {
			return &MockRepo{FindSessionByIDFunc: func(ctx context.Context, sessionID string) ([]byte, error) { return nil, rp.ErrNotFound }}
		}, expectedValue: nil, expectedError: domain.ErrNotFound},
		{name: "database error", sessionID: "session123", mockRepo: func() *MockRepo {
			return &MockRepo{FindSessionByIDFunc: func(ctx context.Context, sessionID string) ([]byte, error) { return nil, rp.ErrDatabase }}
		}, expectedValue: nil, expectedError: domain.ErrQueryFailed},
		{name: "unexpected error", sessionID: "session123", mockRepo: func() *MockRepo {
			return &MockRepo{FindSessionByIDFunc: func(ctx context.Context, sessionID string) ([]byte, error) {
				return nil, errors.New("unexpected error")
			}}
		}, expectedValue: nil, expectedError: domain.ErrUnexpectedType},
		{name: "JSON unmarshal error", sessionID: "session123", mockRepo: func() *MockRepo {
			return &MockRepo{FindSessionByIDFunc: func(ctx context.Context, sessionID string) ([]byte, error) { return nil, domain.ErrUnmarshalJSON }}
		}, expectedValue: nil, expectedError: domain.ErrUnmarshalJSON},
		{name: "successful case", sessionID: "session123", mockRepo: func() *MockRepo {
			return &MockRepo{
				FindSessionByIDFunc: func(ctx context.Context, sessionID string) ([]byte, error) {
					session := &sessionService.Session{ID: "session123", UserID: "user123", LoggedInAt: time.Now(), CreatedAt: time.Now(), ExpiresAt: time.Now().Add(sessionService.SessionDuration)}
					data, _ := json.Marshal(session)
					return data, domain.ErrUnmarshalJSON
				},
			}
		}, expectedValue: nil, expectedError: domain.ErrUnmarshalJSON,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockRepo := tt.mockRepo()
			service := sessionService.New(mockRepo)
			session, err := service.GetSession(ctx, tt.sessionID)
			assert.EqualError(t, err, tt.expectedError)
			if tt.expectedValue == nil {
				assert.Equal(t, session, tt.expectedValue)
			} else {
				assert.ReflectEqual(t, session, tt.expectedValue)
			}
		})
	}
}
