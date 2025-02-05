package userHandler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/repository/redis"
	"github.com/GaryHY/leviosa/internal/server/app"
	"github.com/GaryHY/leviosa/internal/server/handler/user"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/tests/utils"
	"github.com/GaryHY/leviosa/tests/utils/factories"
	"github.com/GaryHY/test-assert"
)

func TestDeleteUser(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../../../sqlite/migrations/tests")
	user := factories.NewBasicUser(nil)
	baseID := user.ID
	// wrongID := strconv.Itoa(593857835)
	wrongID := test.GenerateRandomString(36)
	tests := []struct {
		userID             string
		expectedStatusCode int
		initMap            miniredis.InitMap[*sessionService.Values]
		version            int64
		name               string
	}{
		{userID: baseID, expectedStatusCode: 500, initMap: factories.InitSession, version: 20240811085134, name: "empty database"},
		{userID: wrongID, expectedStatusCode: 500, initMap: factories.InitSession, version: 20240811140841, name: "user not in database"},
		{userID: baseID, expectedStatusCode: 500, initMap: nil, version: 20240811140841, name: "session not found"},
		{userID: baseID, expectedStatusCode: 200, initMap: factories.InitSession, version: 20240811140841, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, _ := http.NewRequest("GET", "/api/v1/me", nil)
			w := httptest.NewRecorder()

			cookie := &http.Cookie{
				Name:     sessionService.SessionName,
				Value:    factories.SessionID,
				Expires:  time.Now().Add(sessionService.SessionDuration),
				HttpOnly: true,
			}
			r.AddCookie(cookie)

			// pass userID to context
			ctx := context.WithValue(context.Background(), contextutil.UserIDKey, tt.userID)
			r = r.WithContext(ctx)

			usersvc, userrepo := factories.SetupUser(t, ctx, tt.version)

			sessionsvc, sessionrepo, sessionteardown := factories.SetupSession(t, ctx, tt.initMap)
			defer sessionteardown()

			appsvc := &app.Services{User: usersvc, Session: sessionsvc}
			apprepo := &app.Repos{User: userrepo, Session: sessionrepo}

			h := app.New(appsvc, apprepo)
			userhandler := userHandler.New(h)

			userhandler.DeleteUser(w, r)

			// parse the body for the user
			var user *models.User
			json.NewDecoder(w.Body).Decode(user)

			assert.Equal(t, w.Code, tt.expectedStatusCode)
		})
	}
}
