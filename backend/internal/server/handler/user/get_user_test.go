package userHandler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/server/app"
	"github.com/GaryHY/leviosa/internal/server/handler/user"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	"github.com/GaryHY/test-assert"
)

func TestGetUser(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../../../sqlite/migrations/test")
	baseID := factories.Johndoe.ID
	wrongID := strconv.Itoa(593857835)
	fields := []string{"ID", "Email", "Role", "BirthDate", "LastName", "FirstName", "Gender", "Telephone", "Address", "City", "PostalCard"}
	tests := []struct {
		userID             string
		expectedStatusCode int
		expectedUser       *models.User
		version            int64
		name               string
	}{
		{userID: baseID, expectedStatusCode: 500, expectedUser: nil, version: 20240811085134, name: "empty database"},
		{userID: "", expectedStatusCode: 500, expectedUser: nil, version: 20240811140841, name: "no userID provided ie failed auth"},
		{userID: wrongID, expectedStatusCode: 500, expectedUser: nil, version: 20240811140841, name: "ID not in database"},
		{userID: baseID, expectedStatusCode: 302, expectedUser: factories.Johndoe, version: 20240811140841, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, _ := http.NewRequest("GET", "/api/v1/me", nil)
			w := httptest.NewRecorder()

			// pass userID to context
			ctx := context.WithValue(r.Context(), contextutil.UserIDKey, tt.userID)
			r = r.WithContext(ctx)

			usersvc, userrepo := factories.SetupUser(t, ctx, tt.version)

			appsvc := &app.Services{User: usersvc}
			apprepo := &app.Repos{User: userrepo}

			h := app.New(appsvc, apprepo)
			userhandler := userHandler.New(h)

			userhandler.GetUser(w, r)

			// parse the body for the user
			var user *models.User
			json.NewDecoder(w.Body).Decode(user)

			assert.Equal(t, w.Code, tt.expectedStatusCode)
			if tt.expectedUser != nil {
				assert.FieldsEqual(t, user, tt.expectedUser, fields)
			}
		})
	}
}
