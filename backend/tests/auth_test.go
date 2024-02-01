package tests

import (
	"github.com/GaryHY/event-reservation-app/internal/api"
	"github.com/GaryHY/event-reservation-app/internal/database"
	// TODO: Try to test using UUID
	// "github.com/google/uuid"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TODO: Test the validate email/password functions

func TestPOSTSignUp(t *testing.T) {
	// NOTE: liste de tests que je veux faire
	// valid email and valid password
	// invalid email
	// invalid passowrd
	// email already used in database

	AuthTests := []struct {
		name        string
		email       string
		password    string
		httpStatus  int // TODO: Check the  status I should send for each case
		succesQuery bool
	}{
		{"Valid email and valid password", "henry.gary@hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusCreated, true},
		{"User already in database", "henry.gary@hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusConflict, true},
		// NOTE: Add this later to complexify the tests sample and complete it with the remaininig field
		// {"Invalid email and valid password", "henry.gary/hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusOK, false},
		// {"Valid email and Invalid password"},
	}

	createEventTable := "CREATE TABLE IF NOT EXISTS users (email TEXT NOT NULL PRIMARY KEY, hashpassword TEXT NOT NULL);"

	store, err := sqlite.NewStore("")
	if err != nil {
		log.Fatal("Something went wrong when creating the database")
	}
	store.Init(createEventTable)
	server := api.NewServer(store)

	for _, tt := range AuthTests {
		t.Run(tt.name, func(t *testing.T) {

			jsonData := []byte(fmt.Sprintf(`{"Email": "%s", "Password": "%s"}`, tt.email, tt.password))
			request, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonData))
			request.Header.Set("Content-Type", "application/json; charset=UTF-8")
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)
			assertStatus(t, response.Code, tt.httpStatus)

			if tt.succesQuery {
				var email string
				var hashpassword string
				var countEmail int

				userQuery := "SELECT email, hashpassword FROM users;"
				store.DB.QueryRow(userQuery).Scan(&email, &hashpassword)
				// 1. tester que le hashpassword et tt.password corrresponde avec bcrypt
				// TODO: Put the password using the bcrypt thing
				assertPasswordHash(t, hashpassword, tt.password)

				// 2. count the number of occurence of the email in the database
				countQuery := "SELECT COUNT(email) FROM users;"
				store.DB.QueryRow(countQuery).Scan(&countEmail)
				assertEmailCount(t, countEmail)
			}
		})
	}

}
