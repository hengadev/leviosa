package reviewHandler

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/google/uuid"
)

// GetProductReviews handles GET /products/{productId}/reviews. Public, unauthenticated:
// returns non-hidden reviews plus an aggregate rating for the product.
func (h *handler) GetProductReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	productID, err := uuid.Parse(r.PathValue("productId"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid product ID format"), http.StatusBadRequest)
		return
	}

	limit, offset := parsePagination(r)

	response, err := h.svc.GetPublicReviewsForProduct(ctx, productID, limit, offset)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get product reviews")
		return
	}

	httpx.RespondWithJSON(w, response, http.StatusOK)
}
