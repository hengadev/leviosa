package settings

import (
	"context"
	"fmt"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

func (s *SettingsService) GetCompanyRegistrationNumber(ctx context.Context) (*domain.GetCompanyRegistrationNumberResponse, error) {
	setting, err := s.repo.GetString(ctx, settings.CompanyRegistrationNumber)
	if err != nil {
		return nil, fmt.Errorf("get company registration number: %w", err)
	}
	return &domain.GetCompanyRegistrationNumberResponse{RegistrationNumber: setting.Value}, nil
}

func (s *SettingsService) SetCompanyRegistrationNumber(ctx context.Context, request *domain.SetCompanyRegistrationNumberRequest) (*domain.SetCompanyRegistrationNumberResponse, error) {
	if err := request.Valid(ctx); err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}

	trimmed := strings.TrimSpace(request.RegistrationNumber)

	setting := NewSetting(settings.CompanyRegistrationNumber, trimmed)
	if err := s.repo.SetString(ctx, setting); err != nil {
		return nil, fmt.Errorf("set company registration number: %w", err)
	}

	if err := s.publisher.PublishSettingUpdate(ctx, settings.CompanyRegistrationNumber, trimmed); err != nil {
		return nil, fmt.Errorf("publish company registration number update: %w", err)
	}

	return &domain.SetCompanyRegistrationNumberResponse{Success: true}, nil
}
