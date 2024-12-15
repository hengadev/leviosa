package userService

import (
	"context"
	"reflect"

	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

type Credentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func (c Credentials) Valid(ctx context.Context) errsx.Map {
	var pbms = make(errsx.Map)
	vf := reflect.VisibleFields(reflect.TypeOf(c))
	for _, f := range vf {
		switch f.Name {
		case "Email":
			if err := ValidateEmail(c.Email); err != nil {
				pbms.Set("email", err)
			}
		case "Password":
			if err := ValidatePassword(c.Password); err != nil {
				pbms.Set("password", err)
			}
		default:
			continue
		}
	}
	return pbms
}
