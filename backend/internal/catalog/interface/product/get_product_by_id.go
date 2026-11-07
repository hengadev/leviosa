package productHandler

import (
	"errors"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

// GetProductByID serves /admin/products/{id}: any product, whatever its status.
func (h *handler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	h.getProductByID(w, r, false)
}

// GetPublishedProductByID serves the public /products/{id}: a draft or
// archived product answers 404, as if it did not exist.
func (h *handler) GetPublishedProductByID(w http.ResponseWriter, r *http.Request) {
	h.getProductByID(w, r, true)
}

func (h *handler) getProductByID(w http.ResponseWriter, r *http.Request, publishedOnly bool) {
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
	if publishedOnly && product.Product.Status != domain.Published {
		err := errs.NewNotFoundErr(errors.New("product is not published"), "product")
		httpx.RespondWithServiceError(w, logger, ctx, err, "get product by ID")
		return
	}

	logger.InfoContext(ctx, "Handler: get product by id completed",
		"operation", "get_product_by_id",
		"product_id", productID,
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, product, http.StatusOK)
}
