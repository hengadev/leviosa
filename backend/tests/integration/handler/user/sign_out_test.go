package userHandler_test

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"
//
// 	"github.com/GaryHY/leviosa/internal/domain/session"
// 	"github.com/GaryHY/leviosa/internal/repository/redis"
// 	"github.com/GaryHY/leviosa/internal/server/app"
// 	"github.com/GaryHY/leviosa/internal/server/handler/user"
// 	test "github.com/GaryHY/leviosa/tests/utils"
// 	"github.com/GaryHY/leviosa/tests/utils/factories"
//
// 	"github.com/GaryHY/test-assert"
// )
//
// func TestSignOut(t *testing.T) {
// 	// FIX:
// 	// - no session in database
// 	// - case user not authenticated (?)
// 	// - session already exists ?t
// 	// TEST: cases to test
// 	session := factories.NewBasicSession(nil)
// 	t.Setenv("TEST_MIGRATION_PATH", test.GetSQLiteMigrationPath())
// 	tests := []struct {
// 		name               string
// 		sessionID          string
// 		initMap            miniredis.InitMap[*sessionService.Values]
// 		wantCookie         bool
// 		expectedStatusCode int
// 		expectedCookieName string
// 	}{
// 		{sessionID: session.ID, wantCookie: false, expectedStatusCode: 500, expectedCookieName: "", initMap: nil, name: "no session in database"},
// 		{sessionID: test.GenerateRandomString(16), wantCookie: false, expectedStatusCode: 500, expectedCookieName: "", initMap: factories.NewBasicInitSession(nil), name: "session provided is not in database"},
// 		{sessionID: session.ID, wantCookie: true, expectedStatusCode: 200, expectedCookieName: sessionService.SessionName, initMap: factories.NewBasicInitSession(nil), name: "nominal case"},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			// create the cookie associated with the request
// 			cookie := &http.Cookie{
// 				Name:     sessionService.SessionName,
// 				Value:    tt.sessionID,
// 				Expires:  time.Now().Add(sessionService.SessionDuration),
// 				HttpOnly: true,
// 			}
// 			// create request and responseRecorder
// 			r, _ := http.NewRequest("POST", "/api/v1/me", nil)
// 			w := httptest.NewRecorder()
// 			r.AddCookie(cookie)
// 			// setup session service and repo
// 			sessionsvc, sessionrepo, sessionteardown := factories.SetupSession(t, r.Context(), tt.initMap)
// 			defer sessionteardown()
// 			appsvc := &app.Services{Session: sessionsvc}
// 			apprepo := &app.Repos{Session: sessionrepo}
//
// 			h := app.New(appsvc, apprepo)
// 			userhandler := userHandler.New(h)
//
// 			userhandler.Signout(w, r)
//
// 			// status code assertions
// 			assert.Equal(t, w.Code, tt.expectedStatusCode)
// 			// cookie related asserts
// 			if tt.wantCookie {
// 				resCookie := w.Result().Cookies()[0]
// 				assert.Equal(t, resCookie.Name, sessionService.SessionName)
// 				assert.Equal(t, resCookie.Expires.Before(time.Now().Add(time.Second)), true)
// 				assert.Equal(t, resCookie.Value, "")
// 			}
//
// 		})
// 	}
// }
