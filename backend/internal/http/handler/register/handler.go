package registerhandler

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/domain/checkout"
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	"github.com/GaryHY/event-reservation-app/internal/http/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/http/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func MakeRegistration(reg *register.Service, e *event.Service, ch checkout.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		eventID := r.PathValue("id")
		event, err := e.Repo.GetEventByID(ctx, eventID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get event with given ID", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		userID := ctx.Value(mw.SessionIDKey).(string)
		spot := r.PathValue("spot")
		// get the client domain
		domain := os.Getenv("domain")
		// get the URL to redirect the client to in case of success or failure.
		url, err := ch.CreateCheckoutSession(ctx, domain, event.PriceID)
		// send the url to the client
		type Response struct {
			URL string `json:"url"`
		}
		if err := serverutil.Encode[Response](w, http.StatusSeeOther, Response{URL: url}); err != nil {
			slog.ErrorContext(ctx, "failed to encode URL to client", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// TODO: How do I get back the information that the client paid for the service ?
		if err := reg.CreateRegistration(ctx, userID, spot, event); err != nil {
			slog.ErrorContext(ctx, "failed creating registration for user", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err = e.DecreasePlacecount(ctx, eventID); err != nil {
			slog.ErrorContext(ctx, "failed decreasing placecount for event", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

// TODO: make the implementation of that for later. Do I need to give money back to people ?
func DeleteRegistration() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}
