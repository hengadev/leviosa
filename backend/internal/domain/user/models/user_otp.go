package models

import (
	"context"

	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

type UserOTP struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (u UserOTP) Valid(ctx context.Context) (problems errsx.Map) {
	var errs errsx.Map
	return errs
}
