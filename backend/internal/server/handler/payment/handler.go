package payment

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/http/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/http/middleware"
	"github.com/GaryHY/event-reservation-app/internal/http/service"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"

	"github.com/stripe/stripe-go/v79"
)

type Handler struct {
	*handler.Handler
}

func NewHandler(handler *handler.Handler) *Handler {
	return &Handler{handler}
}

// this is for the admin only

// func CreateEventProduct(p *payment.Service, e *event.Service) http.Handler {
func (h *Handler) CreateEventProduct() http.Handler {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", stripe.Key))
		event, err := serverutil.Decode[event.Event](r)
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode event", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// use the service to make the request
		priceID, err := h.Svcs.Payment.CreateProduct(&event)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create product", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// use the priceID to update the corresponding field of event.
		event.PriceID = priceID
		// TODO: finish the implementation for the
		_, err = h.Svcs.Event.ModifyEvent(ctx, &event)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update the priceID for event", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// send a status created to the client
		if err = serverutil.Encode(w, http.StatusCreated, struct {
			EventID string `json:"eventid"`
		}{
			EventID: event.ID,
		}); err != nil {
			slog.ErrorContext(ctx, "failed to encode eventID for product registered", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
	return mw.EnableHeaders(handler, "Authorization")
}

// TODO: finish the implementation of that thing.
// func DeleteEventProduct(p *payment.Service, e *event.Service) http.Handler {
func (h *Handler) DeleteEventProduct() http.Handler {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the event id brother for the thing
		eventID := r.PathValue("id")
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		_, err := h.Svcs.Payment.RemoveProduct(ctx, eventID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to remove product", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		type Response struct {
			EventID string `json:"eventid"`
		}
		if err = serverutil.Encode(w, http.StatusCreated, Response{eventID}); err != nil {
			slog.ErrorContext(ctx, "failed to encode eventID for product registered", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
	return mw.EnableHeaders(handler, "Authorization")
}
