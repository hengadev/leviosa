package types

import (
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	// CookieDuration    = time.Duration(30 * time.Minute)
	CookieDuration    = time.Duration(24 * time.Hour)
	SessionCookieName = "sessionId"
)

// NOTE: let's try DDD

// TODO: use json field tags so that I do not have so many User types

// bring the email thing for that
type OtherUser struct {
	ID       string `json:"id" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`

	CreatedAt  time.Time `json:"createdat" validate:"required"`
	LoggedInAt time.Time `json:"loggedinat" validate:"required"`
}

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("forbidden")
)

func NewInvalidInputErr(err error) error {
	return fmt.Errorf("%w %w", ErrInvalidInput, err)
}

// there is :
// - Newaccount (create the object account with the Email and Password type)
// - CreateAccount (the service that create the account in the database entre autre)

func NewOtherUser(email Email, password Password) *OtherUser {
	return &OtherUser{
		Email:    email.String(),
		Password: password.String(),
	}
}

// TODO: that part is the service and is the logic that goes in the handler in reality since we do microservice thing.
// func CreateOtherUser(email, password string) (*OtherUser, error) {
type Service struct{}

func (s *Service) CreateOtherUser(email, password string) (*OtherUser, error) {
	var input struct {
		email    Email
		password Password
	}
	// TODO: do the vaidation in here
	{
		var err error
		if input.email, err = NewEmail(email); err != nil {
			return nil, NewInvalidInputErr(err)
		}
		if input.password, err = NewPassword(password); err != nil {
			return nil, NewInvalidInputErr(err)
		}
	}
	user := NewOtherUser(input.email, input.password)
	user.Login()
	user.Create()
	// TODO: ici je peux rajouter les autres field sur le user
	return user, nil
}

// NOTE: Les deux fonctions qui suivent sont a utiliser dans le service
func (o *OtherUser) Create() {
	o.ID = uuid.NewString()
	o.CreatedAt = time.Now().UTC()
}

func (o *OtherUser) Login() {
	o.LoggedInAt = time.Now().UTC()
}

// Use to parse the information from the request from the /signup endpoint
type User struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
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
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=6"`
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

type UserSent struct {
	Id         string `json:"id"`
	Email      string `json:"email" validate:"required,email"`
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

// TODO: Where do I get that from ?
// TODO: If I try to use that with the DDD
// use all these types to make the best out of the type in golang and have the best function that you can make
// type Id string
// type Email string
// type Password string
// type LastName string
// type FirstName string
// type Gender string
// type BirthDate string
// type Telephone string
// type Address string
// type City string
// type PostalCard string
// type UserFinal struct {
// 	Id         Id
// 	Email      Email
// 	Password   Password
// 	Role       Role
// 	LastName   LastName
// 	FirstName  FirstName
// 	Gender     Gender
// 	BirthDate  BirthDate
// 	Telephone  Telephone
// 	Address    Address
// 	City       City
// 	PostalCard PostalCard
// }

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
	// TODO: Do something to  check password value validity
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
