package userHandler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/server/app"
	"github.com/GaryHY/leviosa/internal/server/handler/user"
	"github.com/GaryHY/leviosa/pkg/sqliteutil"
	"github.com/GaryHY/leviosa/pkg/testutil"
	"github.com/GaryHY/leviosa/tests/assert"

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
		creds              models.UserSignIn
		wantCookie         bool
		expectedStatusCode int
		expectedCookieName string
		version            int64
		name               string
	}{
		{creds: models.UserSignIn{Email: "", Password: testutil.Janedoe.Password}, wantCookie: false, expectedStatusCode: 400, expectedCookieName: "", version: 20240811140841, name: "invalid email"},
		{creds: models.UserSignIn{Email: testutil.Johndoe.Email, Password: ""}, wantCookie: false, expectedStatusCode: 400, expectedCookieName: "", version: 20240811140841, name: "invalid password"},
		{creds: models.UserSignIn{Email: testutil.Johndoe.Email, Password: testutil.Johndoe.Password}, wantCookie: false, expectedStatusCode: 500, expectedCookieName: "", version: 20240811085134, name: "credentials not in database"},
		{creds: models.UserSignIn{Email: testutil.Johndoe.Email, Password: testutil.Johndoe.Password}, wantCookie: true, expectedStatusCode: 200, expectedCookieName: sessionService.SessionName, version: 20240824092110, name: "nominal case"},
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

			appsvc := &app.Services{User: usersvc, Session: sessionsvc}
			apprepo := &app.Repos{User: userrepo, Session: sessionrepo}

			h := app.New(appsvc, apprepo)
			userhandler := userHandler.New(h)

			userhandler.Signin(w, r)

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
