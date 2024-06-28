package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

// TODO: How to handle the webhook ?

func (s *Server) paymentHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	switch r.Method {
	case http.MethodOptions:
		enableJSON(&w)
		enableMethods(&w, http.MethodPost)
	case http.MethodPost:
		eventId := r.URL.Query().Get("eventId")
		fmt.Println("the eventId is : ", eventId)
		priceId := s.Store.GetPriceIDByEventID(eventId)
		fmt.Println("the priceId is : ", priceId)
		if err := createCheckoutSession(w, priceId); err != nil {
			log.Fatal("Failed to create session for the user")
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func createCheckoutSession(w http.ResponseWriter, price_id string) error {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	// TODO: Use the actual price_id in parameter instead of the temp one.
	_ = price_id
	price_temp := "price_1OsTQfHwHXlEm0ohh1sSBXJa"
	domain := os.Getenv("BASE_URL")
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(price_temp),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:         stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:   stripe.String(domain + "app/checkout/success/"),
		CancelURL:    stripe.String(domain + "app/checkout/cancel/"),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
	}
	s, err := session.New(params)
	if err != nil {
		return errors.New(fmt.Sprintf("Error with creating the new session Stripe : %v", err))
	}
	// NOTE: Can not redirect with that for some reason. So I want to send the url in json format. Yet this is in the tutorial
	// fmt.Println("Print the s.URL thing: ", s.URL)
	// http.Redirect(w, r, s.URL, http.StatusSeeOther)

	message := struct {
		Url string `json:"url"`
	}{
		Url: s.URL,
	}
	// NOTE: This is not how it is in the tutorial, they use the http.Redirect function with the s.URL (what is this)
	w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&message); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("Failed to encode message with the url - ", err)
	}
	return nil
}

func getEventName(event_date time.Time) string {
	return fmt.Sprintf("Ticket pour l'evenement du : %s", event_date.Format(time.RFC822))
}
