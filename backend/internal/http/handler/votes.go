package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/GaryHY/event-reservation-app/internal/types"
)

func (s *Server) votesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	sessionId := parseAuthorizationHeader(r)
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodOptions:
		enableJSON(&w)
		enableMethods(&w, http.MethodDelete, http.MethodPost, http.MethodGet, http.MethodPut)
	case http.MethodGet:
		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			WriteResponse(w, fmt.Sprintf("Error while processing the url queries - %s", err), http.StatusBadRequest)
			return
		}
		if len(queryParams) != 0 {
			s.getVote(w, sessionId, queryParams)
		} else {
			s.getNextVotes(w)
		}
	case http.MethodPost:
		s.makeVote(w, r, sessionId)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (s *Server) getVote(w http.ResponseWriter, sessionId string, queryParams url.Values) {
	userId, err := s.Store.GetUserIdBySessionId(sessionId)
	if err != nil {
		WriteResponse(w, "Unable to get the events from the database", http.StatusInternalServerError)
		return
	}
	month, err := strconv.Atoi(queryParams["month"][0])
	if err != nil {
		WriteResponse(w, fmt.Sprintf("Failed to parse the month from the queryy parameters - %s", err), http.StatusInternalServerError)
		return
	}
	year, err := strconv.Atoi(queryParams["year"][0])
	if err != nil {
		WriteResponse(w, fmt.Sprintf("Failed to parse the year from the queryy parameters - %s", err), http.StatusInternalServerError)
		return
	}
	votes, err := s.Store.GetVoteByUserId(month, year, userId)
	// TODO: handle the case with the ErrNotFound
	if err != nil {
		// NOTE: better to do that since it is the idiomatic way
		// http.Error(w, fmt.Sprintf("Failed to get the data from the database - %s", err), http.StatusInternalServerError)
		WriteResponse(w, fmt.Sprintf("Failed to get the data from the database - %s", err), http.StatusInternalServerError)
		return
	}
	resBody := struct {
		IsDefault bool              `json:"isDefault"`
		Votes     []*types.VoteSent `json:"votes"`
	}{
		IsDefault: false,
	}
	resBody.Votes = votes
	// if no votes then fetch the votes with the available votes and not one the one stored.
	if len(votes) == 0 {
		votes, err = s.Store.GetAvailableVotes(month, year)
		if err != nil {
			WriteResponse(w, fmt.Sprintf("Failed to get the data from the database - %s", err), http.StatusInternalServerError)
			return
		}
		resBody.Votes = votes
		resBody.IsDefault = true
	}
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		fmt.Println("error encoding  - ", err)
		WriteResponse(w, fmt.Sprintf("Failed to encode the data to the response body - %s", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) getNextVotes(w http.ResponseWriter) {
	votes, err := s.Store.GetNextVotes()
	if err != nil {
		WriteResponse(w, fmt.Sprintf("Failed to get the list of next votes - %s", err), http.StatusInternalServerError)
		return
	}
	resBody := struct {
		Data []*types.NextVote `json:"futurVotes"`
	}{
		Data: votes,
	}
	if err := encode(w, http.StatusOK, resBody); err != nil {
		WriteResponse(w, fmt.Sprintf("Failed to encode the data to the response body - %s", err), http.StatusInternalServerError)
		return
	}
}

// old api

// TODO: change that function, the vote should be submitted to the Vote[Month][Year] table.
func (s *Server) makeVote(w http.ResponseWriter, r *http.Request, sessionId string) {
	userId, err := s.Store.GetUserIdBySessionId(sessionId)
	if err != nil {
		WriteResponse(w, "Unable to get the events from the database", http.StatusInternalServerError)
		return
	}
	reqBody := []*types.VoteSent{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteResponse(w, fmt.Sprintf("Failed to parse the votes from the request - %s", err), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	// TODO: if there is a vote remove it then write in the database
	if s.Store.HasVote(reqBody[0].Month, reqBody[0].Year, userId) { // get the month and the year in this
		err := s.Store.RemoveVoteByUserId(reqBody[0].Month, reqBody[0].Year, userId)
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
}
