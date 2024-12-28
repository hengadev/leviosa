package productHandler

import (
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/product"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) CreateProduct() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if the user is an admin ?
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := contextutil.ValidateRoleInContext(ctx, models.ADMINISTRATOR); err != nil {
			logger.ErrorContext(ctx, "validate role from context", "error", err)
			serverutil.WriteResponse(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// TODO: decode the product that we need

		product, err := serverutil.DecodeValid[productService.Product](ctx, r.Body)
		if err != nil {
			switch {
			}
		}

		if err := a.Svcs.Product.CreateProduct(ctx, &product); err != nil {
			switch {

			}
		}
		productID, err := a.Svcs.Stripe.CreateProduct(ctx, &product)
		if err != nil {
			switch {
			}
		}
		priceID, err := a.Svcs.Stripe.CreatePrice(ctx, productID, product.Price)
		if err != nil {
			switch {
			}
		}
		if err := a.Svcs.Product.AddPriceID(ctx, productID, priceID); err != nil {
			switch {
			}
		}
		w.WriteHeader(http.StatusCreated)
	})
}
