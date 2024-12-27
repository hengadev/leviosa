package payment

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"

	"github.com/stripe/stripe-go/v79"
)

// TODO: finish the implementation of that thing.
// func DeleteEventProduct(p *payment.Service, e *event.Service) http.Handler {
func (a *AppInstance) DeleteEventProduct() http.Handler {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get event id
		ctx := r.Context()

		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		eventID := r.PathValue("id")
		_, err = a.Svcs.Stripe.RemoveProduct(ctx, eventID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to remove product", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		type Response struct {
			EventID string `json:"eventid"`
		}
		if err = serverutil.Encode(w, http.StatusCreated, Response{eventID}); err != nil {
			logger.ErrorContext(ctx, "failed to encode eventID for product registered", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
	return mw.EnableHeaders(handler, "Authorization")
}
