package types

import (
	"net/mail"
	"reflect"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionDuration   = time.Duration(30 * time.Minute)
	SessionCookieName = "session_token"
)

// Use to parse the information from the request from the /signup endpoint
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Use to parse the information from the request from the /signin endpoint
type UserForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// Role       string `json:"role"`
	LastName   string `json:"lastname"`
	FirstName  string `json:"firstname"`
	Gender     string `json:"gender"`
	BirthDate  string `json:"birthdate"`
	Telephone  string `json:"telephone"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCard string `json:"postalcard"`
}

// Use to store user information in the database
func NewUserStored(user *UserForm) *UserStored {
	return &UserStored{
		Id:         uuid.NewString(),
		Email:      user.Email,
		Password:   user.Password,
		Role:       "basic",
		LastName:   user.LastName,
		FirstName:  user.FirstName,
		Gender:     user.Gender,
		BirthDate:  user.BirthDate,
		Telephone:  user.Telephone,
		Address:    user.Address,
		City:       user.City,
		PostalCard: user.PostalCard,
	}
}

type UserStored struct {
	Id         string `json:"id"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	LastName   string `json:"lastname"`
	FirstName  string `json:"firstname"`
	Gender     string `json:"gender"`
	BirthDate  string `json:"birthdate"`
	Telephone  string `json:"telephone"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCard string `json:"postalcard"`
}

func (u *User) ValidateEmail() bool {
	_, err := mail.ParseAddress(u.Email)
	return err == nil
}

// TODO: Add the logic to verify if the password are good enough
func (u *User) ValidatePassword() bool {
	return true
}

// TODO: DO THE GENERICS OFF THAT !!!
func (u *User) IsNull() bool {
	if u.Email == "" && u.Password == "" {
		return true
	}
	return false
}

// TODO: use both of the following function or use a simple one that just take a reference to a string to make the calculation since I just want to read the values
func ValidatePasswordGeneric[T any](data *T) bool {
	v := reflect.ValueOf(*data)
	if v.Kind() != reflect.Struct {
		panic("Expected a struct")
	}
	password := v.FieldByNameFunc(func(s string) bool {
		return s == "Password"
	})
	// TODO: do something to  check password value validity
	_ = password
	return true
}

func ValidateEmailGeneric[T any](data *T) bool {
	v := reflect.ValueOf(*data)
	if v.Kind() != reflect.Struct {
		panic("Expected a struct")
	}
	email := v.FieldByNameFunc(func(s string) bool {
		return s == "Email"
	})
	_, err := mail.ParseAddress(email.String())
	return err == nil
}

func IsNullGeneric[T any](data *T) bool {
	v := reflect.ValueOf(*data)
	if v.Kind() != reflect.Struct {
		panic("Expected a struct")
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.String() != "" {
			return false
		}
	}
	return true
}

func (u *UserStored) ValidateEmail() bool {
	_, err := mail.ParseAddress(u.Email)
	return err == nil
}

// TODO: Add the logic to verify if the password are good enough
func (u *UserStored) ValidatePassword() bool {
	return true
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
