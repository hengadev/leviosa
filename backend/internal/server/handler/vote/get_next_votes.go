package vote

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

func (a *AppInstance) GetNextVotes() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// specify time info for fetching
		now := time.Now()
		month, year := int(now.Month()), int(now.Year())
		// get votes
		votes, err := a.Repos.Vote.GetNextVotes(ctx, month, year)
		if err != nil {
			switch {
			default:
				logger.ErrorContext(ctx, "failed to get the votes from database", "error", err)
				serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			}
			return
		}
		if err := serverutil.Encode(w, int(http.StatusOK), votes); err != nil {
			logger.WarnContext(ctx, "failed to encode the votes", "error", err)
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
