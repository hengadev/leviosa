package checkouthandler

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/domain/checkout"
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/http/handler"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
	"github.com/stripe/stripe-go/v79"
)

func CreateCheckoutSession(ch checkout.Service, e event.Reader) http.Handler {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get the event ID from the path URL
		eventID := r.PathValue("id")
		// get the priceID associated with the event
		priceID, err := e.GetPriceIDByEventID(ctx, eventID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create account", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// the right value for the priceID that I need to use.
		_ = priceID
		// just to test things out
		price_temp := "price_1OsTQfHwHXlEm0ohh1sSBXJa"
		domain := os.Getenv("BASE_URL")

		sessionURL, err := ch.CreateCheckoutSession(ctx, domain, price_temp)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create session checkout session", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// NOTE: Can not redirect with that for some reason. So I want to send the url in json format. Yet this is in the tutorial
		// fmt.Println("Print the s.URL thing: ", s.URL)
		// http.Redirect(w, r, s.URL, http.StatusSeeOther)

		message := struct {
			Url string `json:"url"`
		}{
			Url: sessionURL,
		}
		if err = serverutil.Encode(w, http.StatusInternalServerError, message); err != nil {
			slog.ErrorContext(ctx, "failed to encode checkout session URL", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
