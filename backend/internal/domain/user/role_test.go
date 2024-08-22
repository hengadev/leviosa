package userService_test

import (
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/tests"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestString(t *testing.T) {
	tests := []struct {
		role     userService.Role
		expected string
		name     string
	}{
		{role: userService.ADMINISTRATOR, expected: "admin", name: "Get string admin"},
		{role: userService.BASIC, expected: "basic", name: "Get string basic"},
		{role: userService.GUEST, expected: "guest", name: "Get string guest"},
		{role: userService.UNKNOWN, expected: "unknown", name: "Get string unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.role.String()
			assert.Equal(t, got, tt.expected)
		})
	}
}

func TestConvertToRole(t *testing.T) {
	// TODO: comment tester le fait que quelque soit le string send autre que ceux definis je dois avoir comme retour "UNKNOWN pour le role" ?
	tests := []struct {
		roleStr  string
		expected userService.Role
		name     string
	}{
		{roleStr: "admin", expected: userService.ADMINISTRATOR, name: "Convert to ADMINISTRATOR"},
		{roleStr: "basic", expected: userService.BASIC, name: "Convert to BASIC"},
		{roleStr: "guest", expected: userService.GUEST, name: "Convert to GUEST"},
		{roleStr: test.GenerateRandomString(5), expected: userService.UNKNOWN, name: "Convert to UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := userService.ConvertToRole(tt.roleStr)
			assert.Equal(t, got, tt.expected)
		})
	}
}

// // TODO: find the equivalent to french "comparant" and "compared" in english
func TestIsSuperior(t *testing.T) {
	tests := []struct {
		comparing userService.Role
		compared  userService.Role
		expected  bool
		name      string
	}{
		{comparing: userService.ADMINISTRATOR, compared: userService.ADMINISTRATOR, name: "Administrator is superior to Administrator ?", expected: true},
		{comparing: userService.ADMINISTRATOR, compared: userService.GUEST, name: "Administration is superior to Guest ?", expected: true},
		{comparing: userService.ADMINISTRATOR, compared: userService.BASIC, name: "Administration is superior to Basic ?", expected: true},
		{comparing: userService.ADMINISTRATOR, compared: userService.UNKNOWN, name: "Administration is superior to Unknown ?", expected: false},
		{comparing: userService.GUEST, compared: userService.GUEST, name: "Guest is superior to Guest ?", expected: true},
		{comparing: userService.GUEST, compared: userService.ADMINISTRATOR, name: "Guest is superior to Administrator ?", expected: false},
		{comparing: userService.GUEST, compared: userService.BASIC, name: "Guest is superior to Basic ?", expected: false},
		{comparing: userService.GUEST, compared: userService.UNKNOWN, name: "Guest is superior to Basic ?", expected: false},
		{comparing: userService.BASIC, compared: userService.BASIC, name: "Basic is superior to Basic ?", expected: true},
		{comparing: userService.BASIC, compared: userService.ADMINISTRATOR, name: "Basic is superior to Administrator ?", expected: false},
		{comparing: userService.BASIC, compared: userService.GUEST, name: "Basic is superior to Guest ?", expected: false},
		{comparing: userService.BASIC, compared: userService.UNKNOWN, name: "Basic is superior to Unknown ?", expected: false},
		{comparing: userService.UNKNOWN, compared: userService.UNKNOWN, name: "Unknown is superior to Unknown ?", expected: false},
		{comparing: userService.UNKNOWN, compared: userService.ADMINISTRATOR, name: "Unknown is superior to Administrator ?", expected: false},
		{comparing: userService.UNKNOWN, compared: userService.GUEST, name: "Unknown is superior to Guest ?", expected: false},
		{comparing: userService.UNKNOWN, compared: userService.BASIC, name: "Unknown is superior to Basic ?", expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.comparing.IsSuperior(tt.compared)
			assert.Equal(t, got, tt.expected)
		})
	}
}
