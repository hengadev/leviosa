package vote

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// Function that create or update vote
func (h *Handler) MakeVote() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get userID from cookie
		userID := ctx.Value(mw.SessionIDKey).(string)
		// get votes from request
		votes, err := serverutil.Decode[[]*vote.Vote](r)
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode vote", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// add userID field to votes.
		for _, vote := range votes {
			userIDInt, err := strconv.Atoi(userID)
			if err != nil {
				slog.ErrorContext(ctx, "failed to convert userID to int", "error", err)
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
			vote.UserID = userIDInt
		}
		// create vote
		if err := h.Svcs.Vote.CreateVote(ctx, votes); err != nil {
			slog.ErrorContext(ctx, "failed to create vote", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return

		}
	})
}
