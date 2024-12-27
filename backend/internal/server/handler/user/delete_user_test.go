package userHandler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/internal/repository/redis"
	"github.com/GaryHY/event-reservation-app/internal/server/app"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/user"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestDeleteUser(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../../../sqlite/migrations/tests")
	baseID := testutil.Johndoe.ID
	wrongID := strconv.Itoa(593857835)
	tests := []struct {
		userID             string
		expectedStatusCode int
		initMap            miniredis.InitMap[*sessionService.Values]
		version            int64
		name               string
	}{
		{userID: baseID, expectedStatusCode: 500, initMap: testutil.InitSession, version: 20240811085134, name: "empty database"},
		{userID: wrongID, expectedStatusCode: 500, initMap: testutil.InitSession, version: 20240811140841, name: "user not in database"},
		{userID: baseID, expectedStatusCode: 500, initMap: nil, version: 20240811140841, name: "session not found"},
		{userID: baseID, expectedStatusCode: 200, initMap: testutil.InitSession, version: 20240811140841, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, _ := http.NewRequest("GET", "/api/v1/me", nil)
			w := httptest.NewRecorder()

			cookie := &http.Cookie{
				Name:     sessionService.SessionName,
				Value:    testutil.SessionID,
				Expires:  time.Now().Add(sessionService.SessionDuration),
				HttpOnly: true,
			}
			r.AddCookie(cookie)

			// pass userID to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, contextutil.UserIDKey, tt.userID)
			r = r.WithContext(ctx)

			usersvc, userrepo := testutil.SetupUser(t, ctx, tt.version)

			sessionsvc, sessionrepo, sessionteardown := testutil.SetupSession(t, ctx, tt.initMap)
			defer sessionteardown()

			appsvc := &app.Services{User: usersvc, Session: sessionsvc}
			apprepo := &app.Repos{User: userrepo, Session: sessionrepo}

			h := app.New(appsvc, apprepo)
			userhandler := userHandler.New(h)

			deleteUser := userhandler.DeleteUser()
			deleteUser.ServeHTTP(w, r)

			// parse the body for the user
			var user *models.User
			json.NewDecoder(w.Body).Decode(user)

			assert.Equal(t, w.Code, tt.expectedStatusCode)
		})
	}
}
