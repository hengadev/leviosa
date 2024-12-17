package sessionService_test

import (
	"context"
)

type MockRepo struct {
	FindSessionByIDFunc func(ctx context.Context, sessionID string) ([]byte, error)
	CreateSessionFunc   func(ctx context.Context, sessionID string, sessionEncoded []byte) error
	RemoveSessionFunc   func(ctx context.Context, sessionID string) error
}

func (m *MockRepo) FindSessionByID(ctx context.Context, sessionID string) ([]byte, error) {
	if m.FindSessionByIDFunc != nil {
		return m.FindSessionByIDFunc(ctx, sessionID)
	}
	return nil, nil
}

func (m *MockRepo) CreateSession(ctx context.Context, sessionID string, sessionEncoded []byte) error {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, sessionID, sessionEncoded)
	}
	return nil
}

func (m *MockRepo) RemoveSession(ctx context.Context, sessionID string) error {
	if m.RemoveSessionFunc != nil {
		return m.RemoveSessionFunc(ctx, sessionID)
	}
	return nil
}
