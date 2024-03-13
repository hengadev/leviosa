package api

import (
	"errors"
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	"log"
	"net/http"
	"os"
)

func (s *Server) paymentHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	switch r.Method {
	case "OPTIONS": // preflight request
		w.Header().Set("Access-Control-Allow-Methods", "GET")
	case http.MethodGet:
		if err := createCheckoutSession(w, r); err != nil {
			log.Fatal("Failed to create session")
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// TODO: : From the bendavis video about stripe.
	// 1. recoit une requete avec un json contenant le produit que je vuex acheter
	// 2. fais une requete au sdk de stripe
	// 3. le sdk de stripe me repond avec les informations sur une nouvelle checkout session et un url vers lequel je redirige le user
	// 4. redirect the user to the checkout page with the url I got from the previous response
	// 5. redirect to some succes or failure page depending on the income of the checkout action
	// 6. stripe make a request to our request handler that we are going to fulfill on our hand
}

func createCheckoutSession(w http.ResponseWriter, r *http.Request) error {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	// TODO: Get the price from the store
	product_id := "price_1OsTQfHwHXlEm0ohh1sSBXJa"
	domain := os.Getenv("HOST")
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				Price:    stripe.String(product_id),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		// SuccessURL:   stripe.String(domain + "/success.html"),
		// CancelURL:    stripe.String(domain + "/cancel.html"),
		SuccessURL:   stripe.String(domain + "/home/"),
		CancelURL:    stripe.String(domain + "/events/"),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
	}

	s, err := session.New(params)

	if err != nil {
		return errors.New(fmt.Sprintf("Error with creating the session.New: %v", err))
	}

	w.Header().Set("Content-Type", "null")
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
	return nil
}

func getEventName(event_date string) string {
	return fmt.Sprintf("Ticket for event ocurring on : %s", event_date)
}

// Function  that return a new Product Stripe associated with an event with the corresponding price
func createEventProductStripe(event_id, event_date string) error {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	product_params := &stripe.ProductParams{
		ID:          &event_id,
		Name:        stripe.String(getEventName(event_date)),
		Description: stripe.String("1 X Pass valuable for all the event."),
	}

	new_product, err := product.New(product_params)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create the new product on the server -", err))
	}

	price_params := &stripe.PriceParams{
		Currency:   stripe.String(string(stripe.CurrencyEUR)),
		Product:    stripe.String(new_product.ID),
		UnitAmount: stripe.Int64(types.EventPrice),
	}
	new_price, _ := price.New(price_params)

	fmt.Println("Success! Here is your starter subscription product id: " + new_product.ID)
	fmt.Println("Success! Here is your starter subscription price id: " + new_price.ID)
	return nil
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
		return errors.New(fmt.Sprintf("Failed to delete the product on the server -", err))
	}
	return nil
}
