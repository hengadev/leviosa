package models

import "fmt"

// ProviderType defines a restricted set of provider types.
type ProviderType string

// Valid provider types
const (
	Google ProviderType = "google"
	Mail   ProviderType = "email"
	Apple  ProviderType = "apple"
)

// validProviders is used for easy iteration or validation
var validProviders = []string{
	string(Google),
	string(Mail),
	string(Apple),
}

// Set assigns a valid value to ProviderType or returns an error
func (m *ProviderType) Set(value string) error {
	switch value {
	case "apple":
		*m = Apple
	case "google":
		*m = Google
	case "mail":
		*m = Mail
	default:
		return fmt.Errorf("provider type value can only be 'apple', 'google' or 'mail', got : %q", value)
	}
	return nil
}

// IsValid checks if the current value of ProviderType is valid
func (p ProviderType) IsValid() bool {
	for _, v := range validProviders {
		if string(p) == v {
			return true
		}
	}
	return false
}
