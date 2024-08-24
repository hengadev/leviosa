package userHandler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/user"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/internal/server/service"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestGetUser(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../../../sqlite/migrations/tests")
	baseID := strconv.Itoa(testutil.Johndoe.ID)
	wrongID := strconv.Itoa(593857835)
	tests := []struct {
		userID             string
		expectedStatusCode int
		expectedUser       *userService.User
		version            int64
		name               string
	}{
		{userID: baseID, expectedStatusCode: 500, expectedUser: nil, version: 20240811085134, name: "empty database"},
		{userID: "", expectedStatusCode: 500, expectedUser: nil, version: 20240811140841, name: "no userID provided ie failed auth"},
		{userID: wrongID, expectedStatusCode: 500, expectedUser: nil, version: 20240811140841, name: "ID not in database"},
		{userID: baseID, expectedStatusCode: 302, expectedUser: testutil.Johndoe, version: 20240811140841, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/api/v1/me", nil)
			w := httptest.NewRecorder()

			// pass userID to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, mw.UserIDKey, tt.userID)
			r = r.WithContext(ctx)

			usersvc, userrepo := testutil.SetupUser(t, ctx, tt.version)

			appsvc := &handler.Services{User: usersvc}
			apprepo := &handler.Repos{User: userrepo}

			h := handler.New(appsvc, apprepo)
			userhandler := userHandler.New(h)

			getUser := userhandler.GetUser()
			getUser.ServeHTTP(w, r)

			// parse the body for the user
			var user *userService.User
			json.NewDecoder(w.Body).Decode(user)

			assert.Equal(t, w.Code, tt.expectedStatusCode)
			if tt.expectedUser != nil {
				defer testutil.RecoverCompareUser()
				testutil.CompareUser(t, testutil.BasicCompareFields, user, tt.expectedUser)
			}
		})
	}
}
