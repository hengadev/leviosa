package photohandler

import (
	"net/http"

	"github.com/GaryHY/leviosa/internal/server/app"
)

type AppInstance struct {
	*app.App
}

func (a *AppInstance) DeletePhoto() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func (a *AppInstance) UpdatePhoto() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func (a *AppInstance) ShowAllPhotosByUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
