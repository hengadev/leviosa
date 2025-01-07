package userService

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
)

type Reader interface {
	FindAccountByID(ctx context.Context, id string) (*models.User, error)
	HasUser(ctx context.Context, emailHash string) error
	HasOAuthUser(ctx context.Context, emailHash string, provider models.ProviderType) error
	GetHashedPasswordByEmail(ctx context.Context, email string) (string, error)
	GetUnverifiedUser(ctx context.Context, emailHash string) (*models.User, error)
	GetPendingUser(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error)
	GetPendingUsers(ctx context.Context) ([]*models.User, error)
	GetUserSessionData(ctx context.Context, email string) (string, models.Role, error)
}
type Writer interface {
	AddUser(ctx context.Context, user *models.User, provider models.ProviderType) error
	AddPendingUser(ctx context.Context, user *models.User, provider models.ProviderType) error
	AddUnverifiedUser(ctx context.Context, user *models.User) error
	ModifyAccount(ctx context.Context, user *models.User, whereMap map[string]any, prohibitedFields ...string) error
	DeleteUser(ctx context.Context, id string) error
}

type ReadWriter interface {
	Reader
	Writer
}
