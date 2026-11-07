package aggregatorHandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

func (h *handler) BecomePractitioner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	sessionInfo, ok := session.SessionInfoFromContext(ctx)
	if !ok {
		httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
		return
	}

	var request domain.BecomePractitionerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("invalid request body: %v", err)), http.StatusBadRequest)
		return
	}

	partnerID, err := h.svc.BecomePractitioner(ctx, sessionInfo.UserID, &request)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "become practitioner")
		return
	}

	logger.InfoContext(ctx, "Handler: Become practitioner completed",
		"operation", "become_practitioner",
		"user_id", sessionInfo.UserID.String(),
		"partner_id", partnerID.String(),
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, map[string]string{"partner_id": partnerID.String()}, http.StatusOK)
}
