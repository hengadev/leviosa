package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	// mw  "github.com/GaryHY/event-reservation-app/internal/middleware"
	"github.com/GaryHY/event-reservation-app/internal/types"
)

// TODO: Maybe make all the routes start with the /api because I already have /app.
func NewServer(
	store Store,
	photostore PhotoStore,
	// TODO: use that one later
	// logger middleware.Logger,
) *Server {
	server := new(Server)
	server.Store = store
	server.PhotoStore = photostore

	router := http.NewServeMux()

	// TODO: use the addroutes function
	server.addRoutes(router, store)

	// TODO: add the middleware here
	// var handler http.Handler = router
	// the logger implememts serveHTTP so it should be good
	// handler = middleware.NewLogger(handler)

	server.Handler = router

	return server
}

// TODO: can the middleware that I have be a method of server since I need to use the store ?
// NOTE: try to use the thing with the most simple ways again friend
// func someMiddleWare(h http.Handler) http.Handler {
func someMiddleWare(h types.Handler) types.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Use the store to get somehting from this
		fmt.Println("I just used the middleware to see the point")
		h(w, r)
	}

}

// NOTE: to remove
// type Handler = func(w http.ResponseWriter, r *http.Request)
//
// func auth(role types.Role, h Handler, store Store) Handler {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		sessionId := parseAuthorizationHeader(r)
// 		if sessionId == "" {
// 			WriteResponse(w, "No session id in the request, you need to login again", http.StatusUnauthorized)
// 			return
// 		}
// 		if !store.HasSession(sessionId) {
// 			WriteResponse(w, "Session has expired, you need to login again.", http.StatusUnauthorized)
// 			return
// 		}
// 		if !store.Authorize(sessionId, role) {
// 			WriteResponse(w, "The user does not have right to access this ressource.", http.StatusUnauthorized)
// 			return
// 		}
// 		h(w, r)
// 	}
// }

type Server struct {
	Store
	http.Handler
	PhotoStore
	// TODO: make the implementation of that using redis.
	// SessionManager SessionManager
}

type PhotoStore interface {
	PostFile(file multipart.File, filename, event_id string)
}

// TODO: make the implementation of that using redis.
// type SessionManager interface {
// 	GetUserFromSessionId(sessionId string) types.UserSent
// }

// Put that in its own file
type Store interface {
	GetEventByID(id string) *types.Event
	GetEventByUserId(userId string) []*types.Event
	GetEventForUser(userId string) (types.EventBody, error)
	GetAllEvents() []*types.Event
	PostEvent(event *types.Event)
	DeleteEvent(eventId string) error
	UpdateEvent(event *types.Event) error
	GetUserId(user_email string) string
	// TODO: use redis to get that faster. The next function should return an error
	GetUserIdBySessionId(session_id string) (string, error)
	CreateUser(newUser *types.UserStored) error
	CheckUser(email string) bool
	CheckUserById(userId string) bool
	GetAllUsers() []*types.UserSent
	GetUserFromSessionId(sessionId string) *types.UserSent
	DeleteUser(userId string) error
	UpdateUser(user *types.UserStored) error
	IsAdmin(sessionId string) bool
	GetHashPassword(user *types.User) (hashpassword string)
	CreateSession(newSession *types.Session) error
	HasSession(sessionId string) bool
	DeleteSession(session *types.Session) error
	DeleteSessionByID(id string) error
	CreateVote(userId, days string, month, year int) error
	DecreaseEventPlacecount(event_id string) error
	CheckEvent(eventId string) bool
	// NOTE: new vote api
	GetVoteByUserId(month, year int, userId string) ([]*types.VoteSent, error)
	GetAvailableVotes(month, year int) ([]*types.VoteSent, error)
	CreateVoteByUserId(votes []*types.VoteSent, userId string) error
	GetNextVotes() ([]*types.NextVote, error)
	HasVote(month, year int, userId string) bool
	RemoveVoteByUserId(month, year int, userId string) error
	// TODO: remove that thing because the API is not working like that anymore
	CheckVote(userId, eventId *string) bool
	CheckVoteById(voteId *string) bool
	DeleteVote(voteId *string) error
	Authorize(sessionId string, role types.Role) bool
	GetPriceIDByEventID(eventId string) (priceId string)
	GetBeginAtByEventId(eventId string) time.Time
	CheckRegistration(registration *types.Registration) bool
	CreateRegistration(registration *types.Registration) error
}
