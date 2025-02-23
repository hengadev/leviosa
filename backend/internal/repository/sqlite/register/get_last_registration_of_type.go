package registerRepository

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain/register"
)

func (r *repository) GetLastRegistrationOfType(ctx context.Context, count int, regType registerService.RegistrationType, userID string) ([]*registerService.Registration, error) {
	return nil, nil
}
