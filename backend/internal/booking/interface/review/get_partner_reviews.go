package reviewHandler

import (
	"net/http"
	"strconv"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/google/uuid"
)

// GetPartnerReviews handles GET /partners/{partnerId}/reviews. Public, unauthenticated:
// returns non-hidden reviews plus an aggregate rating for the partner.
func (h *handler) GetPartnerReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	partnerID, err := uuid.Parse(r.PathValue("partnerId"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid partner ID format"), http.StatusBadRequest)
		return
	}

	limit, offset := parsePagination(r)

	response, err := h.svc.GetPublicReviewsForPartner(ctx, partnerID, limit, offset)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get partner reviews")
		return
	}

	httpx.RespondWithJSON(w, response, http.StatusOK)
}

// parsePagination reads ?limit= and ?offset= query params, ignoring invalid values
// (the service layer applies its own default when limit <= 0).
func parsePagination(r *http.Request) (limit, offset int) {
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = o
	}
	return limit, offset
}
