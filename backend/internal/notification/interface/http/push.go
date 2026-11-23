package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	mw "github.com/Leviosa-care/leviosa/backend/internal/common/middleware"
	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware/auth"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/ports"
)

// maxPushRequestBytes bounds a subscription request body.
const maxPushRequestBytes = 8 << 10

// PushHandler exposes the Cabinet switch of practitioner notifications
// (ADR-0020): whether Web Push is available, and subscribing or
// unsubscribing the current device. Practitioners and administrators only.
type PushHandler struct {
	svc    ports.PushSubscriptionService
	authmw auth.AuthMiddleware
}

func NewPushHandler(svc ports.PushSubscriptionService, authmw auth.AuthMiddleware) *PushHandler {
	return &PushHandler{svc: svc, authmw: authmw}
}

func (h *PushHandler) RegisterRoutes(router *http.ServeMux) {
	requirePartner := h.authmw.RequireMinimumRole(identity.Partner)

	router.HandleFunc("GET "+PushStatusEndpoint, requirePartner(mw.EnableCORS(h.GetStatus)))
	router.HandleFunc("PUT "+PushSubscriptionEndpoint, requirePartner(mw.EnableCORS(h.Subscribe)))
	router.HandleFunc("DELETE "+PushSubscriptionEndpoint, requirePartner(mw.EnableCORS(h.Unsubscribe)))
}

type pushStatusResponse struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key,omitempty"`
}

// GetStatus answers whether notifications are available, with the VAPID
// public key the browser subscribes with.
func (h *PushHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.svc.Status()
	httpx.RespondWithJSON(w, pushStatusResponse{Enabled: status.Enabled, PublicKey: status.PublicKey}, http.StatusOK)
}

// pushSubscriptionRequest is the browser's PushSubscription.toJSON(), plus
// the device's user agent.
type pushSubscriptionRequest struct {
	Endpoint       string  `json:"endpoint"`
	ExpirationTime *int64  `json:"expirationTime,omitempty"`
	Keys           pushKey `json:"keys"`
	UserAgent      string  `json:"user_agent,omitempty"`
}

type pushKey struct {
	P256dh string `json:"p256dh"`
	Auth   string `json:"auth"`
}

// Subscribe stores the current device for the signed-in user (204), 503
// while Web Push is not configured.
func (h *PushHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := requestLogger(r)

	user, ok := session.SessionInfoFromContext(ctx)
	if !ok {
		httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
		return
	}

	var req pushSubscriptionRequest
	if err := decodePushBody(w, r, &req); err != nil {
		httpx.RespondWithError(w, err, http.StatusBadRequest)
		return
	}
	userAgent := req.UserAgent
	if userAgent == "" {
		userAgent = r.UserAgent()
	}

	err := h.svc.Subscribe(ctx, user.UserID, ports.PushSubscriptionInput{
		Endpoint:  req.Endpoint,
		P256dh:    req.Keys.P256dh,
		Auth:      req.Keys.Auth,
		UserAgent: userAgent,
	})
	if errors.Is(err, domain.ErrPushUnavailable) {
		httpx.RespondWithError(w, err, http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "subscribe to push notifications")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type pushUnsubscribeRequest struct {
	Endpoint string `json:"endpoint"`
}

// Unsubscribe deletes the current device of the signed-in user (204, also
// when it was not stored).
func (h *PushHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := requestLogger(r)

	user, ok := session.SessionInfoFromContext(ctx)
	if !ok {
		httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
		return
	}

	var req pushUnsubscribeRequest
	if err := decodePushBody(w, r, &req); err != nil {
		httpx.RespondWithError(w, err, http.StatusBadRequest)
		return
	}
	if err := h.svc.Unsubscribe(ctx, user.UserID, req.Endpoint); err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "unsubscribe from push notifications")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodePushBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxPushRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return errs.NewInvalidValueErr(fmt.Sprintf("invalid request body: %v", err))
	}
	return nil
}

func requestLogger(r *http.Request) *slog.Logger {
	if logger, err := ctxutil.GetLoggerFromContext(r.Context()); err == nil {
		return logger
	}
	return slog.Default()
}
