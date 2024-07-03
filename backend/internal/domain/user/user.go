package user

import (
	"reflect"
	"time"

	"github.com/google/uuid"
)

// TODO: do better with the address
type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email" validate:"required,email"`
	Password   string    `json:"-" validate:"required,min=6"`
	CreatedAt  time.Time `json:"-"`
	LoggedInAt time.Time `json:"-"`
	Role       string    `json:"-"`
	BirthDate  time.Time `json:"birthdate"`
	LastName   string    `json:"lastname"`
	FirstName  string    `json:"firstname"`
	Gender     string    `json:"gender"`
	Telephone  string    `json:"telephone"`
	Address    string    `json:"address"`
	City       string    `json:"city"`
	PostalCard int       `json:"postalcard"`
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
	city string,
	postalcard int,
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

// do some generic fucntion to that so that I can use it for all function
// that function is not something that I want, when I have a user, it should already be checked.
func (u *User) Validate() map[string]string {
	var pbms = make(map[string]string)
	vf := reflect.VisibleFields(reflect.TypeOf(u))
	// PERF: use concurrency for that ? How about the concurrent writing on the map ?
	for _, f := range vf {
		switch f.Name {
		case "Email":
			if err := ValidateEmail(u.Email); err != nil {
				pbms["email"] = err.Error()
			}
		case "Password":
			if err := ValidatePassword(u.Password); err != nil {
				pbms["password"] = err.Error()
			}
		case "Telephone":
			// do the validation using the rule that follows :
			// if len(u.Telephone) < 10 && strings.HasPrefix(u.Telephone) {
			// 	pbms["telephone"] = ""
			// }
		default:
			continue
		}
	}
	return pbms
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
