package app

import (
	"github.com/GaryHY/leviosa/internal/domain/event"
	"github.com/GaryHY/leviosa/internal/domain/media"
	"github.com/GaryHY/leviosa/internal/domain/otp"
	"github.com/GaryHY/leviosa/internal/domain/product"
	"github.com/GaryHY/leviosa/internal/domain/register"
	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/domain/throttler"
	"github.com/GaryHY/leviosa/internal/domain/user"
	"github.com/GaryHY/leviosa/internal/domain/vote"
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
