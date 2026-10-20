package reviewHandler

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	mw "github.com/Leviosa-care/leviosa/backend/internal/common/middleware"
)

func (h *handler) RegisterRoutes(router *http.ServeMux) {
	RequireAdmin := h.authmw.RequireAdmin
	RequireStandard := h.authmw.RequireMinimumRole(identity.Standard)

	// Client submits a review for a completed booking they own.
	router.HandleFunc("POST /bookings/{id}/review", RequireStandard(mw.EnableCORS(h.CreateReview)))

	// Public read endpoints (unauthenticated) — non-hidden reviews plus an aggregate
	// rating for a partner/product, used on services/[id] and team/[id] pages.
	// The {id}/<subresource> shape avoids net/http ServeMux pattern conflicts:
	// /products/reviews/{productId} conflicts with /products/{id}/prices (two
	// incomparable wildcard/literal mixes), which panics route registration at
	// startup.
	router.HandleFunc("GET /partners/reviews/{partnerId}", mw.EnableCORS(h.GetPartnerReviews))
	router.HandleFunc("GET /products/{id}/reviews", mw.EnableCORS(h.GetProductReviews))

	// Staff/admin moderation: hide (not delete) a review, or reverse that.
	router.HandleFunc("POST /reviews/{id}/hide", RequireAdmin(mw.EnableCORS(h.HideReview)))
	router.HandleFunc("POST /reviews/{id}/unhide", RequireAdmin(mw.EnableCORS(h.UnhideReview)))
}
