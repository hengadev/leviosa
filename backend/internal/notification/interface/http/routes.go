package http

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	mw "github.com/Leviosa-care/leviosa/backend/internal/common/middleware"
)

func (h *handler) RegisterRoutes(router *http.ServeMux) {
	RequireAdministrator := h.authmw.RequireMinimumRole(identity.Administrator)

	// Public contact form — unauthenticated, used by the (public)/contact page.
	router.HandleFunc("POST "+ContactEndpoint, mw.EnableCORS(h.SubmitContact))

	// Admin-only email previews, rendered with sample data.
	router.HandleFunc("GET "+AdminEmailPreviewsEndpoint, RequireAdministrator(mw.EnableCORS(h.ListEmailPreviews)))
	router.HandleFunc("GET "+AdminEmailPreviewEndpoint, RequireAdministrator(mw.EnableCORS(h.GetEmailPreview)))
}
