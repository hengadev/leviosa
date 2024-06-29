package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// handler
func (s *Server) eventHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	sessionId := parseAuthorizationHeader(r)
	switch r.Method {
	case http.MethodOptions:
		enableMethods(&w, http.MethodGet)
	case http.MethodGet:
		// TODO: use the r.PathValue method instead and split the handler so that I cna use only one function
		// id := r.PathValue("id")
		id := r.URL.Query().Get("id") // the id I get is the eventId for specific information around it.
		if id == "" {
			s.showAllEvents(w, sessionId)
		} else {
			s.showEvent(w, id)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (s *Server) getEventById(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")
	event := s.Store.GetEventByID(eventID)
	if err := encode(w, http.StatusOK, event); err != nil {
		http.Error(w, fmt.Sprintf("Unable to get event with the id of %q", eventID), http.StatusInternalServerError)
	}
}

const requestDuration = 500

type userch struct {
	userID string
	err    error
}

func otherGetUserIdBySessionId(
	ctx context.Context,
	sessionId string,
	wg *sync.WaitGroup,
) chan<- userch {
	var ch = make(chan userch)
	defer wg.Done()
	userID := "302933498234"
	err := errors.New("some error")
	go func() {
		defer close(ch)
		ch <- userch{
			userID: userID,
			err:    err,
		}
	}()
	return ch
}

// the function that I try to code that uses go routines to speed up the process.
func (s *Server) otherGetAllEvents(w http.ResponseWriter, r *http.Request) {
	sessionId := parseAuthorizationHeader(r)
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(requestDuration)*time.Millisecond)
	defer cancel()
	var userch = make(chan struct {
		userID string
		err    error
	})
	var eventch = make(chan struct {
		event types.EventBody
		err   error
	})
	var wg sync.WaitGroup
	wg.Add(2)
	// that thing should return a channel
	otherGetUserIdBySessionId(ctx, sessionId, userch, &wg)
	// TODO: do the same thing for the event thing.
	for i := 0; i < 2; i++ {
		select {
		case <-userch:
		case <-eventch:
		case <-ctx.Done():
			http.Error(w, "The request takes too long to finish", http.StatusInternalServerError)
		}
	}
}

func (s *Server) getAllEvents(w http.ResponseWriter, r *http.Request) {
	sessionId := parseAuthorizationHeader(r)
	userID, err := s.Store.GetUserIdBySessionId(sessionId)
	if err != nil {
		WriteResponse(w, "Unable to get the events from the database", http.StatusInternalServerError)
		return
	}
	resBody, err := s.Store.GetEventForUser(userID)
	if err != nil {
		WriteResponse(w, "Unable to get the events from the database", http.StatusInternalServerError)
		return
	}
	if err := encode(w, http.StatusOK, resBody); err != nil {
		http.Error(w, "Unable to serialize data for the events", http.StatusInternalServerError)
	}
}

// NOTE: old api with the general handler
// Return an event specific by ID.
func (s *Server) showEvent(w http.ResponseWriter, eventId string) {
	event := s.Store.GetEventByID(eventId)
	if err := json.NewEncoder(w).Encode(&event); err != nil {
		WriteResponse(w, fmt.Sprintf("Unable to encode the data for the events of the event identified by the id %q - %s", eventId, err), http.StatusInternalServerError)
		return
	}
}

// Return events for specific user.
func (s *Server) showAllEvents(w http.ResponseWriter, sessionId string) {
	userId, err := s.Store.GetUserIdBySessionId(sessionId)
	if err != nil {
		WriteResponse(w, "Unable to get the events from the database", http.StatusInternalServerError)
		return
	}
	resBody, err := s.Store.GetEventForUser(userId)
	if err != nil {
		WriteResponse(w, "Unable to get the events from the database", http.StatusInternalServerError)
		return
	}
	// the encode send some problem for some reason
	if err := encode(w, http.StatusOK, resBody); err != nil {
		http.Error(w, "Unable to serialize data for the events", http.StatusInternalServerError)
	}
}
