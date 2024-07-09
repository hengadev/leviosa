package registerhandler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/checkout"
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	"github.com/GaryHY/event-reservation-app/internal/http/handler"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/webhook"
)

func MakeRegistration(reg *register.Service, e *event.Service, ch *checkout.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get read request body", "error", err)
			http.Error(w, handler.NewServiceUnavailableErr(err), http.StatusServiceUnavailable)
			return
		}
		// TODO: that thing should be hidden in an env variable : stripe_webhook_secret
		endpointSecret := "whsec_3c9b438ee0a665d78da90fc39667834f64e68766bef84013382a01d70e9711e9 "
		stripeEvent, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), endpointSecret)
		if err := json.Unmarshal(payload, &stripeEvent); err != nil {
			slog.ErrorContext(ctx, "failed to parse webhook body json", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// I get that part from the part types of event of the documentation: https://docs.stripe.com/api/events/types
		var metadata map[string]string
		switch stripeEvent.Type {
		case "checkout.session.completed":
			var sessionCompletion stripe.CheckoutSession
			err = json.Unmarshal(stripeEvent.Data.Raw, &sessionCompletion)
			if err != nil {
				slog.ErrorContext(ctx, "failed to parse webhook body json", "error", err)
				http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
				return
			}
			metadata = sessionCompletion.Metadata
		}
		_ = metadata
		// fmt.Println("the metadata", metadata)

		event, err := e.Repo.GetEventByID(ctx, metadata["eventID"])
		if err != nil {
			slog.ErrorContext(ctx, "failed to get event with given ID", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err := reg.CreateRegistration(ctx, metadata["userID"], metadata["spot"], event); err != nil {
			slog.ErrorContext(ctx, "failed creating registration for user", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err = e.DecreasePlacecount(ctx, metadata["eventID"]); err != nil {
			slog.ErrorContext(ctx, "failed decreasing placecount for event", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
