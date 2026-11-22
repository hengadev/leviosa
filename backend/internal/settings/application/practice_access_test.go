package settings_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	settingsApp "github.com/Leviosa-care/leviosa/backend/internal/settings/application"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/ports"

	"github.com/hengadev/encx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// memoryEncryptedRepository keeps encrypted settings in memory, so the
// practice access tests exercise the real encx round trip.
type memoryEncryptedRepository struct {
	mockSettingsRepository
	mu     sync.Mutex
	stored map[string]*domain.SettingEncryptedEncx
	writes []string
}

func newMemoryEncryptedRepository() *memoryEncryptedRepository {
	return &memoryEncryptedRepository{stored: map[string]*domain.SettingEncryptedEncx{}}
}

func (m *memoryEncryptedRepository) GetEncryptedSetting(ctx context.Context, key string) (*domain.SettingEncryptedEncx, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.stored[key]
	if !ok {
		return nil, errs.NewRepositoryNotFoundErr(errors.New("no rows"), key)
	}
	return s, nil
}

func (m *memoryEncryptedRepository) SetEncryptedSetting(ctx context.Context, s *domain.SettingEncryptedEncx) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stored[s.Key] = s
	m.writes = append(m.writes, s.Key)
	return nil
}

var _ ports.SettingsRepository = (*memoryEncryptedRepository)(nil)

// recordingAudit captures the audit entries written.
type recordingAudit struct {
	entries []audit.Entry
}

func (a *recordingAudit) Write(ctx context.Context, entry audit.Entry) error {
	a.entries = append(a.entries, entry)
	return nil
}

func newPracticeAccessService(t *testing.T) (ports.SettingsService, *memoryEncryptedRepository, *recordingAudit) {
	t.Helper()
	crypto, err := encx.NewTestCrypto(t)
	require.NoError(t, err)
	repo := newMemoryEncryptedRepository()
	auditLog := &recordingAudit{}
	return settingsApp.New(repo, nil, crypto, &mockEventPublisher{}, auditLog), repo, auditLog
}

func TestGetPracticeAccess(t *testing.T) {
	ctx := context.Background()

	t.Run("answers empty values while unset", func(t *testing.T) {
		svc, _, _ := newPracticeAccessService(t)

		resp, err := svc.GetPracticeAccess(ctx)
		require.NoError(t, err)
		assert.Empty(t, resp.DoorCode)
		assert.Empty(t, resp.AccessInstructions)
	})

	t.Run("propagates a repository failure without the values", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetEncryptedSetting", ctx, settings.PracticeAccessDoorCode).Return(nil, errors.New("connection refused"))
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.GetPracticeAccess(ctx)
		assert.Error(t, err)
	})
}

func TestSetPracticeAccess(t *testing.T) {
	ctx := context.Background()

	t.Run("round trips trimmed values and keeps line breaks", func(t *testing.T) {
		svc, _, _ := newPracticeAccessService(t)

		resp, err := svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{
			DoorCode:           "  4521B ",
			AccessInstructions: "\n Bâtiment B, 2e étage.\nSonner à Leviosa. \n",
		})
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.True(t, resp.CodeChanged)

		got, err := svc.GetPracticeAccess(ctx)
		require.NoError(t, err)
		assert.Equal(t, "4521B", got.DoorCode)
		assert.Equal(t, "Bâtiment B, 2e étage.\nSonner à Leviosa.", got.AccessInstructions)
	})

	t.Run("stores the encrypted value, never the plain text", func(t *testing.T) {
		svc, repo, _ := newPracticeAccessService(t)

		_, err := svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{
			DoorCode:           "4521B",
			AccessInstructions: "Bâtiment B",
		})
		require.NoError(t, err)

		for key, plain := range map[string]string{
			settings.PracticeAccessDoorCode:     "4521B",
			settings.PracticeAccessInstructions: "Bâtiment B",
		} {
			stored := repo.stored[key]
			require.NotNil(t, stored, key)
			assert.NotEmpty(t, stored.ValueEncrypted, key)
			assert.NotEmpty(t, stored.DEKEncrypted, key)
			assert.False(t, bytes.Contains(stored.ValueEncrypted, []byte(plain)), "%s holds the plain text", key)
		}
	})

	t.Run("answers code_changed only when the door code differs", func(t *testing.T) {
		svc, repo, _ := newPracticeAccessService(t)

		first, err := svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{DoorCode: "1234", AccessInstructions: "A"})
		require.NoError(t, err)
		assert.True(t, first.CodeChanged, "unset to a code is a change")

		same, err := svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{DoorCode: " 1234 ", AccessInstructions: "B"})
		require.NoError(t, err)
		assert.False(t, same.CodeChanged, "the same code after trimming is not a change")

		changed, err := svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{DoorCode: "5678", AccessInstructions: "B"})
		require.NoError(t, err)
		assert.True(t, changed.CodeChanged)

		got, err := svc.GetPracticeAccess(ctx)
		require.NoError(t, err)
		assert.Equal(t, "5678", got.DoorCode)
		assert.Equal(t, "B", got.AccessInstructions)
		assert.Equal(t, []string{
			settings.PracticeAccessDoorCode, settings.PracticeAccessInstructions, // first
			settings.PracticeAccessInstructions, // same code, new instructions
			settings.PracticeAccessDoorCode,     // new code, same instructions
		}, repo.writes, "only changed fields are written")
	})

	t.Run("clears both values", func(t *testing.T) {
		svc, repo, _ := newPracticeAccessService(t)

		_, err := svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{DoorCode: "1234", AccessInstructions: "Bâtiment B"})
		require.NoError(t, err)

		resp, err := svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{DoorCode: "  ", AccessInstructions: ""})
		require.NoError(t, err)
		assert.True(t, resp.CodeChanged, "clearing the code is a change")

		got, err := svc.GetPracticeAccess(ctx)
		require.NoError(t, err)
		assert.Empty(t, got.DoorCode)
		assert.Empty(t, got.AccessInstructions)
		// encx encrypts the empty string: the stored record is a real
		// ciphertext, not a leftover of the old value.
		assert.NotEmpty(t, repo.stored[settings.PracticeAccessDoorCode].ValueEncrypted)
	})

	t.Run("refuses values over the limits naming the field, not the value", func(t *testing.T) {
		longCode := strings.Repeat("7", domain.MaxPracticeAccessDoorCodeLength+1)
		longInstructions := strings.Repeat("x", domain.MaxPracticeAccessInstructionsLength) + "SECRETTAIL"

		for name, tc := range map[string]struct {
			request domain.SetPracticeAccessRequest
			field   string
			value   string
		}{
			"door code":    {domain.SetPracticeAccessRequest{DoorCode: longCode}, "door code", longCode},
			"instructions": {domain.SetPracticeAccessRequest{AccessInstructions: longInstructions}, "access instructions", "SECRETTAIL"},
			"control char": {domain.SetPracticeAccessRequest{DoorCode: "12\x0034"}, "door code", "12\x0034"},
		} {
			t.Run(name, func(t *testing.T) {
				svc, repo, auditLog := newPracticeAccessService(t)

				_, err := svc.SetPracticeAccess(ctx, &tc.request)
				require.Error(t, err)
				assert.ErrorIs(t, err, errs.ErrInvalidValue)
				assert.Contains(t, err.Error(), tc.field)
				assert.NotContains(t, err.Error(), tc.value)
				assert.Empty(t, repo.writes)
				assert.Empty(t, auditLog.entries)
			})
		}
	})

	t.Run("accepts the limits exactly, counting characters", func(t *testing.T) {
		svc, _, _ := newPracticeAccessService(t)

		_, err := svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{
			DoorCode:           strings.Repeat("é", domain.MaxPracticeAccessDoorCodeLength),
			AccessInstructions: strings.Repeat("é", domain.MaxPracticeAccessInstructionsLength),
		})
		assert.NoError(t, err)
	})

	t.Run("audits each changed field with the value redacted", func(t *testing.T) {
		svc, _, auditLog := newPracticeAccessService(t)

		_, err := svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{DoorCode: "4521B", AccessInstructions: "Bâtiment B"})
		require.NoError(t, err)
		_, err = svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{DoorCode: "9876", AccessInstructions: "Bâtiment B"})
		require.NoError(t, err)

		require.Len(t, auditLog.entries, 3, "two fields, then the code only")
		fields := []string{}
		for _, e := range auditLog.entries {
			assert.Equal(t, "settings.field_updated", e.Action)
			assert.Equal(t, "[redacted]", e.Payload["new_value"])
			for _, v := range e.Payload {
				assert.NotContains(t, v, "4521B")
				assert.NotContains(t, v, "9876")
				assert.NotContains(t, v, "Bâtiment")
			}
			fields = append(fields, e.TargetID)
		}
		assert.Equal(t, []string{
			settings.PracticeAccessDoorCode, settings.PracticeAccessInstructions, settings.PracticeAccessDoorCode,
		}, fields)
	})

	t.Run("propagates a storage failure", func(t *testing.T) {
		crypto, err := encx.NewTestCrypto(t)
		require.NoError(t, err)
		notFound := errs.NewRepositoryNotFoundErr(errors.New("no rows"), "encrypted setting")
		repo := &mockSettingsRepository{}
		repo.On("GetEncryptedSetting", ctx, mock.Anything).Return(nil, notFound)
		repo.On("SetEncryptedSetting", ctx, mock.Anything).Return(errors.New("connection refused"))
		svc := settingsApp.New(repo, nil, crypto, &mockEventPublisher{}, nil)

		_, err = svc.SetPracticeAccess(ctx, &domain.SetPracticeAccessRequest{DoorCode: "4521B"})
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "4521B")
	})
}
