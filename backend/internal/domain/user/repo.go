package user

import (
	"context"
)

type Reader interface {
	FindAccountByID(ctx context.Context, id string) (*User, error)
}
type Writer interface {
	AddAccount(ctx context.Context, user *User) error
	ModifyAccount(ctx context.Context, user *User) error
}

type ReadWriter interface {
	Reader
	Writer
}
