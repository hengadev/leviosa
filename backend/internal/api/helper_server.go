package api

import (
	"encoding/json"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
	"strings"
)

// TODO: Put that in a .env file
const (
	FRONTENDORIGIN = "http://localhost:4321"
)

// A function to enable the frontend (at http://localhost:4321 for testing) to access the endpoint where it is called from.
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", FRONTENDORIGIN)
}

func enableJSON(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func enableMethods(w *http.ResponseWriter, methods ...string) {
	(*w).Header().Set("Access-Control-Allow-Methods", strings.Join(methods, " "))
}

func enableHeaders(w *http.ResponseWriter, headers ...string) {
	(*w).Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
}

// TODO: Can I use generic for that ?
func getUserFromRequest(r *http.Request) (user *types.User) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return &types.User{}
	}
	return
}

func getUserFormFromRequest(r *http.Request) (user *types.UserForm) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return &types.UserForm{}
	}
	return
}

func getUserStoredFromRequest(r *http.Request) (user *types.UserStored) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return &types.UserStored{}
	}
	return
}

func getEventFromRequest(w http.ResponseWriter, r *http.Request) (event types.Event) {
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return types.Event{}
	}
	return
}
