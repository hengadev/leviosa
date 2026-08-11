package user

import (
	"github.com/Leviosa-care/leviosa/backend/internal/authuser/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"
	"github.com/hengadev/encx"
)

type UserService struct {
	repo   ports.UserRepository
	crypto encx.CryptoService
	stripe ports.StripeService
	audit  audit.Logger
}

// New creates a new instance of the aggregator service. audit may be nil
// (e.g. in tests that don't need audit assertions); use cases that write
// audit entries no-op in that case rather than failing.
func New(user ports.UserRepository, crypto encx.CryptoService, stripe ports.StripeService, auditLogger audit.Logger) ports.UserService {
	return &UserService{
		repo:   user,
		crypto: crypto,
		stripe: stripe,
		audit:  auditLogger,
	}
}
