package userHandler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/redis"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/user"
	"github.com/GaryHY/event-reservation-app/internal/server/service"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"

	"github.com/google/uuid"
)

func TestCreateAccount(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../../../sqlite/migrations/tests")
	tests := []struct {
		user               *userService.User
		wantCookie         bool
		expectedStatusCode int
		expectedCookieName string
		initMap            miniredis.InitMap[*sessionService.Values]
		version            int64
		name               string
	}{
		{user: nil, wantCookie: false, expectedStatusCode: 400, expectedCookieName: "", initMap: testutil.InitSession, version: 20240811085134, name: "no user in database"},
		{user: testutil.Johndoe, wantCookie: false, expectedStatusCode: 500, expectedCookieName: "", initMap: testutil.InitSession, version: 20240824092110, name: "user already exists"},
		{user: testutil.Johndoe, wantCookie: true, expectedStatusCode: 201, expectedCookieName: sessionService.SessionName, initMap: testutil.InitSession, version: 20240811085134, name: "nominal case with user creation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// encode credentials for request
			body := testutil.EncodeForBody(t, tt.user)

			// create request and responseRecorder
			r, _ := http.NewRequest("POST", "/api/v1/signup", body)
			w := httptest.NewRecorder()

			// setup session service and repo
			usersvc, userrepo := testutil.SetupUser(t, ctx, tt.version)
			// setup session service and repo
			sessionsvc, sessionrepo, sessionteardown := testutil.SetupSession(t, ctx, tt.initMap)
			defer sessionteardown()

			appsvc := &handler.Services{User: usersvc, Session: sessionsvc}
			apprepo := &handler.Repos{User: userrepo, Session: sessionrepo}

			h := handler.New(appsvc, apprepo)
			userhandler := userHandler.New(h)

			signUp := userhandler.CreateAccount()
			signUp.ServeHTTP(w, r)

			// status code assertions
			assert.Equal(t, w.Code, tt.expectedStatusCode)

			// cookie related asserts
			if tt.wantCookie {
				resCookie := w.Result().Cookies()[0]
				assert.Equal(t, resCookie.Name, sessionService.SessionName)
				assert.Equal(t, resCookie.Expires.After(time.Now()), true)
				if _, err := uuid.Parse(resCookie.Value); err != nil {
					t.Errorf("cookie value is not UUID type")
				}
			}

		})
	}
}
