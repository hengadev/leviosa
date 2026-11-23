package bookingHandler

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	mw "github.com/Leviosa-care/leviosa/backend/internal/common/middleware"
)

func (h *handler) RegisterRoutes(router *http.ServeMux) {
	RequireAdmin := h.authmw.RequireAdmin
	RequirePartner := h.authmw.RequireMinimumRole(identity.Partner)
	RequireStandard := h.authmw.RequireMinimumRole(identity.Standard)

	// Admin endpoints
	router.HandleFunc("GET /admin/dashboard/stats", RequireAdmin(mw.EnableCORS(h.GetDashboardStats)))
	router.HandleFunc("GET /admin/bookings", RequireAdmin(mw.EnableCORS(h.GetAdminBookings)))
	router.HandleFunc("GET /admin/analytics/summary", RequireAdmin(mw.EnableCORS(h.GetAnalyticsSummary)))
	router.HandleFunc("GET /admin/bookings/financial-summary", RequireAdmin(mw.EnableCORS(h.GetFinancialSummary)))
	// Email the practice access to the clients of upcoming bookings (issue 026)
	router.HandleFunc("GET /admin/bookings/access-update", RequireAdmin(mw.EnableCORS(h.GetAccessUpdateRecipients)))
	router.HandleFunc("POST /admin/bookings/access-update", RequireAdmin(mw.EnableCORS(h.SendAccessUpdate)))
	router.HandleFunc("POST /admin/bookings/{id}/override-cancel", RequireAdmin(mw.EnableCORS(h.AdminOverrideCancelBooking)))

	// Public booking endpoints (unauthenticated — used by guest booking page)
	router.HandleFunc("GET /bookings/lookup", mw.EnableCORS(h.LookupBooking))
	router.HandleFunc("POST /bookings/{id}/cancel-public", mw.EnableCORS(h.CancelBookingPublic))
	// The client left the payment step: free the slot now (ADR-0018)
	router.HandleFunc("POST /bookings/release-hold", mw.EnableCORS(h.ReleaseHold))
	// Free slots for the practice page's calendar (slot engine, issue 010)
	router.HandleFunc("GET /slots", mw.EnableCORS(h.GetFreeSlots))

	// Internal booking claim endpoint (called by authuser after account creation).
	// Protected by service key auth: only authuser may call this.
	router.HandleFunc("POST /bookings/claim", h.authmw.RequireServiceAuth(mw.EnableCORS(h.ClaimBookings)))

	// Booking management endpoints
	// POST /bookings is public to allow guest bookings without authentication
	router.HandleFunc("POST /bookings", mw.EnableCORS(h.CreateBooking))
	router.HandleFunc("GET /bookings/{id}", RequireStandard(mw.EnableCORS(h.GetBooking)))
	router.HandleFunc("GET /clients/{clientId}/bookings", RequireStandard(mw.EnableCORS(h.GetClientBookings)))
	router.HandleFunc("GET /partners/bookings/{partnerId}", RequirePartner(mw.EnableCORS(h.GetPartnerBookings)))
	// Next working day for Aujourd'hui (issue 037): the endpoint checks the
	// same authorization as the partner bookings list above.
	router.HandleFunc("GET /partners/next-working-day/{partnerId}", RequirePartner(mw.EnableCORS(h.GetPartnerNextWorkingDay)))
	// Practitioner cancellation from the partner agenda (issue 018): full
	// refund; the booking's partner or an admin only (checked by the service).
	router.HandleFunc("POST /partners/bookings/{id}/cancel", RequirePartner(mw.EnableCORS(h.CancelBookingByPractitioner)))
	router.HandleFunc("GET /partners/earnings/{partnerId}", RequirePartner(mw.EnableCORS(h.GetPartnerEarnings)))
	// The client file (issue 042): the file of a booking's client, and the
	// distinct-client search. Same authorization as the partner bookings list.
	router.HandleFunc("GET /partners/bookings/{id}/client", RequirePartner(mw.EnableCORS(h.GetBookingClientFile)))
	router.HandleFunc("GET /partners/clients/{partnerId}", RequirePartner(mw.EnableCORS(h.SearchPartnerClients)))
	router.HandleFunc("GET /bookings", RequirePartner(mw.EnableCORS(h.GetUpcomingBookings)))
	router.HandleFunc("POST /bookings/{id}/cancel", RequireStandard(mw.EnableCORS(h.CancelBooking)))
	router.HandleFunc("POST /bookings/{id}/complete", RequirePartner(mw.EnableCORS(h.CompleteBooking)))
	router.HandleFunc("POST /bookings/{id}/no-show", RequirePartner(mw.EnableCORS(h.MarkNoShow)))
	router.HandleFunc("POST /bookings/{id}/reopen", RequirePartner(mw.EnableCORS(h.ReopenBooking)))
	// The on-site payment (issue 039): record it, or undo it the same day.
	// The booking's partner or an administrator only (checked by the handler).
	router.HandleFunc("POST /bookings/{id}/on-site-payment", RequirePartner(mw.EnableCORS(h.RecordOnSitePayment)))
	router.HandleFunc("DELETE /bookings/{id}/on-site-payment", RequirePartner(mw.EnableCORS(h.UndoOnSitePayment)))
	router.HandleFunc("POST /bookings/{id}/payment", RequireStandard(mw.EnableCORS(h.ProcessPayment)))
	router.HandleFunc("POST /bookings/{id}/refund", RequireAdmin(mw.EnableCORS(h.RefundBooking)))

	// Webhook endpoints (no auth required - verified by signature)
	router.HandleFunc("POST /webhooks/stripe", h.HandleStripeWebhook)
}
