package userHandler

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

type billingPortalLinkRequest struct {
	ReturnURL string `json:"return_url"`
}

type billingPortalLinkResponse struct {
	URL string `json:"url"`
}

// GetBillingPortalLink generates a Stripe Billing Portal session URL for the authenticated
// user, so they can manage their saved payment method through Stripe's hosted UI.
func (h *handler) GetBillingPortalLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	// Get session info from context (set by middleware)
	sessionInfo, ok := session.SessionInfoFromContext(ctx)
	if !ok {
		logger.ErrorContext(ctx, "Handler: No session info in context",
			"operation", "get_billing_portal_link",
			"method", r.Method,
			"path", r.URL.Path)
		httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
		return
	}

	req, err := httpx.Decode[billingPortalLinkRequest](r.Body)
	if err != nil || req.ReturnURL == "" {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("return_url is required"), http.StatusBadRequest)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing get billing portal link request",
		"operation", "get_billing_portal_link",
		"method", r.Method,
		"path", r.URL.Path,
		"user_id", sessionInfo.UserID)

	url, err := h.svc.GetBillingPortalLink(ctx, sessionInfo.UserID, req.ReturnURL)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get billing portal link")
		return
	}

	logger.InfoContext(ctx, "Handler: Get billing portal link completed",
		"operation", "get_billing_portal_link",
		"method", r.Method,
		"path", r.URL.Path,
		"user_id", sessionInfo.UserID,
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, billingPortalLinkResponse{URL: url}, http.StatusOK)
}
