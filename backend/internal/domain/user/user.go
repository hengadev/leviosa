package user

import (
	"time"

	// "where is the role type ?"

	"github.com/google/uuid"
)

// TODO: do better with the address
type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email" validate:"required,email"`
	Password   string    `json:"-" validate:"required,min=6"`
	CreatedAt  time.Time `json:"createdat"`
	LoggedInAt time.Time `json:"loggedinat"`
	Role       string    `json:"role"`
	BirthDate  time.Time `json:"birthdate"`
	LastName   string    `json:"lastname"`
	FirstName  string    `json:"firstname"`
	Gender     string    `json:"gender"`
	Telephone  string    `json:"telephone"`
	Address    string    `json:"address"`
	City       string    `json:"city"`
	PostalCard string    `json:"postalcard"`
}

// what is the option pattern in golang ? Can I use it in here since some element are not going to be used to send by the user
func NewUser(
	email Email,
	password Password,
	birthdate time.Time,
	lastname,
	firstname,
	gender,
	telephone,
	address,
	city,
	postalcard string,
) *User {
	return &User{
		Email:      email.String(),
		Password:   password.String(),
		Role:       BASIC.String(),
		BirthDate:  birthdate,
		LastName:   lastname,
		FirstName:  firstname,
		Gender:     gender,
		Telephone:  telephone,
		Address:    address,
		City:       city,
		PostalCard: postalcard,
	}
}

func (a *User) Create() {
	a.ID = uuid.NewString()
	a.CreatedAt = time.Now().UTC()
}

func (a *User) Login() {
	a.LoggedInAt = time.Now().UTC()
}

type Role int8

const (
	UNKNOWN       Role = iota
	BASIC         Role = iota
	GUEST         Role = iota
	ADMINISTRATOR Role = iota
)

func (r Role) String() string {
	roles := []string{
		"unknown",
		"basic",
		"guest",
		"administrator",
	}
	return roles[r]
}

// Function qui retourne si un role est superieur (ou egal a un autre role).
func (r Role) IsSuperior(role Role) bool {
	switch r {
	case ADMINISTRATOR:
		return role == ADMINISTRATOR || role == GUEST || role == BASIC
	case GUEST:
		return role == GUEST
	case BASIC:
		return role == BASIC
	default:
		return false
	}
}
