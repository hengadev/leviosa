package user

import (
	"time"

	// "where is the role type ?"

	"github.com/google/uuid"
)

// TODO: do better with the address
type Account struct {
	ID         string    `json:"id"`
	Email      string    `json:"email" validate:"required,email"`
	Password   string    `json:"-" validate:"required,min=6"`
	CreatedAt  time.Time `json:"createdat"`
	LoggedInAt time.Time `json:"loggedinat"`
	Role       string    `json:"role"`
	LastName   string    `json:"lastname"`
	FirstName  string    `json:"firstname"`
	Gender     string    `json:"gender"`
	BirthDate  time.Time `json:"birthdate"`
	Telephone  string    `json:"telephone"`
	Address    string    `json:"address"`
	City       string    `json:"city"`
	PostalCard string    `json:"postalcard"`
}

// what is the option pattern in golang ? Can I use it in here since some element are not going to be used to send by the user
func NewAccount(email Email, password Password) *Account {
	return &Account{
		Email:    email.String(),
		Password: password.String(),
		// put the role in a parameter please
		// Role: role.User.String(),
	}
}

func (a *Account) Create() {
	a.ID = uuid.NewString()
	a.CreatedAt = time.Now().UTC()
}

func (a *Account) Login() {
	a.LoggedInAt = time.Now().UTC()
}
