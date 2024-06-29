package serverutil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// func Decode[T any](r io.Reader) (*T, error) {
func Decode[T any](r *http.Request) (T, error) {
	var res T
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		return res, fmt.Errorf("decode json: %w", err)
	}
	return res, nil
}

func Encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(&v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
