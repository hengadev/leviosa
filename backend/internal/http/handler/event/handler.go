package eventhandler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/http/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/http/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// handler
func FindEventByID(eventRepo event.Reader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		eventID := r.PathValue("id")
		event, err := eventRepo.GetEventByID(ctx, eventID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the user ID", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err := serverutil.Encode(w, http.StatusOK, event); err != nil {
			slog.ErrorContext(ctx, "failed to encode the user ID", "error", err)
			http.Error(w, fmt.Sprintf("Unable to get event with the id of %q", eventID), http.StatusInternalServerError)
			return
		}
	})
}

func FindEventsForUser(eventRepo event.Reader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		userID := ctx.Value(mw.SessionIDKey).(string)
		if userID == "" { // TODO: find better thing to do to check if the userID stored is valid.
			// some error with the auth that did not sent the userID in the context.
		}
		resBody, err := eventRepo.GetEventForUser(ctx, userID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the events for the user", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err := serverutil.Encode(w, http.StatusOK, resBody); err != nil {
			slog.ErrorContext(ctx, "failed to send the user", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func CreateEvent(svc *event.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get the event from the body
		event, err := serverutil.Decode[event.Event](r)
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode the event", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// use the service to make that event
		eventID, err := svc.CreateEvent(ctx, &event)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create the event", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// send the event id to the user
		if err := serverutil.Encode(w, http.StatusOK, eventID); err != nil {
			slog.ErrorContext(ctx, "failed to send the event", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func DeleteEvent(svc *event.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get the event id from the url path
		eventID := r.PathValue("id")
		// use the service to delete that event
		resEventID, err := svc.RemoveEvent(ctx, eventID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to delete the event", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// send the event id to the user to specify that the event is deleted properly.
		if err = serverutil.Encode(w, http.StatusInternalServerError, resEventID); err != nil {
			slog.ErrorContext(ctx, "failed to send the event ID", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func ModifyEvent(svc *event.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		event, err := serverutil.Decode[event.Event](r)
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode the event", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		eventID, err := svc.ModifyEvent(ctx, &event)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update the event", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err = serverutil.Encode(w, http.StatusInternalServerError, eventID); err != nil {
			slog.ErrorContext(ctx, "failed to send the event ID", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
