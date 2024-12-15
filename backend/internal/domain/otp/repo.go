package otpService

import (
	"context"
)

type Reader interface {
	GetOTP(ctx context.Context, email string) (*OTP, error)
}
type Writer interface {
	StoreOTP(ctx context.Context, otp *OTP) error
	InvalidateOTP(ctx context.Context, email string) error
}

type ReadWriter interface {
	Reader
	Writer
}
