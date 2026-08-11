package waitlistHandler

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
)

type Handler interface {
	RegisterRoutes(router *http.ServeMux)
	JoinWaitlist(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	svc ports.WaitlistService
}

func New(svc ports.WaitlistService) Handler {
	return &handler{svc: svc}
}
