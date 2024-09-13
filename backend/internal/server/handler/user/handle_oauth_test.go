package userHandler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/user"
	"github.com/GaryHY/event-reservation-app/internal/server/service"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
)

func TestHandleOAuth(t *testing.T) {
	// TEST: test cases
	// - nominal case google
	// - nominal case apple
	t.Setenv("TEST_MIGRATION_PATH", "../../../sqlite/migrations/tests")

	tests := []struct {
		oauthUser          userService.OAuthUser
		provider           string
		wantCookie         string
		expectedStatusCode int
		expectedCookieName string
		version            int64
		name               string
	}{
		// {},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// do something for that test brother
			// TODO:
			// - setup the user database
			// - setup the session database
			body := testutil.EncodeForBody(t, tt.oauthUser)
			endpoint := fmt.Sprintf("/api/v1/oauth/%s/user", tt.provider)
			r, _ := http.NewRequest("POST", endpoint, body)
			ctx := r.Context()
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

			handleOAuth := userhandler.HandleOAuth()
			handleOAuth.ServeHTTP(w, r)
			// TODO: do all the assertions in here
		})
	}
}
