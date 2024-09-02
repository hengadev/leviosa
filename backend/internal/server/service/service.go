package handler

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/checkout"
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/payment"
	"github.com/GaryHY/event-reservation-app/internal/domain/photo"
	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
)

type Services struct {
	User     *userService.Service
	Session  *sessionService.Service
	Event    *eventService.Service
	Payment  *payment.Service
	Vote     *vote.Service
	Checkout *checkout.Service
	Register *register.Service
	Photo    *photo.Service
}

type Repos struct {
	User     userService.Reader
	Session  sessionService.Reader
	Event    eventService.Reader
	Vote     vote.Reader
	Register register.Reader
	Photo    photo.Reader
}

type Handler struct {
	Svcs  *Services
	Repos *Repos
}

// Function to use in the main, once all the services and repos are built.
func New(svcs *Services, repos *Repos) *Handler {
	return &Handler{
		Svcs:  svcs,
		Repos: repos,
	}
}
