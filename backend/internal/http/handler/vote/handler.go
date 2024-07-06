package votehandler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	"github.com/GaryHY/event-reservation-app/internal/http/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/http/middleware"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// Function that get all the votes for a user.
func GetVotesByUserID(v *vote.Service) http.Handler {
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
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// get the votes
		votes, err := v.GetVotesByUserID(ctx, month, year, userID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the votes from database", "error", err)
			http.Error(w, fmt.Sprintf("Failed to get the data from the database - %s", err), http.StatusInternalServerError)
			return
		}
		// encode the result to the user
		if err := serverutil.Encode(w, http.StatusFound, votes); err != nil {
			slog.ErrorContext(ctx, "failed to encode the votes", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func GetNextVotes(v vote.Reader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get the votes
		votes, err := v.GetNextVotes(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the votes from database", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// encode the result to the user
		if err := serverutil.Encode(w, http.StatusFound, votes); err != nil {
			slog.ErrorContext(ctx, "failed to encode the votes", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

// Function that create or update vote
func MakeVote(v *vote.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get the userID
		userID := ctx.Value(mw.SessionIDKey).(string)
		// get the votes from the client
		votes, err := serverutil.Decode[[]*vote.Vote](r)
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode vote", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// add the userID field to the votes.
		for _, vote := range votes {
			vote.UserID = userID
		}
		// create the vote
		if err := v.CreateVote(ctx, votes); err != nil {
			slog.ErrorContext(ctx, "failed to create vote", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return

		}
	})
}

// old api

// TODO: change that function, the vote should be submitted to the Vote[Month][Year] table.
func OldMakeVote() http.Handler {
	// use the sessionID from the application
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get userID from context
		userID := ctx.Value(mw.SessionIDKey)
		// TODO: change that type brother
		// get the votes from the request body
		reqBody := []*types.VoteSent{}
		votes, err := serverutil.Decode[*types.VoteSent](r)
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode the votes from request body", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusBadRequest)
			return
		}

		if s.Store.HasVote(reqBody[0].Month, reqBody[0].Year, userID) { // get the month and the year in this
			err := s.Store.RemoveVoteByUserID(reqBody[0].Month, reqBody[0].Year, userId)
			if err != nil {
				fmt.Println("failed to remove the vote ! - ", err)
				WriteResponse(w, fmt.Sprintf("Failed to remove previous vote from the database- %s", err), http.StatusInternalServerError)
				return
			}
		}
		// TODO: That thing is not working, why ? (before I use the writeResponse at the end.)
		err = s.Store.CreateVoteByUserId(reqBody, userId)
		if err != nil {
			WriteResponse(w, fmt.Sprintf("Failed to create the vote in the database - %s", err), http.StatusInternalServerError)
			return
		}
		WriteResponse(w, fmt.Sprintf("Vote created successfully"), http.StatusCreated)
	})
}
