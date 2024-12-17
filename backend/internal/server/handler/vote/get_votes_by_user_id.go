package vote

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// Function that get all the votes for a user.
func (a *AppInstance) GetVotesByUserID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get userID from context
		// NOTE: the old way but it does not seem to work
		// userID := ctx.Value(mw.SessionIDKey).(string)
		userID := ctx.Value(mw.UserIDKey).(int)
		// use pathValues to get all values.
		month := r.PathValue("month")
		year := r.PathValue("year")
		if month == "" || year == "" {
			err := fmt.Errorf("month or year are not well formatted")
			slog.ErrorContext(ctx, "failed to parse month or year", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// get votes
		votes, err := a.Svcs.Vote.GetVotesByUserID(ctx, month, year, userID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the votes from database", "error", err)
			http.Error(w, fmt.Sprintf("Failed to get the data from the database - %s", err), http.StatusInternalServerError)
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
