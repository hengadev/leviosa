package productHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/product"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

func (a *AppInstance) CreateOffer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := contextutil.ValidateRoleInContext(ctx, models.ADMINISTRATOR); err != nil {
		logger.ErrorContext(ctx, "validate role from context", "error", err)
		http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
		return
	}

	product, err := serverutil.DecodeValid[productService.Offer](ctx, r.Body)
	if err != nil {
		switch {
		case errors.Is(err, serverutil.ErrDecodeJSON):
			logger.WarnContext(ctx, "decode product", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		default:
			logger.WarnContext(ctx, "invalid product creation", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}

	productID, err := a.Svcs.Stripe.CreateProduct(ctx, &product)
	if errors.Is(err, domain.ErrNotCreated) {
		logger.WarnContext(ctx, "stripe product not created")
		http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
		return
	}
	priceID, err := a.Svcs.Stripe.CreatePrice(ctx, productID, product.Price)
	if errors.Is(err, domain.ErrNotCreated) {
		logger.WarnContext(ctx, "stripe price not created")
		http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
		return
	}
	product.PriceID = priceID

	if err := a.Svcs.Product.CreateOffer(ctx, &product); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidValue):
			logger.WarnContext(ctx, "ivnalid product given")
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
		case errors.Is(err, domain.ErrNotCreated):
			logger.WarnContext(ctx, "product not created")
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrQueryFailed):
			logger.WarnContext(ctx, "database query create product failed")
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, rp.ErrContext):
			logger.WarnContext(ctx, "context error", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "product successfully created", http.StatusCreated)
}
