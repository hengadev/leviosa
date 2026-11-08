package settings

import (
	"context"
	"fmt"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

// Practice page settings: which partner and product the public practice page
// shows (docs/issues/009-practice-page.md). The settings hold the IDs only;
// the page resolves the partner profile, product and price through the public
// catalog and partner endpoints, so changing either here changes the page
// without a release.
//
// None of these settings is guaranteed to be set: the getters surface the
// repository's not-found error so callers can treat "unset" distinctly.

func (s *SettingsService) GetPracticePagePartner(ctx context.Context) (*domain.GetPracticePagePartnerResponse, error) {
	setting, err := s.repo.GetString(ctx, settings.PracticePagePartnerID)
	if err != nil {
		return nil, fmt.Errorf("get practice page partner: %w", err)
	}
	return &domain.GetPracticePagePartnerResponse{PartnerID: setting.Value}, nil
}

func (s *SettingsService) SetPracticePagePartner(ctx context.Context, request *domain.SetPracticePagePartnerRequest) (*domain.SetPracticePagePartnerResponse, error) {
	if err := request.Valid(ctx); err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}

	setting := NewSetting(settings.PracticePagePartnerID, strings.TrimSpace(request.PartnerID))
	if err := s.repo.SetString(ctx, setting); err != nil {
		return nil, fmt.Errorf("set practice page partner: %w", err)
	}

	s.auditFieldUpdate(ctx, settings.PracticePagePartnerID, setting.Value)

	return &domain.SetPracticePagePartnerResponse{Success: true}, nil
}

func (s *SettingsService) GetPracticePageProduct(ctx context.Context) (*domain.GetPracticePageProductResponse, error) {
	setting, err := s.repo.GetString(ctx, settings.PracticePageProductID)
	if err != nil {
		return nil, fmt.Errorf("get practice page product: %w", err)
	}
	return &domain.GetPracticePageProductResponse{ProductID: setting.Value}, nil
}

func (s *SettingsService) SetPracticePageProduct(ctx context.Context, request *domain.SetPracticePageProductRequest) (*domain.SetPracticePageProductResponse, error) {
	if err := request.Valid(ctx); err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}

	setting := NewSetting(settings.PracticePageProductID, strings.TrimSpace(request.ProductID))
	if err := s.repo.SetString(ctx, setting); err != nil {
		return nil, fmt.Errorf("set practice page product: %w", err)
	}

	s.auditFieldUpdate(ctx, settings.PracticePageProductID, setting.Value)

	return &domain.SetPracticePageProductResponse{Success: true}, nil
}

func (s *SettingsService) GetPracticePageVideo(ctx context.Context) (*domain.GetPracticePageVideoResponse, error) {
	setting, err := s.repo.GetString(ctx, settings.PracticePageVideoURL)
	if err != nil {
		return nil, fmt.Errorf("get practice page video: %w", err)
	}
	return &domain.GetPracticePageVideoResponse{VideoURL: setting.Value}, nil
}

func (s *SettingsService) SetPracticePageVideo(ctx context.Context, request *domain.SetPracticePageVideoRequest) (*domain.SetPracticePageVideoResponse, error) {
	if err := request.Valid(ctx); err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}

	// An empty URL is a valid value: it clears the video so the practice page
	// shows its placeholder frame.
	if err := s.replacePracticePageVideo(ctx, strings.TrimSpace(request.VideoURL)); err != nil {
		return nil, err
	}
	return &domain.SetPracticePageVideoResponse{Success: true}, nil
}
