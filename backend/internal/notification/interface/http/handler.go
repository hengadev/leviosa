package http

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware/auth"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/ports"
)

// Handler exposes the notification endpoints: the public contact form,
// meant for anonymous site visitors, and the administrator-only email
// previews.
type Handler interface {
	RegisterRoutes(router *http.ServeMux)
	SubmitContact(w http.ResponseWriter, r *http.Request)
	ListEmailPreviews(w http.ResponseWriter, r *http.Request)
	GetEmailPreview(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	svc    ports.NotificationService
	authmw auth.AuthMiddleware
}

func New(svc ports.NotificationService, authmw auth.AuthMiddleware) *handler {
	return &handler{svc: svc, authmw: authmw}
}
