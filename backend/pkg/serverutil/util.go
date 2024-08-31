package serverutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const SIGNINENDPOINT = "signin"
const SIGNUPENDPOINT = "signup"
const SIGNOUTENDPOINT = "signout"

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
	v, err := Decode[T](r)
	if err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}
	if pbms := v.Valid(r.Context()); len(pbms) > 0 {
		err := FormatError(pbms, fmt.Sprintf("%T", v))
		return v, pbms, fmt.Errorf("invalid %T with %d problems : %w", v, len(pbms), err)
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
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

func FormatError(pbms map[string]string, name string) error {
	var temp string
	for field, pbm := range pbms {
		temp += fmt.Sprintf("invalid %s: %s, ", field, pbm)
	}
	return errors.New(fmt.Sprintf("%s error : [%s]", name, temp))
}
