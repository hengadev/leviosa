package app

import (
	"github.com/hengadev/leviosa/internal/domain/event"
	"github.com/hengadev/leviosa/internal/domain/mail"
	"github.com/hengadev/leviosa/internal/domain/media"
	"github.com/hengadev/leviosa/internal/domain/message"
	"github.com/hengadev/leviosa/internal/domain/otp"
	"github.com/hengadev/leviosa/internal/domain/product"
	"github.com/hengadev/leviosa/internal/domain/register"
	"github.com/hengadev/leviosa/internal/domain/session"
	"github.com/hengadev/leviosa/internal/domain/stripe"
	"github.com/hengadev/leviosa/internal/domain/throttler"
	"github.com/hengadev/leviosa/internal/domain/user"
	"github.com/hengadev/leviosa/internal/domain/vote"
)

type Services struct {
	User      *userService.Service
	Session   *sessionService.Service
	Event     *eventService.Service
	Vote      *vote.Service
	Stripe    *stripeService.Service
	Register  *registerService.Service
	Media     *mediaService.Service
	Throttler *throttlerService.Service
	OTP       *otpService.Service
	Mail      *mailService.Service
	Product   *productService.Service
	Message   *messageService.Service
}
