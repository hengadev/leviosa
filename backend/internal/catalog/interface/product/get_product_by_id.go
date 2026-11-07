package productHandler

import (
	"errors"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

func (h *handler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing get product by id",
		"operation", "get_product_by_id",
		"method", r.Method,
		"path", r.URL.Path)

	// Served on /products/{id} and /admin/products/{id}: the id comes from the route.
	productID := r.PathValue("id")
	if productID == "" {
		httpx.RespondWithError(w, errors.New("product ID is missing from the URL"), http.StatusBadRequest)
		return
	}

	product, err := h.aggr.GetProductByID(ctx, productID)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get product by ID")
		return
	}

	logger.InfoContext(ctx, "Handler: get product by id completed",
		"operation", "get_product_by_id",
		"product_id", productID,
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, product, http.StatusOK)
}
