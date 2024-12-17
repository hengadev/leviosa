package app

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/photo"
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
	Register  register.Reader
	Photo     photo.Reader
	Throttler throttlerService.Reader
}
