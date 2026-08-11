package settings_test

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/ports"

	"github.com/stretchr/testify/mock"
)

// mockSettingsRepository is an in-memory test double for ports.SettingsRepository.
// Per ADR-0005, application-layer unit tests exercise pure business logic
// against a mocked repository; DB-backed behavior is covered separately by
// integration tests in test/integration/settings.
type mockSettingsRepository struct {
	mock.Mock
}

func (m *mockSettingsRepository) GetString(ctx context.Context, key string) (*domain.Setting[string], error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Setting[string]), args.Error(1)
}

func (m *mockSettingsRepository) GetInt(ctx context.Context, key string) (*domain.Setting[int], error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Setting[int]), args.Error(1)
}

func (m *mockSettingsRepository) SetString(ctx context.Context, setting *domain.Setting[string]) error {
	args := m.Called(ctx, setting)
	return args.Error(0)
}

func (m *mockSettingsRepository) SetInt(ctx context.Context, setting *domain.Setting[int]) error {
	args := m.Called(ctx, setting)
	return args.Error(0)
}

func (m *mockSettingsRepository) GetEncryptedSetting(ctx context.Context, key string) (*domain.SettingEncryptedEncx, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SettingEncryptedEncx), args.Error(1)
}

func (m *mockSettingsRepository) SetEncryptedSetting(ctx context.Context, settingEncx *domain.SettingEncryptedEncx) error {
	args := m.Called(ctx, settingEncx)
	return args.Error(0)
}

var _ ports.SettingsRepository = (*mockSettingsRepository)(nil)

// mockEventPublisher is an in-memory test double for ports.EventPublisher,
// used to assert that setting a value publishes an update through the
// existing publish mechanism.
type mockEventPublisher struct {
	mock.Mock
}

func (m *mockEventPublisher) PublishSettingUpdate(ctx context.Context, key string, value any) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

var _ ports.EventPublisher = (*mockEventPublisher)(nil)
