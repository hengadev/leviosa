package app

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/media"
	"github.com/GaryHY/event-reservation-app/internal/domain/otp"
	"github.com/GaryHY/event-reservation-app/internal/domain/product"
	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/throttler"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
)

type Repos struct {
	User      userService.Reader
	Session   sessionService.Reader
	Event     eventService.Reader
	Vote      vote.Reader
	Register  registerService.Reader
	Media     mediaService.Reader
	Throttler throttlerService.Reader
	Product   productService.Reader
	OTP       otpService.Reader
}
