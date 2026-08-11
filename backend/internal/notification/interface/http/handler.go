package http

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/notification/ports"
)

// Handler exposes the public-facing notification endpoints (currently just
// the contact form). Unlike other domain handlers, this one is intentionally
// unauthenticated end to end — it's meant for anonymous site visitors.
type Handler interface {
	RegisterRoutes(router *http.ServeMux)
	SubmitContact(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	svc ports.NotificationService
}

func New(svc ports.NotificationService) *handler {
	return &handler{svc: svc}
}
