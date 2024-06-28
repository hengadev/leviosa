package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/types"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
)

// TODO: Toutes ces fonctions sont utilisees a la creation ou suppression d'evenement
func (s *Server) adminPaymentHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	enableJSON(&w)
	switch r.Method {
	case http.MethodPost:
		// TODO: find the event id that you want to remove from the request and the event date
		event_id := ""
		event_date := time.Now()
		_, err := createEventProductStripe(&w, event_id, event_date)
		if err != nil {
			log.Fatal("Failed to create session for the user")
		}
	case http.MethodDelete:
		// TODO: find the event id that you want to remove from the request
		event_id := ""
		if err := deleteEventProductStripe(event_id); err != nil {
			log.Fatal("Failed to create session for the user")
		}
		// TODO: Add the update product or is just delete then create the product
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// Function  that return a new Product Stripe associated with an event with the corresponding price
func createEventProductStripe(w *http.ResponseWriter, event_id string, event_date time.Time) (string, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	(*w).Header().Set("Authorization", fmt.Sprintf("Bearer %s", stripe.Key))
	product_params := &stripe.ProductParams{
		ID:          &event_id,
		Name:        stripe.String(getEventName(event_date)),
		Description: stripe.String("1 X Pass valuable for all the event."),
	}
	product, err := product.New(product_params)
	if err != nil {
		return "", errors.New(fmt.Sprintln("Failed to create the new product on the server - ", err))
	}
	price_params := &stripe.PriceParams{
		Currency:   stripe.String(string(stripe.CurrencyEUR)),
		Product:    stripe.String(product.ID),
		UnitAmount: stripe.Int64(types.EventPrice),
	}
	price, err := price.New(price_params)
	if err != nil {
		return "", errors.New(fmt.Sprintln("Failed to create the price on the server - ", err))
	}
	return price.ID, nil
}

// Function that gets the product by id and then delete it from the dashboard.
func deleteEventProductStripe(event_id string) error {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	product_params := &stripe.ProductParams{
		ID:          &event_id,
		Description: stripe.String("1 X Pass valuable for all the event."),
	}
	_, err := product.Del(event_id, product_params)
	if err != nil {
		return errors.New(fmt.Sprintln("Failed to delete the product on the server -", err))
	}
	return nil
}
