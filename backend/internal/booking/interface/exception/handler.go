// Package exceptionHandler exposes availability exceptions to the partner
// area (docs/issues/012-availability-exceptions.md). Every route is behind
// the partner role, and the service checks that the availability belongs to
// the caller (another partner's answers 404).
package exceptionHandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	mw "github.com/Leviosa-care/leviosa/backend/internal/common/middleware"
	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware/auth"

	"github.com/google/uuid"
)

type Handler interface {
	RegisterRoutes(router *http.ServeMux)
}

type handler struct {
	svc    ports.AvailabilityExceptionService
	authmw auth.AuthMiddleware
}

func New(svc ports.AvailabilityExceptionService, authmw auth.AuthMiddleware) Handler {
	return &handler{svc: svc, authmw: authmw}
}

func (h *handler) RegisterRoutes(router *http.ServeMux) {
	RequirePartner := h.authmw.RequireMinimumRole(identity.Partner)

	router.HandleFunc("GET /availabilities/{id}/exceptions", RequirePartner(mw.EnableCORS(h.ListExceptions)))
	router.HandleFunc("POST /availabilities/{id}/exceptions", RequirePartner(mw.EnableCORS(h.CreateException)))
	router.HandleFunc("DELETE /availabilities/{id}/exceptions/{exceptionId}", RequirePartner(mw.EnableCORS(h.DeleteException)))
}

func (h *handler) ListExceptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}
	partnerUserID, availabilityID, ok := caller(w, r)
	if !ok {
		return
	}

	exceptions, err := h.svc.ListExceptions(ctx, partnerUserID, availabilityID)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "list availability exceptions")
		return
	}
	response := make([]domain.AvailabilityExceptionResponse, 0, len(exceptions))
	for _, exception := range exceptions {
		response = append(response, domain.NewAvailabilityExceptionResponse(exception))
	}
	httpx.RespondWithJSON(w, response, http.StatusOK)
}

func (h *handler) CreateException(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}
	partnerUserID, availabilityID, ok := caller(w, r)
	if !ok {
		return
	}

	var request domain.CreateAvailabilityExceptionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("invalid request body: %v", err)), http.StatusBadRequest)
		return
	}
	startDate, endDate, err := request.Dates()
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr(err.Error()), http.StatusBadRequest)
		return
	}

	exception, affected, err := h.svc.CreateException(ctx, partnerUserID, availabilityID, startDate, endDate)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "create availability exception")
		return
	}
	httpx.RespondWithJSON(w, domain.CreateAvailabilityExceptionResponse{
		Exception:        domain.NewAvailabilityExceptionResponse(*exception),
		AffectedBookings: affected,
	}, http.StatusCreated)
}

func (h *handler) DeleteException(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}
	partnerUserID, availabilityID, ok := caller(w, r)
	if !ok {
		return
	}
	exceptionID, err := uuid.Parse(r.PathValue("exceptionId"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid exception ID format"), http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteException(ctx, partnerUserID, availabilityID, exceptionID); err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "delete availability exception")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// caller returns the calling partner's user ID and the availability ID of
// the path, or answers the request and returns false.
func caller(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	sessionInfo, ok := session.SessionInfoFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("session information from context required"), http.StatusUnauthorized)
		return uuid.Nil, uuid.Nil, false
	}
	availabilityID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid availability ID format"), http.StatusBadRequest)
		return uuid.Nil, uuid.Nil, false
	}
	return sessionInfo.UserID, availabilityID, true
}
