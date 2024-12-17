package vote

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) GetNextVotes() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// specify time info for fetching
		now := time.Now()
		month, year := int(now.Month()), int(now.Year())
		// get votes
		votes, err := a.Repos.Vote.GetNextVotes(ctx, month, year)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the votes from database", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// encode result to user
		if err := serverutil.Encode(w, http.StatusFound, votes); err != nil {
			slog.ErrorContext(ctx, "failed to encode the votes", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
