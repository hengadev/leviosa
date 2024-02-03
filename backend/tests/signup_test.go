package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPOSTSignUp(t *testing.T) {
	SignUpTests := []struct {
		name        string
		email       string
		password    string
		httpStatus  int // TODO: Check the  status I should send for each case
		succesQuery bool
	}{
		{"Valid email and valid password", "henry.gary@hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusCreated, true},
		{"User already in database", "henry.gary@hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusConflict, true},
		{"Invalid email and valid password", "henry.gary/hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusForbidden, false},
		{"Valid email and invalid password", "henry.gary/hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusForbidden, false},
	}
	server, store := makeServerAndStoreWithUsersTable()
	for _, tt := range SignUpTests {
		t.Run(tt.name, func(t *testing.T) {
			request := newPostJSONRequest(tt.email, tt.password, "/signup")
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			assertStatus(t, response.Code, tt.httpStatus)
			if tt.succesQuery {
				var hashpassword string
				var countEmail int
				store.DB.QueryRow("SELECT COUNT(email) FROM users;").Scan(&countEmail)
				assertEqualOne(t, countEmail, "email")
				store.DB.QueryRow("SELECT hashpassword FROM users;").Scan(&hashpassword)
				assertPasswordHash(t, hashpassword, tt.password)
			}
		})
	}
}
