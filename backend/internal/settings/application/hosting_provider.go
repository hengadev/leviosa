package settings

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

func (s *SettingsService) GetCompanyHostingProvider(ctx context.Context) (*domain.GetCompanyHostingProviderResponse, error) {
	setting, err := s.repo.GetString(ctx, settings.CompanyHostingProvider)
	if err != nil {
		return nil, fmt.Errorf("get company hosting provider: %w", err)
	}
	return &domain.GetCompanyHostingProviderResponse{HostingProvider: setting.Value}, nil
}

func (s *SettingsService) SetCompanyHostingProvider(ctx context.Context, request *domain.SetCompanyHostingProviderRequest) (*domain.SetCompanyHostingProviderResponse, error) {
	if err := request.Valid(ctx); err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}

	setting := NewSetting(settings.CompanyHostingProvider, request.HostingProvider)
	if err := s.repo.SetString(ctx, setting); err != nil {
		return nil, fmt.Errorf("set company hosting provider: %w", err)
	}

	if err := s.publisher.PublishSettingUpdate(ctx, settings.CompanyHostingProvider, request.HostingProvider); err != nil {
		return nil, fmt.Errorf("publish company hosting provider update: %w", err)
	}

	return &domain.SetCompanyHostingProviderResponse{Success: true}, nil
}
