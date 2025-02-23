package app

import (
	"github.com/GaryHY/leviosa/internal/domain/event"
	"github.com/GaryHY/leviosa/internal/domain/mail"
	"github.com/GaryHY/leviosa/internal/domain/media"
	"github.com/GaryHY/leviosa/internal/domain/message"
	"github.com/GaryHY/leviosa/internal/domain/otp"
	"github.com/GaryHY/leviosa/internal/domain/product"
	"github.com/GaryHY/leviosa/internal/domain/register"
	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/domain/stripe"
	"github.com/GaryHY/leviosa/internal/domain/throttler"
	"github.com/GaryHY/leviosa/internal/domain/user"
	"github.com/GaryHY/leviosa/internal/domain/vote"
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
