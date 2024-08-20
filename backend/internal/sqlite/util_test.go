package sqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
)

// general
type RepoConstructor[T sqlite.Repository] func(context.Context, *sql.DB) T

func setupRepo[T sqlite.Repository](ctx context.Context, version int64, constructor RepoConstructor[T]) (T, error) {
	var repo T
	db, err := testdb.NewDatabase(ctx)
	if err != nil {
		return repo, fmt.Errorf("database connection: %s", err)
	}
	repo = constructor(ctx, db)
	if err := testdb.Setup(ctx, repo.GetDB(), version); err != nil {
		return repo, fmt.Errorf("migration to the database: %s", err)
	}
	return repo, nil
}

func createModifiedObject[T any](baseObject T, changeMap map[string]any) (*T, error) {
	newObjectPtr := reflect.New(reflect.TypeOf(baseObject)).Interface().(*T)
	v := reflect.ValueOf(newObjectPtr).Elem()
	t := reflect.TypeOf(baseObject)
	vf := reflect.VisibleFields(t)

	*newObjectPtr = baseObject

	for _, field := range vf {
		if value, ok := changeMap[field.Name]; ok {
			fieldValue := v.FieldByName(field.Name)
			switch fieldValue.Kind() {
			// case reflect.Int, reflect.Int64:
			case reflect.Int:
				if val, ok := value.(int); ok {
					fieldValue.SetInt(int64(val))
				} else {
					return nil, fmt.Errorf("type mismatch for field %s: expected int, got %T", field.Name, value)
				}
			case reflect.String:
				if val, ok := value.(string); ok {
					fieldValue.SetString(val)
				} else {
					return nil, fmt.Errorf("type mismatch for field %s: expected string, got %T", field.Name, value)
				}
				// ... Add additional cases for other types I want to support
			default:
				return nil, fmt.Errorf("unsupported field type: %s", field.Name)
			}
		}
	}
	return newObjectPtr, nil
}

func createWithZeroFieldModifiedObject[T any](baseObject T, changeMap map[string]any) (*T, error) {
	newObjectPtr := reflect.New(reflect.TypeOf(baseObject)).Interface().(*T)
	v := reflect.ValueOf(newObjectPtr).Elem()
	t := reflect.TypeOf(baseObject)
	vf := reflect.VisibleFields(t)

	for _, field := range vf {
		if value, ok := changeMap[field.Name]; ok {
			fieldValue := v.FieldByName(field.Name)
			switch fieldValue.Kind() {
			// case reflect.Int, reflect.Int64:
			case reflect.Int:
				if val, ok := value.(int); ok {
					fieldValue.SetInt(int64(val))
				} else {
					return nil, fmt.Errorf("type mismatch for field %s: expected int, got %T", field.Name, value)
				}
			case reflect.String:
				if val, ok := value.(string); ok {
					fieldValue.SetString(val)
				} else {
					return nil, fmt.Errorf("type mismatch for field %s: expected string, got %T", field.Name, value)
				}
				// ... Add additional cases for other types I want to support
			default:
				return nil, fmt.Errorf("unsupported field type: %s", field.Name)
			}
		}
	}
	return newObjectPtr, nil
}

// users for tests
var (
	// ce sont des triplets
	johndoe = &user.User{
		ID:         1,
		Email:      "john.doe@gmail.com",
		Password:   "$a9rfNhA$N$A78#m",
		CreatedAt:  time.Now().Add(-time.Hour * 4),
		LoggedInAt: time.Now().Add(-time.Hour * 4),
		Role:       user.BASIC.String(),
		BirthDate:  "1998-07-12",
		LastName:   "DOE",
		FirstName:  "John",
		Gender:     "M",
		Telephone:  "0123456789",
		Address:    "Impasse Inconnue",
		City:       "Paris",
		PostalCard: 12345,
	}
	janedoe = &user.User{
		ID:         2,
		Email:      "jane.doe@gmail.com",
		Password:   "w4w3f09QF&h)#fwe",
		CreatedAt:  time.Now().Add(-time.Hour * 4),
		LoggedInAt: time.Now().Add(-time.Hour * 4),
		Role:       user.BASIC.String(),
		BirthDate:  "1998-07-12",
		LastName:   "DOE",
		FirstName:  "Jane",
		Gender:     "F",
		Telephone:  "0123456780",
		Address:    "Impasse Inconnue",
		City:       "Paris",
		PostalCard: 12345,
	}
	jeandoe = &user.User{
		ID:         1,
		Email:      "jean.doe@gmail.com",
		Password:   "wf0fT^9f2$$_aewa",
		CreatedAt:  time.Now().Add(-time.Hour * 4),
		LoggedInAt: time.Now().Add(-time.Hour * 4),
		Role:       user.BASIC.String(),
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

func compareUser(t testing.TB, fields []string, userDB *user.User, realUser *user.User) {
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

// events for tests
var (
	beginAt, _ = sqlite.ExportedParseBeginAt("08:00:00", 12, 7, 1998)

	baseEvent = &event.Event{
		ID:              "ea1d74e2-1612-47ec-aee9-c6a46b65640f",
		Location:        "Impasse Inconnue",
		PlaceCount:      16,
		FreePlace:       14,
		BeginAt:         beginAt,
		SessionDuration: time.Minute * 30,
		Day:             12,
		Month:           7,
		Year:            1998,
	}

	baseEventWithPriceID = &event.Event{
		ID:              "ea1d74e2-1612-47ec-aee9-c6a46b65640f",
		Location:        "Impasse Inconnue",
		PlaceCount:      16,
		FreePlace:       14,
		BeginAt:         beginAt,
		SessionDuration: time.Minute * 30,
		PriceID:         "4fe0vuw3ef0223",
		Day:             12,
		Month:           7,
		Year:            1998,
	}
)
