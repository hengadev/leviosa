package app

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/checkout"
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/payment"
	"github.com/GaryHY/event-reservation-app/internal/domain/photo"
	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/throttler"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
)

type Services struct {
	User      *userService.Service
	Session   *sessionService.Service
	Event     *eventService.Service
	Payment   *payment.Service
	Vote      *vote.Service
	Checkout  *checkout.Service
	Register  *register.Service
	Photo     *photo.Service
	Throttler *throttlerService.Service
}
