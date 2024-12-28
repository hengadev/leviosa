package payment

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"

	"github.com/stripe/stripe-go/v79"
)

// func CreateEventProduct(p *payment.Service, e *event.Service) http.Handler {
func (a *AppInstance) CreateEventProduct() http.Handler {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", stripe.Key))
		event, err := serverutil.Decode[eventService.Event](r.Body)
		if err != nil {
			logger.ErrorContext(ctx, "failed to decode event", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// use service to make request
		priceID, err := a.Svcs.Stripe.CreateProduct(ctx, &event)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create product", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// use priceID to update corresponding field of event
		event.PriceID = priceID
		// TODO: finish the implementation for the
		err = a.Svcs.Event.ModifyEvent(ctx, &event)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update the priceID for event", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// send status created to client
		if err = serverutil.Encode(w, http.StatusCreated, struct {
			EventID string `json:"eventid"`
		}{
			EventID: event.ID,
		}); err != nil {
			logger.ErrorContext(ctx, "failed to encode eventID for product registered", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
	return mw.EnableHeaders(handler, "Authorization")
}
