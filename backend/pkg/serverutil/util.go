package serverutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const SIGNINENDPOINT = "/signin"
const SIGNUPENDPOINT = "/signup"

type Validator interface {
	Valid(ctx context.Context) (problems map[string]string)
}

func Decode[T any](r *http.Request) (T, error) {
	var res T
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		return res, fmt.Errorf("decode json: %w", err)
	}
	return res, nil
}

func DecodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
	var v T
	if v, err := Decode[T](r); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}
	if pbms := v.Valid(r.Context()); len(pbms) > 0 {
		return v, pbms, fmt.Errorf("invalid %T: %d problems", v, len(pbms))
	}
	return v, nil, nil
}

func Encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(&v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
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

func GetSessionIDFromHeader(r *http.Request) string {
	header := r.Header["Authorization"][0]
	return strings.TrimPrefix(header, "Bearer ")
}
