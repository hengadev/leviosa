package settings

import (
	"context"
	"fmt"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

func (s *SettingsService) GetCompanyVatNumber(ctx context.Context) (*domain.GetCompanyVatNumberResponse, error) {
	setting, err := s.repo.GetString(ctx, settings.CompanyVatNumber)
	if err != nil {
		return nil, fmt.Errorf("get company VAT number: %w", err)
	}
	return &domain.GetCompanyVatNumberResponse{VatNumber: setting.Value}, nil
}

// SetCompanyVatNumber persists the company's VAT number. An empty value is
// accepted without error: businesses under the French "franchise en base de
// TVA" regime are not VAT-registered and have none.
func (s *SettingsService) SetCompanyVatNumber(ctx context.Context, request *domain.SetCompanyVatNumberRequest) (*domain.SetCompanyVatNumberResponse, error) {
	if err := request.Valid(ctx); err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}

	trimmed := strings.TrimSpace(request.VatNumber)

	setting := NewSetting(settings.CompanyVatNumber, trimmed)
	if err := s.repo.SetString(ctx, setting); err != nil {
		return nil, fmt.Errorf("set company VAT number: %w", err)
	}

	if err := s.publisher.PublishSettingUpdate(ctx, settings.CompanyVatNumber, trimmed); err != nil {
		return nil, fmt.Errorf("publish company VAT number update: %w", err)
	}

	return &domain.SetCompanyVatNumberResponse{Success: true}, nil
}
