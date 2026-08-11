package http

import (
	"net/http"

	mw "github.com/Leviosa-care/leviosa/backend/internal/common/middleware"
)

func (h *handler) RegisterRoutes(router *http.ServeMux) {
	// Public contact form — unauthenticated, used by the (public)/contact page.
	router.HandleFunc("POST "+ContactEndpoint, mw.EnableCORS(h.SubmitContact))
}
