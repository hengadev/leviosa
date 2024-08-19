package user

import (
	"context"
)

type Reader interface {
	FindAccountByID(ctx context.Context, id int) (*User, error)
	ValidateCredentials(ctx context.Context, usr *Credentials) (int, Role, error)
}
type Writer interface {
	AddAccount(ctx context.Context, user *User) (int, error)
	ModifyAccount(ctx context.Context, user *User, whereMap map[string]any, prohibitedFields ...string) (int, error)
}

type ReadWriter interface {
	Reader
	Writer
}
