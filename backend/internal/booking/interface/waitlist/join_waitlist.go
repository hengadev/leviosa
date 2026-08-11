package waitlistHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/Leviosa-care/leviosa/backend/internal/common/validation"
)

func (h *handler) JoinWaitlist(w http.ResponseWriter, r *http.Request) {
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

	var request domain.JoinWaitlistRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("invalid request body: %v", err)), http.StatusBadRequest)
		return
	}

	// Validate guest contact format when provided
	if request.GuestEmail != "" {
		if err := validation.ValidateEmail(request.GuestEmail); err != nil {
			httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("guest_email: %v", err)), http.StatusBadRequest)
			return
		}
	}
	if request.GuestPhone != "" {
		if err := validation.ValidatePhone(request.GuestPhone); err != nil {
			httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("guest_phone: %v", err)), http.StatusBadRequest)
			return
		}
	}

	response, err := h.svc.JoinWaitlist(ctx, &request)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "join waitlist")
		return
	}

	logger.InfoContext(ctx, "Handler: Waitlist entry created successfully",
		"waitlist_entry_id", response.ID,
		"product_id", response.ProductID,
		"is_guest", response.ClientID == nil,
		"operation", "join_waitlist")

	httpx.RespondWithJSON(w, response, http.StatusCreated)
}
