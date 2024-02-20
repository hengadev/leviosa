package tests

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TODO: Test the case when someone is making an account as an admin and by specifying the role of the user

func TestPOSTSignUp(t *testing.T) {
	SignUpTests := []struct {
		name        string
		email       string
		password    string
		httpStatus  int
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
	t.Run("Incorred method used", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/signup", nil)
		if err != nil {
			log.Fatal("Fail to create new GET request")
		}
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusMethodNotAllowed)
	})
}
