package paymenthandler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/payment"
	"github.com/GaryHY/event-reservation-app/internal/http/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/http/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"

	"github.com/stripe/stripe-go/v79"
)

// this is for the admin only

func CreateEventProduct(p *payment.Service, e *event.Service) http.Handler {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", stripe.Key))
		event, err := serverutil.Decode[event.Event](r)
		if err != nil {

		}
		// use the service to make the request
		priceID, err := p.CreateProduct(&event)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create product", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// use the priceID to update the corresponding field of event.
		event.PriceID = priceID
		// TODO: finish the implementation for the
		_, err = e.ModifyEvent(ctx, &event)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update the priceID for event", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// send a status created to the client
		if err = serverutil.Encode(w, http.StatusCreated, struct {
			EventID string `json:"eventid"`
		}{
			EventID: event.ID,
		}); err != nil {
			slog.ErrorContext(ctx, "failed to encode eventID for product registered", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
	return mw.EnableHeaders(handler, "Authorization")
}

// TODO: finish the implementation of that thing.
func DeleteEventProduct(p *payment.Service, e *event.Service) http.Handler {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the event id brother for the thing
		eventID := r.PathValue("id")
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		_, err := p.RemoveProduct(ctx, eventID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to remove product", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err = serverutil.Encode(w, http.StatusCreated, struct {
			EventID string `json:"eventid"`
		}{
			EventID: eventID,
		}); err != nil {
			slog.ErrorContext(ctx, "failed to encode eventID for product registered", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
	return mw.EnableHeaders(handler, "Authorization")
}
