package handler

// NOTE: I get this from this link : https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/

import (
	// "github.com/GaryHY/event-reservation-app/api/middleware"
	// "github.com/GaryHY/event-reservation-app/internal/stores"
	// "github.com/GaryHY/event-reservation-app/util"
	// "github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
)

// NOTE: The api I want at the end of the project (or something like that)
// func addRoutes(
// 	mux                 *http.ServeMux,
// 	logger              *logging.Logger,
// 	config              Config,
// 	tenantsStore        *TenantsStore,
// 	commentsStore       *CommentsStore,
// 	conversationService *ConversationService,
// 	chatGPTService      *ChatGPTService,
// 	authProxy           *authProxy
// ) {
// 	mux.Handle("/api/v1/", handleTenantsGet(logger, tenantsStore))
// 	mux.Handle("/oauth2/", handleOAuth2Proxy(logger, authProxy))
// 	mux.HandleFunc("/healthz", handleHealthzPlease(logger))
// 	mux.Handle("/", http.NotFoundHandler())
// }

func (s *Server) addRoutes(
	mux *http.ServeMux,
	store Store,
	// logger  *logging.Logger
) {
	// for admins
	// NOTE: The first one without modification of the previous auth that I used.
	// mux.HandleFunc("/admin/events", auth(types.ADMIN, s.adminEventHandler, store))
	mux.HandleFunc("/admin/events", Auth(s.adminEventHandler, store))
	mux.HandleFunc("/admin/users", Auth(s.adminUsersHandler, store))
	mux.HandleFunc("/admin/payment", Auth(s.adminPaymentHandler, store))
	// mux.Handle("/admin/votes", Auth(types.ADMIN, adminVotesHandler, store))

	// for helpers
	mux.HandleFunc("/helper/photos", Auth(s.photosHandler, store))

	// for users
	// NOTE: the new way that I can use to right api
	// mux.HandleFunc("GET /event/{id}", Auth(s.getEventById, store)) // avec un get avec le query string et un get sans pour prendre tous les events d'un user
	// mux.HandleFunc("GET /events", Auth(s.getAllEvents, store))     // avec un get avec le query string et un get sans pour prendre tous les events d'un user
	mux.HandleFunc("/events", Auth(s.eventHandler, store)) // avec un get avec le query string et un get sans pour prendre tous les events d'un user
	// mux.Handle("/events", handleTenantsGet(store))                    // avec un get avec le query string et un get sans pour prendre tous les events d'un user
	mux.HandleFunc("/votes", Auth(s.votesHandler, store))
	mux.HandleFunc("/me", Auth(s.userHandler, store))
	// TODO: make a route /register/eventid? For- the user to pay for shit
	mux.HandleFunc("/checkout", Auth(s.paymentHandler, store))

	mux.HandleFunc("/signup", s.signUpHandler)
	mux.HandleFunc("/signin", s.signInHandler)
	mux.HandleFunc("/signout", Auth(s.signOutHandler, store))
}
