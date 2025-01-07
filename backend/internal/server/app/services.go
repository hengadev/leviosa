package app

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/mail"
	"github.com/GaryHY/event-reservation-app/internal/domain/media"
	"github.com/GaryHY/event-reservation-app/internal/domain/otp"
	"github.com/GaryHY/event-reservation-app/internal/domain/product"
	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/stripe"
	"github.com/GaryHY/event-reservation-app/internal/domain/throttler"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
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
}
