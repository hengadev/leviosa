package register

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/webhook"
)

func (app *AppInstance) MakeRegistration() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			logger.ErrorContext(ctx, "failed to get read request body", "error", err)
			serverutil.WriteResponse(w, handler.NewServiceUnavailableErr(err), http.StatusServiceUnavailable)
			return
		}
		// TODO: that thing should be hidden in an env variable : stripe_webhook_secret
		endpointSecret := "whsec_3c9b438ee0a665d78da90fc39667834f64e68766bef84013382a01d70e9711e9 "
		stripeEvent, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), endpointSecret)
		if err := json.Unmarshal(payload, &stripeEvent); err != nil {
			logger.ErrorContext(ctx, "failed to parse webhook body json", "error", err)
			serverutil.WriteResponse(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// I get that part from the part types of event of the documentation: https://docs.stripe.com/api/events/types
		var metadata map[string]string
		switch stripeEvent.Type {
		case "checkout.session.completed":
			var sessionCompletion stripe.CheckoutSession
			err = json.Unmarshal(stripeEvent.Data.Raw, &sessionCompletion)
			if err != nil {
				logger.ErrorContext(ctx, "failed to parse webhook body json", "error", err)
				serverutil.WriteResponse(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
				return
			}
			metadata = sessionCompletion.Metadata
		}
		_ = metadata
		// fmt.Println("the metadata", metadata)

		event, err := app.Repos.Event.GetEventByID(ctx, metadata["eventID"])
		if err != nil {
			logger.ErrorContext(ctx, "failed to get event with given ID", "error", err)
			serverutil.WriteResponse(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err := app.Svcs.Register.CreateRegistration(ctx, metadata["userID"], metadata["spot"], event); err != nil {
			logger.ErrorContext(ctx, "failed creating registration for user", "error", err)
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err = app.Svcs.Event.DecreasePlacecount(ctx, metadata["eventID"]); err != nil {
			logger.ErrorContext(ctx, "failed decreasing placecount for event", "error", err)
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
