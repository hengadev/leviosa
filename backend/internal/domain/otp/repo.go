package otpService

import (
	"context"
)

type Reader interface {
	GetOTP(ctx context.Context, email string) ([]byte, error)
}
type Writer interface {
	StoreOTP(ctx context.Context, emailHash string, otpEncoded []byte) error
	InvalidateOTP(ctx context.Context, email string) error
	ValidateOTP(ctx context.Context, emailHash, providedOTP string) error
}

type ReadWriter interface {
	Reader
	Writer
}
