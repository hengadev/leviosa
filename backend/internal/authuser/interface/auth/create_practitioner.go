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

func (h *handler) CreatePractitioner(w http.ResponseWriter, r *http.Request) {
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

	var request domain.CreatePractitionerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("invalid request body: %v", err)), http.StatusBadRequest)
		return
	}

	response, err := h.svc.CreatePractitioner(ctx, sessionInfo.UserID, &request)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "create practitioner")
		return
	}

	// The generated password is in the response only, never in a log.
	logger.InfoContext(ctx, "Handler: Create practitioner completed",
		"operation", "create_practitioner",
		"user_id", response.UserID.String(),
		"partner_id", response.PartnerID.String(),
		"status_code", http.StatusCreated)

	httpx.RespondWithJSON(w, response, http.StatusCreated)
}
