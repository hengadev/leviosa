package productHandler

import (
	"errors"
	"fmt"
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

func (a *AppInstance) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := contextutil.ValidateRoleInContext(ctx, models.ADMINISTRATOR); err != nil {
		logger.ErrorContext(ctx, "validate role from context", "error", err)
		serverutil.WriteResponse(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
		return
	}

	product, err := serverutil.DecodeValid[productService.Product](ctx, r.Body)
	if err != nil {
		switch {
		case errors.Is(err, serverutil.ErrDecodeJSON):
			logger.WarnContext(ctx, "decode product", "error", err)
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		default:
			logger.WarnContext(ctx, "invalid product creation", "error", err)
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}

	// TODO: now get the product from this thing
	err = a.Svcs.Product.UpdateProduct(ctx, &product)
	if err != nil {
		// TODO: handle the remaining the update product
		switch {
		case errors.Is(err, domain.ErrNotFound):
		case errors.Is(err, rp.ErrContext):
		}
		return
	}

	serverutil.WriteResponse(w, fmt.Sprintf("product with name %s udpated", product.Name), http.StatusNoContent)
}
