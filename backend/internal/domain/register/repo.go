package registerService

import "context"

type Reader interface {
	HasRegistration(ctx context.Context, day, year int, month, userID string) error
	GetLastRegistrationOfType(ctx context.Context, count int, regType RegistrationType, userID string) ([]*Registration, error)
}

type Writer interface {
	AddRegistration(ctx context.Context, r *Registration, day, year int, month string) error
	RemoveRegistration(ctx context.Context, day, year int, month string) error
}

type ReadWriter interface {
	Reader
	Writer
}
