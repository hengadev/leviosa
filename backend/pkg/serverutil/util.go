package serverutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

const SIGNINENDPOINT = "signin"
const SIGNUPENDPOINT = "signup"
const SIGNOUTENDPOINT = "signout"

type Validator interface {
	Valid(ctx context.Context) (problems errsx.Map)
}

var ErrDecodeJSON = errors.New("decoding json")

func NewDecodeJSONErr(err error) error {
	return fmt.Errorf("%w: %w", ErrDecodeJSON, err)
}

func Decode[T any](body io.ReadCloser) (T, error) {
	var res T
	defer body.Close()
	if err := json.NewDecoder(body).Decode(&res); err != nil {
		return res, NewDecodeJSONErr(err)
	}
	return res, nil
}

func DecodeValid[T Validator](ctx context.Context, body io.ReadCloser) (T, error) {
	var v T
	v, err := Decode[T](body)
	if err != nil {
		return v, err
	}
	if pbms := v.Valid(ctx); len(pbms) > 0 {
		return v, pbms
	}
	return v, nil
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
