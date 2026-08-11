package reviewHandler

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware/auth"
)

type Handler interface {
	RegisterRoutes(router *http.ServeMux)
	CreateReview(w http.ResponseWriter, r *http.Request)
	GetPartnerReviews(w http.ResponseWriter, r *http.Request)
	GetProductReviews(w http.ResponseWriter, r *http.Request)
	HideReview(w http.ResponseWriter, r *http.Request)
	UnhideReview(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	svc    ports.ReviewService
	authmw auth.AuthMiddleware
}

func New(svc ports.ReviewService, authmw auth.AuthMiddleware) Handler {
	return &handler{svc: svc, authmw: authmw}
}
