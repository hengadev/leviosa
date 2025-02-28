package models

import (
	"fmt"

	"github.com/hengadev/leviosa/pkg/errsx"
)

type Password string

const passwordMinLength = 8

func ValidatePassword(p string) errsx.Map {
	var errs errsx.Map
	if len(p) < passwordMinLength {
		errs.Set("password length", fmt.Sprintf("expect at least %d caracter", passwordMinLength))
	}
	return errs
}
func NewPassword(p string) (Password, errsx.Map) {
	var errs errsx.Map
	if pbms := ValidatePassword(p); len(pbms) > 0 {
		errs.Set("validate password", pbms)
	}
	return Password(p), errs
}

func (p Password) String() string {
	return string(p)
}
