package bookingHandler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"

	"github.com/google/uuid"
)

// maxSlotRangeDays bounds one slot query; the policy horizon narrows it
// further (at most 52 weeks).
const maxSlotRangeDays = 400

// GetFreeSlots is the public slot endpoint (docs/issues/010-slot-engine.md):
//
//	GET /slots?partner_id=<uuid>&product_id=<uuid>&from=YYYY-MM-DD&to=YYYY-MM-DD
//
// from and to are calendar dates in the practice's time zone (Europe/Paris),
// both inclusive. The answer lists the free slots the slot engine offers in
// that range, plus the earliest start and horizon end the booking policy
// allows, so the calendar can stop paging there. No authentication: the
// practice page is public and the answer carries no personal data.
func (h *handler) GetFreeSlots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	partnerID, err := uuid.Parse(query.Get("partner_id"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("partner_id: must be a UUID"), http.StatusBadRequest)
		return
	}
	productID, err := uuid.Parse(query.Get("product_id"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("product_id: must be a UUID"), http.StatusBadRequest)
		return
	}
	fromDate, err := time.ParseInLocation(time.DateOnly, query.Get("from"), domain.PracticeLocation)
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("from: must be a date (YYYY-MM-DD)"), http.StatusBadRequest)
		return
	}
	toDate, err := time.ParseInLocation(time.DateOnly, query.Get("to"), domain.PracticeLocation)
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("to: must be a date (YYYY-MM-DD)"), http.StatusBadRequest)
		return
	}
	if toDate.Before(fromDate) {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("to: must not be before from"), http.StatusBadRequest)
		return
	}
	// AddDate keeps local midnight across a daylight-saving change.
	toExclusive := toDate.AddDate(0, 0, 1)
	if toExclusive.After(fromDate.AddDate(0, 0, maxSlotRangeDays)) {
		httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("range: at most %d days", maxSlotRangeDays)), http.StatusBadRequest)
		return
	}

	slots, err := h.svc.GetFreeSlots(ctx, partnerID, productID, fromDate, toExclusive)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, errs.ErrRepositoryNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, errs.ErrInvalidValue):
			statusCode = http.StatusBadRequest
		}
		httpx.RespondWithError(w, err, statusCode)
		return
	}

	httpx.RespondWithJSON(w, slots.ToResponse(), http.StatusOK)
}
