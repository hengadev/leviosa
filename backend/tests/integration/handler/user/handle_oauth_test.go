package userHandler_test

// import (
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
//
// 	"github.com/GaryHY/leviosa/internal/domain/user/models"
// 	"github.com/GaryHY/leviosa/internal/server/app"
// 	userHandler "github.com/GaryHY/leviosa/internal/server/handler/user"
// 	test "github.com/GaryHY/leviosa/tests/utils"
// 	"github.com/GaryHY/leviosa/tests/utils/factories"
// )
//
// func TestHandleOAuth(t *testing.T) {
// 	// TEST: test cases
// 	// - nominal case google
// 	// - nominal case apple
// 	t.Setenv("TEST_MIGRATION_PATH", test.GetSQLiteMigrationPath())
//
// 	tests := []struct {
// 		oauthUser          models.OAuthUser
// 		provider           string
// 		wantCookie         string
// 		expectedStatusCode int
// 		expectedCookieName string
// 		version            int64
// 		name               string
// 	}{
// 		// {},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			// do something for that test brother
// 			// TODO:
// 			// - setup the user database
// 			// - setup the session database
// 			body := factories.EncodeForBody(t, tt.oauthUser)
// 			endpoint := fmt.Sprintf("/api/v1/oauth/%s/user", tt.provider)
// 			r, _ := http.NewRequest("POST", endpoint, body)
// 			ctx := r.Context()
// 			w := httptest.NewRecorder()
//
// 			// setup session service and repo
// 			usersvc, userrepo := factories.SetupUser(t, ctx, tt.version)
// 			// setup session service and repo
// 			sessionsvc, sessionrepo, sessionteardown := factories.SetupSession(t, ctx, nil)
// 			defer sessionteardown()
//
// 			appsvc := &app.Services{User: usersvc, Session: sessionsvc}
// 			apprepo := &app.Repos{User: userrepo, Session: sessionrepo}
//
// 			h := app.New(appsvc, apprepo)
// 			userhandler := userHandler.New(h)
//
// 			userhandler.HandleOAuth(w, r)
// 			// TODO: do all the assertions in here
// 		})
// 	}
// }
