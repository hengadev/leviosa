package userService

import (
	"fmt"
	"strings"
)

func convIntToMonth(month int) (string, error) {
	return convIntToStr(month, 13)
}

func convIntToDay(day, month int) (string, error) {
	var maxValue int
	if month == 2 {
		maxValue = 28
	} else if month%2 == 0 {
		maxValue = 30
	} else {
		maxValue = 31
	}
	return convIntToStr(day, maxValue)
}

func convIntToStr(value, maxValue int) (string, error) {
	if value < 0 {
		return "", fmt.Errorf("%d is negative value", value)
	} else if value < 10 {
		return fmt.Sprintf("0%d", value), nil
	} else if value < maxValue {
		return fmt.Sprintf("%d", value), nil
	} else {
		return "", fmt.Errorf("%d is too large value", value)
	}
}

// TODO: Be more specific on the cases where the user has gender other than male, female
func convToGender(gender string) string {
	var formattedGender string
	switch gender {
	case "male", "female":
		formattedGender = fmt.Sprintf("%s", strings.ToUpper(gender[0:1]))
	default:
		return "NB"
	}
	return formattedGender
}
