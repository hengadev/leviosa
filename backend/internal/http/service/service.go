package handler

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/checkout"
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/payment"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

type Services struct {
	User     *user.Service
	Session  *session.Service
	Event    *event.Service
	Payment  *payment.Service
	Checkout *checkout.Service
}

type Repos struct {
	User    user.Reader
	Session session.Reader
	Event   event.Reader
}

type Handler struct {
	Svcs  *Services
	Repos *Repos
}

func NewHandler(svcs *Services, repos *Repos) *Handler {
	return &Handler{
		Svcs:  svcs,
		Repos: repos,
	}
}
