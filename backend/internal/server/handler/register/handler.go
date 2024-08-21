package register

import (
	"github.com/GaryHY/event-reservation-app/internal/server/service"
)

type Handler struct {
	*handler.Handler
}

func NewHandler(handler *handler.Handler) *Handler {
	return &Handler{handler}
}
