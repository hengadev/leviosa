package userRepository_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

var (
	// ce sont des triplets
	johndoe = &userService.User{
		ID:         1,
		Email:      "john.doe@gmail.com",
		Password:   "$a9rfNhA$N$A78#m",
		CreatedAt:  time.Now().Add(-time.Hour * 4),
		LoggedInAt: time.Now().Add(-time.Hour * 4),
		Role:       userService.BASIC.String(),
		BirthDate:  "1998-07-12",
		LastName:   "DOE",
		FirstName:  "John",
		Gender:     "M",
		Telephone:  "0123456789",
		Address:    "Impasse Inconnue",
		City:       "Paris",
		PostalCard: 12345,
	}
	janedoe = &userService.User{
		ID:         2,
		Email:      "jane.doe@gmail.com",
		Password:   "w4w3f09QF&h)#fwe",
		CreatedAt:  time.Now().Add(-time.Hour * 4),
		LoggedInAt: time.Now().Add(-time.Hour * 4),
		Role:       userService.BASIC.String(),
		BirthDate:  "1998-07-12",
		LastName:   "DOE",
		FirstName:  "Jane",
		Gender:     "F",
		Telephone:  "0123456780",
		Address:    "Impasse Inconnue",
		City:       "Paris",
		PostalCard: 12345,
	}
	jeandoe = &userService.User{
		ID:         1,
		Email:      "jean.doe@gmail.com",
		Password:   "wf0fT^9f2$$_aewa",
		CreatedAt:  time.Now().Add(-time.Hour * 4),
		LoggedInAt: time.Now().Add(-time.Hour * 4),
		Role:       userService.BASIC.String(),
		BirthDate:  "1998-07-12",
		LastName:   "DOE",
		FirstName:  "Jean",
		Gender:     "M",
		Telephone:  "0123456781",
		Address:    "Impasse Inconnue",
		City:       "Paris",
		PostalCard: 12345,
	}
)

func compareUser(t testing.TB, fields []string, userDB *userService.User, realUser *userService.User) {
	t.Helper()
	userDBValue := reflect.ValueOf(*userDB)
	userRealValue := reflect.ValueOf(*realUser)
	for _, field := range fields {
		dbValue := userDBValue.FieldByName(field).Interface()
		realValue := userRealValue.FieldByName(field).Interface()
		if dbValue != realValue {
			t.Errorf("got %v, want %v", dbValue, realValue)
		}
	}
}
