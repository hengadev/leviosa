package user

import (
	"context"
	"reflect"
)

type Credentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func (c Credentials) Valid(ctx context.Context) map[string]string {
	var pbms = make(map[string]string)
	vf := reflect.VisibleFields(reflect.TypeOf(c))
	for _, f := range vf {
		switch f.Name {
		case "Email":
			if err := ValidateEmail(c.Email); err != nil {
				pbms["email"] = err.Error()
			}
		case "Password":
			if err := ValidatePassword(c.Password); err != nil {
				pbms["password"] = err.Error()
			}
		default:
			continue
		}
	}
	return pbms
}
