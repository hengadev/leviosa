package settings

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

// Practice access (docs/issues/023-practice-access-settings.md): the door
// code and access instructions of the practice. Both are private: stored
// encrypted with encx, as the company phone is, and served only to
// administrators. An unset value reads as empty. Clearing stores an encrypted
// empty string (encx serializes "" as a 4-byte length prefix, so the stored
// ciphertext is never empty and the old value never survives). Every read goes
// to storage: no cache. Neither value ever appears in an error, a log line or
// an audit payload.

func (s *SettingsService) GetPracticeAccess(ctx context.Context) (*domain.GetPracticeAccessResponse, error) {
	code, err := s.encryptedOrEmpty(ctx, settings.PracticeAccessDoorCode, "practice access door code")
	if err != nil {
		return nil, err
	}
	instructions, err := s.encryptedOrEmpty(ctx, settings.PracticeAccessInstructions, "practice access instructions")
	if err != nil {
		return nil, err
	}
	return &domain.GetPracticeAccessResponse{DoorCode: code, AccessInstructions: instructions}, nil
}

// SetPracticeAccess trims and validates both values, then stores and audits
// each one that differs from the stored value. CodeChanged tells whether the
// door code changed.
func (s *SettingsService) SetPracticeAccess(ctx context.Context, request *domain.SetPracticeAccessRequest) (*domain.SetPracticeAccessResponse, error) {
	if err := request.Valid(ctx); err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}

	current, err := s.GetPracticeAccess(ctx)
	if err != nil {
		return nil, err
	}

	code := strings.TrimSpace(request.DoorCode)
	instructions := strings.TrimSpace(request.AccessInstructions)

	codeChanged := code != current.DoorCode
	if codeChanged {
		if err := s.setEncrypted(ctx, settings.PracticeAccessDoorCode, code, "practice access door code"); err != nil {
			return nil, err
		}
	}
	if instructions != current.AccessInstructions {
		if err := s.setEncrypted(ctx, settings.PracticeAccessInstructions, instructions, "practice access instructions"); err != nil {
			return nil, err
		}
	}

	return &domain.SetPracticeAccessResponse{Success: true, CodeChanged: codeChanged}, nil
}

// encryptedOrEmpty reads and decrypts an encrypted setting, answering "" when
// it is unset. resource names the setting in errors.
func (s *SettingsService) encryptedOrEmpty(ctx context.Context, key, resource string) (string, error) {
	settingEncx, err := s.repo.GetEncryptedSetting(ctx, key)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("get %s: %w", resource, err)
	}
	setting, err := domain.DecryptSettingEncryptedEncx(ctx, s.crypto, settingEncx)
	if err != nil {
		return "", errs.NewNotDecryptedErr(resource, err)
	}
	return setting.Value, nil
}

// setEncrypted encrypts and stores one setting, then audits the change with
// the value redacted.
func (s *SettingsService) setEncrypted(ctx context.Context, key, value, resource string) error {
	settingEncx, err := domain.ProcessSettingEncryptedEncx(ctx, s.crypto, &domain.SettingEncrypted{Key: key, Value: value})
	if err != nil {
		return errs.NewNotEncryptedErr(resource, err)
	}
	if err := s.repo.SetEncryptedSetting(ctx, settingEncx); err != nil {
		return fmt.Errorf("set %s: %w", resource, err)
	}
	s.auditFieldUpdate(ctx, key, "[redacted]")
	return nil
}
