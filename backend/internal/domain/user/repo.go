package user

import (
	"context"
)

type Reader interface {
	FindAccountByID(ctx context.Context, id string) (*User, error)
}
type Writer interface {
	AddAccount(ctx context.Context, account *User) error
}

type ReadWriter interface {
	Reader
	Writer
}
