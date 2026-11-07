package priceHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

// ReplaceOneTimePrice handles PUT /admin/products/{id}/price: the product's
// price changes, through a new Stripe price replacing the active one.
func (h *handler) ReplaceOneTimePrice(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		httpx.RespondWithError(w, errors.New("unsupported media type: please send 'application/json'"), http.StatusUnsupportedMediaType)
		return
	}

	ctx := r.Context()
	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	productID := r.PathValue("id")

	var request domain.CreatePriceRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("invalid request body: %v", err)), http.StatusBadRequest)
		return
	}

	priceID, err := h.svc.ReplaceOneTimePrice(ctx, productID, &request)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "replace one-time price")
		return
	}

	logger.InfoContext(ctx, "Handler: replace one-time price completed",
		"operation", "replace_one_time_price",
		"product_id", productID,
		"price_id", priceID)

	httpx.RespondWithJSON(w, struct {
		PriceID string `json:"price_id"`
	}{PriceID: priceID}, http.StatusOK)
}
