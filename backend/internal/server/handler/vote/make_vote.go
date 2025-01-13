package vote

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/leviosa/internal/domain/vote"
	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

// Function that create or update vote
func (a *AppInstance) MakeVote() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userID, ok := ctx.Value(contextutil.UserIDKey).(string)
		if !ok {
			logger.ErrorContext(ctx, "user ID not found in context")
			serverutil.WriteResponse(w, errors.New("failed to get user ID from context").Error(), http.StatusInternalServerError)
			return
		}

		// get votes from request
		votes, err := serverutil.Decode[[]*vote.Vote](r.Body)
		if err != nil {
			logger.ErrorContext(ctx, "failed to decode vote", "error", err)
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// add userID field to votes.
		for _, vote := range votes {
			vote.UserID = userID
		}
		// create vote
		if err := a.Svcs.Vote.CreateVote(ctx, votes); err != nil {
			logger.ErrorContext(ctx, "failed to create vote", "error", err)
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return

		}
	})
}
