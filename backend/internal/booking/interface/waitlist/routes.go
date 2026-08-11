package waitlistHandler

import (
	"net/http"

	mw "github.com/Leviosa-care/leviosa/backend/internal/common/middleware"
)

func (h *handler) RegisterRoutes(router *http.ServeMux) {
	// POST /waitlist is public to allow guest waitlist joins without
	// authentication, mirroring POST /bookings (ADR-0008).
	router.HandleFunc("POST /waitlist", mw.EnableCORS(h.JoinWaitlist))
}
