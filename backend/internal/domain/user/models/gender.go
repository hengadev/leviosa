package models

import "fmt"

type Gender string

// TODO: do I need to expose that function ?
func ValidateGender(gender string) error {
	validGenders := map[string]bool{
		"M":  true,
		"F":  true,
		"NB": true,
		"NP": true,
	}
	if !validGenders[gender] {
		return fmt.Errorf("invalid gender value: %s", gender)
	}
	return nil
}

func NewGender(gender string) (Gender, error) {
	if err := ValidateGender(gender); err != nil {
		return "", fmt.Errorf("validate gender")
	}
	return Gender(gender), nil
}

func (g Gender) String() string {
	return string(g)
}
