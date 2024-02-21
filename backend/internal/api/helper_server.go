package api

import (
	"encoding/json"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
)

func getUserFromRequest(w http.ResponseWriter, r *http.Request) (user *types.User) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return &types.User{}
	}
	return
}

func getUserFormFromRequest(w http.ResponseWriter, r *http.Request) (user *types.UserForm) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return &types.UserForm{}
	}
	return
}

func getUserStoredFromRequest(w http.ResponseWriter, r *http.Request) (user *types.UserStored) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
