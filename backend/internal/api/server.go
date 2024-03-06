package api

import (
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"mime/multipart"
	"net/http"
)

const (
	YYYYMMDD = "2006-01-02"
)

// TODO: Think if you prefer the singular over the plural for the endpoints

func NewServer(store Store, photostore PhotoStore) *Server {
	server := new(Server)
	server.Store = store
	server.PhotoStore = photostore

	router := http.NewServeMux()

	router.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Quelqu'un est sur le serveur")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Je suis la\n"))
	}))

	// for admins
	// TODO: Peut etre mettre un "events" au lieu de event
	router.Handle("/admin/event", http.HandlerFunc(server.adminEventHandler))
	// router.Handle("/admin/votes", http.HandlerFunc(server.adminVotesHandler))
	router.Handle("/admin/users", http.HandlerFunc(server.adminUsersHandler))

	// Pour gerer le post de photo dans mon stockage sur S3
	router.Handle("/photos", http.HandlerFunc(server.photosHandler))

	// for users
	router.Handle("/event", http.HandlerFunc(server.eventHandler)) // avec un get avec le query string et un get sans pour prendre tous les events d'un user
	router.Handle("/votes", http.HandlerFunc(server.votesHandler))
	// TODO: Which one should I use ?
	router.Handle("/checkout", http.HandlerFunc(server.paymentHandler))
	router.Handle("/status/checkout/success", http.HandlerFunc(server.paymentHandler))
	router.Handle("/status/checkout/fail", http.HandlerFunc(server.paymentHandler))

	router.Handle("/signup", http.HandlerFunc(server.signUpHandler))
	router.Handle("/signin", http.HandlerFunc(server.signInHandler))
	router.Handle("/signout", http.HandlerFunc(server.signOutHandler))

	server.Handler = router

	return server
}

type Server struct {
	Store Store
	http.Handler
	PhotoStore PhotoStore
}

type PhotoStore interface {
	PostFile(file multipart.File, filename, event_id string)
}

type Store interface {
	GetEventByID(id string) *types.Event
	GetEventByUserId(user_id string) []*types.Event
	GetAllEvents() []*types.Event
	PostEvent(event *types.Event)
	DeleteEvent(event_id string) error
	UpdateEvent(event *types.Event) error
	GetUserId(user_email string) string
	GetUserIdBySessionId(session_id string) string
	CreateUser(newUser *types.UserStored) error
	CheckUser(email string) bool
	CheckUserById(user_id string) bool
	GetAllUsers() []*types.UserStored
	DeleteUser(user_id string) error
	UpdateUser(user *types.UserStored) error
	IsAdmin(session_id string) bool
	GetHashPassword(user *types.User) (hashpassword string)
	CreateSession(newSession *types.Session) error
	HasSession(email string) bool
	DeleteSession(session *types.Session) error
	DeleteSessionByID(id string) error
	CreateVote(newVote *types.Vote) error
	DecreaseEventPlacecount(event_id string) error
	CheckEvent(event_id string) bool
	CheckVote(userId, eventId *string) bool
	CheckVoteById(voteId *string) bool
	DeleteVote(voteId *string) error
	Authorize(session_id string, role types.Role) bool
}
