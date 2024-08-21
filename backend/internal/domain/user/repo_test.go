package user_test

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

type StubUserRepository struct {
	users []*user.User
}

func NewStubUserRepository() *StubUserRepository {
	return &StubUserRepository{}
}

func (s *StubUserRepository) FindAccountByID(ctx context.Context, id int) (*user.User, error) {
	return nil, nil
}

func (s *StubUserRepository) GetCredentials(ctx context.Context, usr *user.Credentials) (int, string, user.Role, error) {
	return 0, "", user.UNKNOWN, nil
}

func (s *StubUserRepository) AddAccount(ctx context.Context, user *user.User) error {
	for _, usr := range s.users {
		if usr == user {
			return fmt.Errorf("user already exists")
		}
	}
	s.users = append(s.users, user)
	return nil
}

func (s *StubUserRepository) ModifyAccount(ctx context.Context, user *user.User, whereMap map[string]any, prohibitedFields ...string) error {
	return nil
}
