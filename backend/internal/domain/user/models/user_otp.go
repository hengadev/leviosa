package models

import (
	"context"

	"github.com/hengadev/leviosa/pkg/errsx"
)

type UserOTP struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (u UserOTP) Valid(ctx context.Context) (problems errsx.Map) {
	var errs errsx.Map
	return errs
}
