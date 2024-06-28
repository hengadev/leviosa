package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/GaryHY/event-reservation-app/internal/types"
)

// TODO: Put that in a .env file
const (
	FRONTENDORIGIN = "http://localhost:5173"
)

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
// Yes I can, and this the function encode at the end of the file.
// -> Use  the decode function for this or it doesnt make sense
func getUserFromRequest(r *http.Request) (user *types.User) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return &types.User{}
	}
	defer r.Body.Close()
	return
}

func getUserFormFromRequest(r *http.Request) (user *types.UserForm) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return &types.UserForm{}
	}
	defer r.Body.Close()
	return
}

func getUserStoredFromRequest(r *http.Request) (user *types.UserStored) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return &types.UserStored{}
	}
	defer r.Body.Close()
	return
}

func getEventFromRequest(w http.ResponseWriter, r *http.Request) (event types.Event) {
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return types.Event{}
	}
	defer r.Body.Close()
	return
}

func parseAuthorizationHeader(r *http.Request) string {
	header := r.Header["Authorization"][0]
	return strings.TrimPrefix(header, "Bearer ")
}

func WriteResponse(w http.ResponseWriter, message string, status int) error {
	resBody := struct {
		Message string `json:"message"`
	}{message}
	// TODO: that gives an error for whatever reason
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	defer r.Body.Close()
	return v, nil
}
