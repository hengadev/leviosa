package vote

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) GetNextVotes() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// specify time info for fetching
		now := time.Now()
		month, year := int(now.Month()), int(now.Year())
		// get votes
		votes, err := a.Repos.Vote.GetNextVotes(ctx, month, year)
		if err != nil {
			logger.ErrorContext(ctx, "failed to get the votes from database", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// encode result to user
		if err := serverutil.Encode(w, http.StatusFound, votes); err != nil {
			logger.ErrorContext(ctx, "failed to encode the votes", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
