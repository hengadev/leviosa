package validation

import (
	"regexp"
	"strings"

	"github.com/hengadev/errsx"
)

// Phone validation constants
const (
	PhoneMinLength = 10
	PhoneMaxLength = 20
)

// Phone validation error keys
const (
	PhoneTooShortKey = "phone_too_short"
	PhoneTooLongKey  = "phone_too_long"
	PhoneFormatKey   = "phone_format"
)

// Phone validation error messages
const (
	PhoneTooShortMsg = "phone must be at least 10 characters"
	PhoneTooLongMsg  = "phone cannot exceed 20 characters"
	PhoneFormatMsg   = "invalid phone number format"
)

// French phone number validation (can be extended for international later)
var frenchPhoneStrict = regexp.MustCompile(`^(0[1-5]|06|07)\d{8}$`)

func ValidatePhone(phone string) error {
	var errs errsx.Map

	// Trim whitespace for validation
	trimmed := strings.TrimSpace(phone)

	// Length checks
	if len(trimmed) < PhoneMinLength {
		errs.Set(PhoneTooShortKey, PhoneTooShortMsg)
	}
	if len(trimmed) > PhoneMaxLength {
		errs.Set(PhoneTooLongKey, PhoneTooLongMsg)
	}

	// French phone format validation
	if !frenchPhoneStrict.MatchString(trimmed) {
		errs.Set(PhoneFormatKey, PhoneFormatMsg)
	}

	return errs.AsError()
}

// Guest phone numbers on the public booking form: a French number in the
// compact domestic form (0 then 9 digits; 08 and 09 are valid French numbers
// too), or any other country in E.164 (+ then the country code and number,
// 8 to 15 digits in all). The frontend checks the number against its country;
// this only checks the shape it is stored in.
var (
	frenchPhoneDomestic = regexp.MustCompile(`^0[1-9]\d{8}$`)
	e164Phone           = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
)

// ValidateGuestPhone accepts what ValidatePhone accepts plus international
// numbers in E.164, for clients who book without a French number.
func ValidateGuestPhone(phone string) error {
	trimmed := strings.TrimSpace(phone)
	if frenchPhoneDomestic.MatchString(trimmed) || e164Phone.MatchString(trimmed) {
		return nil
	}
	var errs errsx.Map
	errs.Set(PhoneFormatKey, PhoneFormatMsg)
	return errs.AsError()
}
