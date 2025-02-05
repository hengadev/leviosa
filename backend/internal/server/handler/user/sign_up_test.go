package userHandler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/repository/redis"
	"github.com/GaryHY/leviosa/internal/server/app"
	"github.com/GaryHY/leviosa/internal/server/handler/user"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	"github.com/GaryHY/test-assert"
	"github.com/google/uuid"
)

func TestCreateAccount(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../../../sqlite/migrations/tests")
	tests := []struct {
		user               *models.User
		wantCookie         bool
		expectedStatusCode int
		expectedCookieName string
		initMap            miniredis.InitMap[*sessionService.Values]
		version            int64
		name               string
	}{
		{user: nil, wantCookie: false, expectedStatusCode: 400, expectedCookieName: "", initMap: factories.InitSession, version: 20240811085134, name: "no user in database"},
		{user: factories.Johndoe, wantCookie: false, expectedStatusCode: 500, expectedCookieName: "", initMap: factories.InitSession, version: 20240824092110, name: "user already exists"},
		{user: factories.Johndoe, wantCookie: true, expectedStatusCode: 201, expectedCookieName: sessionService.SessionName, initMap: factories.InitSession, version: 20240811085134, name: "nominal case with user creation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// encode credentials for request
			body := factories.EncodeForBody(t, tt.user)

			// create request and responseRecorder
			r, _ := http.NewRequest("POST", "/api/v1/signup", body)
			_ = r
			w := httptest.NewRecorder()

			// setup session service and repo
			usersvc, userrepo := factories.SetupUser(t, ctx, tt.version)
			// setup session service and repo
			sessionsvc, sessionrepo, sessionteardown := factories.SetupSession(t, ctx, tt.initMap)
			defer sessionteardown()

			appsvc := &app.Services{User: usersvc, Session: sessionsvc}
			apprepo := &app.Repos{User: userrepo, Session: sessionrepo}

			h := app.New(appsvc, apprepo)
			userhandler := userHandler.New(h)
			_ = userhandler

			// signUp := userhandler.CreateAccount()
			// signUp := userhandler.CreateUser()
			// signUp.ServeHTTP(w, r)

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
