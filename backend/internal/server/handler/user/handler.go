package userHandler

import (
	"github.com/GaryHY/event-reservation-app/internal/server/service"
)

type Handler struct {
	*handler.Handler
}

func New(handler *handler.Handler) *Handler {
	return &Handler{handler}
}
