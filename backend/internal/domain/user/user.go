package userService

import (
	"context"
	"reflect"
	"time"

	"github.com/google/uuid"
)

const BirthdayLayout = "2006-01-02"

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email,omitempty" validate:"required,email"`
	Password      string    `json:"password" validate:"required,min=6"`
	CreatedAt     time.Time `json:"createdAt"`
	LoggedInAt    time.Time `json:"loggedinat"`
	Role          string    `json:"role"`
	BirthDate     string    `json:"birthdate,omitempty"`
	LastName      string    `json:"lastname,omitempty"`
	FirstName     string    `json:"firstname,omitempty"`
	Gender        string    `json:"gender,omitempty"`
	Telephone     string    `json:"telephone,omitempty"`
	OAuthProvider string    `json:"oauthProvider"`
	OAuthID       string    `json:"oauthId"`
}

type UserResponse struct {
	Email     string `json:"email,omitempty" validate:"required,email"`
	BirthDate string `json:"birthdate,omitempty"`
	LastName  string `json:"lastname,omitempty"`
	FirstName string `json:"firstname,omitempty"`
	Gender    string `json:"gender,omitempty"`
	Telephone string `json:"telephone,omitempty"`
}

func NewUser(
	email Email,
	password Password,
	birthdate string,
	lastname,
	firstname,
	gender string,
	telephone Telephone,
) *User {
	return &User{
		ID:        uuid.NewString(),
		Email:     email.String(),
		Password:  password.String(),
		Role:      BASIC.String(),
		BirthDate: birthdate,
		LastName:  lastname,
		FirstName: firstname,
		Gender:    gender,
		Telephone: telephone.String(),
	}
}

func (a *User) Create() {
	a.CreatedAt = time.Now().UTC()
}

func (a *User) Login() {
	a.LoggedInAt = time.Now().UTC()
}

func (u User) Valid(ctx context.Context) map[string]string {
	var pbms = make(map[string]string)
	vf := reflect.VisibleFields(reflect.TypeOf(u))
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
		case "Birthday":
			parsedDate, err := time.Parse(BirthdayLayout, u.BirthDate)
			nonValidDate, _ := time.Parse(BirthdayLayout, "01-01-01")
			if err != nil && parsedDate != nonValidDate {
				pbms["birthday"] = err.Error()
			}
		default:
			continue
		}
	}
	return pbms
}

func (u User) ToResponse() UserResponse {
	return UserResponse{
		Email:     u.Email,
		BirthDate: u.BirthDate,
		LastName:  u.LastName,
		FirstName: u.FirstName,
		Gender:    u.Gender,
		Telephone: u.Telephone,
	}
}
