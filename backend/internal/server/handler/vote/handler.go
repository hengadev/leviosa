package vote

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/internal/server/service"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

type Handler struct {
	*handler.Handler
}

func NewHandler(handler *handler.Handler) *Handler {
	return &Handler{handler}
}

// Function that get all the votes for a user.
func (h *Handler) GetVotesByUserID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get that from the context
		userID := ctx.Value(mw.SessionIDKey).(string)
		// use the pathValues to get all these values.
		month := r.PathValue("month")
		year := r.PathValue("year")
		if month == "" || year == "" {
			err := fmt.Errorf("month or year are not well formatted")
			slog.ErrorContext(ctx, "failed to parse month or year", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// get the votes
		votes, err := h.Svcs.Vote.GetVotesByUserID(ctx, month, year, userID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the votes from database", "error", err)
			http.Error(w, fmt.Sprintf("Failed to get the data from the database - %s", err), http.StatusInternalServerError)
			return
		}
		// encode the result to the user
		if err := serverutil.Encode(w, http.StatusFound, votes); err != nil {
			slog.ErrorContext(ctx, "failed to encode the votes", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func (h *Handler) GetNextVotes() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// specify the time info for the fetching
		now := time.Now()
		month, year := int(now.Month()), int(now.Year())
		// get the votes
		votes, err := h.Repos.Vote.GetNextVotes(ctx, month, year)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the votes from database", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// encode the result to the user
		if err := serverutil.Encode(w, http.StatusFound, votes); err != nil {
			slog.ErrorContext(ctx, "failed to encode the votes", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

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
		// add the userID field to the votes.
		for _, vote := range votes {
			userIDInt, err := strconv.Atoi(userID)
			if err != nil {
				slog.ErrorContext(ctx, "failed to convert userID to int", "error", err)
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
			vote.UserID = userIDInt
		}
		// create the vote
		if err := h.Svcs.Vote.CreateVote(ctx, votes); err != nil {
			slog.ErrorContext(ctx, "failed to create vote", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return

		}
	})
}

// old api

// // TODO: change that function, the vote should be submitted to the Vote[Month][Year] table.
// func OldMakeVote() http.Handler {
// 	// use the sessionID from the application
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		ctx, cancel := context.WithCancel(r.Context())
// 		defer cancel()
// 		// get userID from context
// 		userID := ctx.Value(mw.SessionIDKey)
// 		// TODO: change that type brother
// 		// get the votes from the request body
// 		reqBody := []*types.VoteSent{}
// 		votes, err := serverutil.Decode[*types.VoteSent](r)
// 		if err != nil {
// 			slog.ErrorContext(ctx, "failed to decode the votes from request body", "error", err)
// 			http.Error(w, handler.NewInternalErr(err), http.StatusBadRequest)
// 			return
// 		}
//
// 		if s.Store.HasVote(reqBody[0].Month, reqBody[0].Year, userID) { // get the month and the year in this
// 			err := s.Store.RemoveVoteByUserID(reqBody[0].Month, reqBody[0].Year, userId)
// 			if err != nil {
// 				fmt.Println("failed to remove the vote ! - ", err)
// 				WriteResponse(w, fmt.Sprintf("Failed to remove previous vote from the database- %s", err), http.StatusInternalServerError)
// 				return
// 			}
// 		}
// 		// TODO: That thing is not working, why ? (before I use the writeResponse at the end.)
// 		err = s.Store.CreateVoteByUserId(reqBody, userId)
// 		if err != nil {
// 			WriteResponse(w, fmt.Sprintf("Failed to create the vote in the database - %s", err), http.StatusInternalServerError)
// 			return
// 		}
// 		WriteResponse(w, fmt.Sprintf("Vote created successfully"), http.StatusCreated)
// 	})
// }
