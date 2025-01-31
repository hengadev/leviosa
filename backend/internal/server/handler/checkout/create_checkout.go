package checkout

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
	"github.com/stripe/stripe-go/v79"
)

func (a *AppInstance) CreateCheckoutSession() http.Handler {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userID, ok := ctx.Value(contextutil.UserIDKey).(string)
		_ = userID
		if !ok {
			logger.ErrorContext(ctx, "user ID not found in context")
			serverutil.WriteResponse(w, errors.New("failed to get user ID from context").Error(), http.StatusInternalServerError)
			return
		}

		eventID := r.PathValue("id")
		priceID, err := a.Repos.Event.GetPriceIDByEventID(ctx, eventID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to get priceID for event", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		_ = priceID
		// just to test things out, I created a product with a priceID to do thing with it.
		price_temp := "price_1OsTQfHwHXlEm0ohh1sSBXJa"
		sessionURL, err := a.Svcs.Stripe.CreateCheckoutSession(ctx, price_temp, 1)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create checkout session", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// NOTE: Can not redirect with that for some reason. So I want to send the url in json format. Yet this is in the tutorial
		// fmt.Println("Print the s.URL thing: ", s.URL)
		// http.Redirect(w, r, s.URL, http.StatusSeeOther)

		// NOTE: the old thing used to do that.
		// message := struct {
		// 	Url string `json:"url"`
		// }{
		// 	Url: sessionURL,
		// }
		// if err = serverutil.Encode(w, http.StatusInternalServerError, message); err != nil {
		type Response struct {
			URL string `json:"url"`
		}
		if err = serverutil.Encode(w, http.StatusInternalServerError, Response{URL: sessionURL}); err != nil {
			logger.ErrorContext(ctx, "failed to encode checkout session URL", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
