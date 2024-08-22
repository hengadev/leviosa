package userHandler

import (
	"net/http"
)

func (h *Handler) DeleteUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		print("delete the user handler")
	})
}
