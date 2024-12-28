package otpService

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain"
)

const (
	OTPDURATION    = 15 * time.Minute
	MaxOTPAttempts = 3
)

type OTP struct {
	EmailHash string    `json:"email_hash"`
	Code      string    `json:"code" validate:"len=6"`
	Attempts  int       `json:"attempts"`
	ExpiresAt time.Time `json:"expires_at"`
	Created   time.Time `json:"created"`
}

func NewOTP(emailHash string) (*OTP, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate secure random number: %w", err)
	}
	num := int(binary.BigEndian.Uint32(bytes) % 100000000)
	return &OTP{
		EmailHash: emailHash,
		Code:      fmt.Sprintf("%06d", num),
		Attempts:  1,
		Created:   time.Now(),
		ExpiresAt: time.Now().Add(OTPDURATION),
	}, nil
}

func (o *OTP) IncreaseAttempt() error {
	if o.Attempts+1 >= MaxOTPAttempts {
		return domain.NewInvalidValueErr("max attempts reached for provided OTP")
	}
	o.Attempts++
	return nil
}
