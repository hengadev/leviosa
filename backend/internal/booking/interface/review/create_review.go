package reviewHandler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/google/uuid"
)

// CreateReview handles POST /bookings/{id}/review. Only the booking's own client may
// submit a review, and only once the booking is `completed` — see ports.ReviewService
// for the full authorization contract.
func (h *handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		httpx.RespondWithError(w, errors.New("unsupported media type: please send 'application/json'"), http.StatusUnsupportedMediaType)
		return
	}

	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	sessionInfo, ok := session.SessionInfoFromContext(ctx)
	if !ok {
		httpx.RespondWithError(w, errs.NewUnauthorizedErr("session information required"), http.StatusUnauthorized)
		return
	}

	bookingID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid booking ID format"), http.StatusBadRequest)
		return
	}

	var request domain.CreateReviewRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid request body: "+err.Error()), http.StatusBadRequest)
		return
	}

	review, err := h.svc.CreateReview(ctx, bookingID, sessionInfo.UserID, request.Rating, request.Comment)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "create review")
		return
	}

	logger.InfoContext(ctx, "Handler: Review created", "review_id", review.ID, "booking_id", bookingID, "operation", "create_review")

	httpx.RespondWithJSON(w, domain.ReviewResponseFromReview(review), http.StatusCreated)
}
