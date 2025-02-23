package productHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

func (a *AppInstance) GetProduct(w http.ResponseWriter, r *http.Request) {
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

	productID := r.URL.Query().Get("id")
	if productID == "" {
		err := errors.New("product ID should not be empty")
		logger.WarnContext(ctx, err.Error())
		http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
	}

	product, err := a.Svcs.Product.GetProduct(ctx, productID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
		case errors.Is(err, rp.ErrContext):
		}
		return
	}

	if err := serverutil.Encode(w, int(http.StatusOK), *product); err != nil {
		logger.ErrorContext(ctx, "failed to encode product with provided ID", "error", err)
		http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
	}
}
