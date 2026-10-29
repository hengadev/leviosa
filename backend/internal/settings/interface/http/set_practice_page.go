package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
)

func (h *handler) SetPracticePagePartner(w http.ResponseWriter, r *http.Request) {
	setPracticePageSetting(w, r, domain.SetPracticePagePartnerRequest{},
		"set_practice_page_partner", "set practice page partner",
		func(ctx context.Context, request *domain.SetPracticePagePartnerRequest) (*domain.SetPracticePagePartnerResponse, error) {
			return h.svc.SetPracticePagePartner(ctx, request)
		})
}

func (h *handler) SetPracticePageProduct(w http.ResponseWriter, r *http.Request) {
	setPracticePageSetting(w, r, domain.SetPracticePageProductRequest{},
		"set_practice_page_product", "set practice page product",
		func(ctx context.Context, request *domain.SetPracticePageProductRequest) (*domain.SetPracticePageProductResponse, error) {
			return h.svc.SetPracticePageProduct(ctx, request)
		})
}

func (h *handler) SetPracticePageVideo(w http.ResponseWriter, r *http.Request) {
	setPracticePageSetting(w, r, domain.SetPracticePageVideoRequest{},
		"set_practice_page_video", "set practice page video",
		func(ctx context.Context, request *domain.SetPracticePageVideoRequest) (*domain.SetPracticePageVideoResponse, error) {
			return h.svc.SetPracticePageVideo(ctx, request)
		})
}

// setPracticePageSetting decodes one practice page setting request and runs
// its use case, with the same content-type check, logging and error mapping
// as the other settings setters.
func setPracticePageSetting[T any, R any](
	w http.ResponseWriter,
	r *http.Request,
	_ T,
	operation, action string,
	call func(ctx context.Context, request *T) (R, error),
) {
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

	logger.InfoContext(ctx, "Handler: Processing "+action+" request",
		"operation", operation,
		"method", r.Method,
		"path", r.URL.Path)

	var request T
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		logger.DebugContext(ctx, fmt.Sprintf("Handler: Error decoding JSON body: %v", err))
		httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("invalid request body: %v", err)), http.StatusBadRequest)
		return
	}

	response, err := call(ctx, &request)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, action)
		return
	}

	logger.InfoContext(ctx, "Handler: "+action+" completed",
		"operation", operation,
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, response, http.StatusOK)
}
