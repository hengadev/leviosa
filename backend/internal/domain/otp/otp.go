package otpService

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

const OTPDURATION = 15 * time.Minute

type OTP struct {
	Email     string
	Value     string `json:"value" validate:"len=6"`
	ExpiresAt time.Time
}

func NewOTP(email string) (*OTP, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate secure random number: %w", err)
	}
	num := int(binary.BigEndian.Uint32(bytes) % 100000000)
	return &OTP{
		Email:     email,
		Value:     fmt.Sprintf("%06d", num),
		ExpiresAt: time.Now().Add(OTPDURATION),
	}, nil
}

// the function service that I need to implement:
// - create otp
// - validate otp
// - invalidate otp (if the user is now signed in, I do not need someone to be able to use his otp for something else)
