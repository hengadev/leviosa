package userHandler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/user"
	"github.com/GaryHY/event-reservation-app/internal/server/service"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"

	"github.com/google/uuid"
)

func TestSignIn(t *testing.T) {
	// TEST: cases to test
	// - session already exists ?
	t.Setenv("TEST_MIGRATION_PATH", "../../../sqlite/migrations/tests")
	// hash password and set an env variable with its value
	pwd := hashPassword(t, testutil.Johndoe.Password)
	t.Setenv("HASHED_PASSWORD", pwd)
	tests := []struct {
		creds              userService.Credentials
		wantCookie         bool
		expectedStatusCode int
		expectedCookieName string
		version            int64
		name               string
	}{
		{creds: userService.Credentials{Email: "", Password: testutil.Janedoe.Password}, wantCookie: false, expectedStatusCode: 400, expectedCookieName: "", version: 20240811140841, name: "invalid email"},
		{creds: userService.Credentials{Email: testutil.Johndoe.Email, Password: ""}, wantCookie: false, expectedStatusCode: 400, expectedCookieName: "", version: 20240811140841, name: "invalid password"},
		{creds: userService.Credentials{Email: testutil.Johndoe.Email, Password: testutil.Johndoe.Password}, wantCookie: false, expectedStatusCode: 500, expectedCookieName: "", version: 20240811085134, name: "credentials not in database"},
		{creds: userService.Credentials{Email: testutil.Johndoe.Email, Password: testutil.Johndoe.Password}, wantCookie: true, expectedStatusCode: 200, expectedCookieName: sessionService.SessionName, version: 20240824092110, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// encode credentials for request
			body := testutil.EncodeForBody(t, tt.creds)

			// create request and responseRecorder
			r, _ := http.NewRequest("POST", "/api/v1/signin", body)
			w := httptest.NewRecorder()

			// setup session service and repo
			usersvc, userrepo := testutil.SetupUser(t, ctx, tt.version)
			// setup session service and repo
			sessionsvc, sessionrepo, sessionteardown := testutil.SetupSession(t, ctx, nil)
			defer sessionteardown()

			appsvc := &handler.Services{User: usersvc, Session: sessionsvc}
			apprepo := &handler.Repos{User: userrepo, Session: sessionrepo}

			h := handler.New(appsvc, apprepo)
			userhandler := userHandler.New(h)

			signIn := userhandler.Signin()
			signIn.ServeHTTP(w, r)

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

func hashPassword(t testing.TB, pwd string) string {
	pwd, err := sqliteutil.HashPassword(pwd)
	if err != nil {
		t.Errorf("password hashing: %s", err)
	}
	return pwd
}
