package register

import "context"

type Reader interface {
	HasRegistration(ctx context.Context, day, year int, month, userID string) error
}

type Writer interface {
	AddRegistration(ctx context.Context, r *Registration, day, year int, month string) error
	RemoveRegistration(ctx context.Context, day, year int, month string) error
}

type ReadWriter interface {
	Reader
	Writer
}
