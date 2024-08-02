package user

import (
	"context"
)

type Reader interface {
	FindAccountByID(ctx context.Context, id string) (*User, error)
	ValidateCredentials(ctx context.Context, usr *Credentials) (string, Role, error)
}
type Writer interface {
	AddAccount(ctx context.Context, user *User) (string, error)
	ModifyAccount(ctx context.Context, user *User) error
}

type ReadWriter interface {
	Reader
	Writer
}
