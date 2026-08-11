package reviewHandler

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/google/uuid"
)

// HideReview handles POST /reviews/{id}/hide. Restricted to staff/admin via
// RequireAdmin at the route layer. The review is retained, not deleted.
func (h *handler) HideReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	reviewID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid review ID format"), http.StatusBadRequest)
		return
	}

	review, err := h.svc.HideReview(ctx, reviewID)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "hide review")
		return
	}

	logger.InfoContext(ctx, "Handler: Review hidden", "review_id", review.ID, "operation", "hide_review")

	httpx.RespondWithJSON(w, domain.ReviewResponseFromReview(review), http.StatusOK)
}

// UnhideReview handles POST /reviews/{id}/unhide. Restricted to staff/admin via
// RequireAdmin at the route layer.
func (h *handler) UnhideReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	reviewID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid review ID format"), http.StatusBadRequest)
		return
	}

	review, err := h.svc.UnhideReview(ctx, reviewID)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "unhide review")
		return
	}

	logger.InfoContext(ctx, "Handler: Review unhidden", "review_id", review.ID, "operation", "unhide_review")

	httpx.RespondWithJSON(w, domain.ReviewResponseFromReview(review), http.StatusOK)
}
