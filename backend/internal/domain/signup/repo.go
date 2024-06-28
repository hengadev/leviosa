package user

import (
	"context"
)

type Reader interface{}
type Writer interface {
	AddAccount(ctx context.Context, account *Account) error
}

type ReadWriter interface {
	Reader
	Writer
}
