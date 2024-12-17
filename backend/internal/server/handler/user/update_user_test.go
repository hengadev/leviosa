package userHandler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/server/app"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/user"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestUpdateUser(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../../../sqlite/migrations/tests")
	userToModify := *testutil.Johndoe
	userToModify.Telephone = "0234567890"
	var zerotime time.Time
	tests := []struct {
		user               userService.User
		expectedStatusCode int
		version            int64
		name               string
	}{
		{user: *testutil.Johndoe, expectedStatusCode: 500, version: 20240811085134, name: "no user in database"},
		{user: *testutil.Janedoe, expectedStatusCode: 500, version: 20240811140841, name: "user not in database"},
		{user: *testutil.Johndoe, expectedStatusCode: 200, version: 20240811140841, name: "user with no modification to do"},
		{user: userToModify, expectedStatusCode: 200, version: 20240811140841, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userID := tt.user.ID
			// remove forbidden field
			tt.user.ID = 0
			tt.user.Email = ""
			tt.user.Password = ""
			tt.user.CreatedAt = zerotime
			tt.user.LoggedInAt = zerotime
			tt.user.Role = ""

			// encode credentials for request
			body := testutil.EncodeForBody(t, tt.user)

			// create request and responseRecorder
			r, _ := http.NewRequest("PUT", "/api/v1/me", body)
			w := httptest.NewRecorder()

			ctx := r.Context()
			ctx = context.WithValue(ctx, mw.UserIDKey, userID)
			r = r.WithContext(ctx)

			// setup session service and repo
			usersvc, userrepo := testutil.SetupUser(t, ctx, tt.version)

			appsvc := &app.Services{User: usersvc}
			apprepo := &app.Repos{User: userrepo}

			h := app.New(appsvc, apprepo)
			userhandler := userHandler.New(h)

			updateUser := userhandler.UpdateUser()
			updateUser.ServeHTTP(w, r)

			// status code assertions
			assert.Equal(t, w.Code, tt.expectedStatusCode)
		})
	}
}
