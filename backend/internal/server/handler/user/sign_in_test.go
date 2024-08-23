package userHandler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/user"
	"github.com/GaryHY/event-reservation-app/internal/server/service"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestSignIn(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../../../sqlite/migrations/tests")
	// baseID := strconv.Itoa(testutil.Johndoe.ID)
	// wrongID := strconv.Itoa(593857835)
	// TEST: the test cases that I need
	// - invalid email
	// - invalid password
	// - credentials that are not valid
	// TODO: what do I need ?
	// - the credential in the body of the request : email + password
	// - create the session service thing with testutil
	// NOTE: the assert that I need
	// - check if err
	// - check userID is right
	// - check role is right
	// - check the session ID is valid or something
	// - check the cookie
	tests := []struct {
		userID             string
		expectedStatusCode int
		expectedUserID     int
		expectedRole       userService.Role
		wantSessionID      bool
		userVersion        int64
		sessionVersion     int64
		name               string
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			r, _ := http.NewRequest("GET", "/api/v1/me", nil)
			w := httptest.NewRecorder()

			usersvc, userrepo := testutil.SetupUser(t, ctx, tt.userVersion)

			sessionsvc, sessionrepo, sessionteardown := testutil.SetupSession(t, ctx, tt.sessionVersion)
			defer sessionteardown()

			// NOTE: the final service thing
			appsvc := &handler.Services{
				User:    usersvc,
				Session: sessionsvc,
			}
			apprepo := &handler.Repos{
				User:    userrepo,
				Session: sessionrepo,
			}

			h := handler.New(appsvc, apprepo)
			userhandler := userHandler.New(h)

			signIn := userhandler.Signin()
			signIn.ServeHTTP(w, r)

			// TODO: find a way to get the cookie from the response
			// w.Result().Cookies()

			assert.Equal(t, w.Code, tt.expectedStatusCode)
		})
	}
}
