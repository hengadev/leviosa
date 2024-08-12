package user_test

import (
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/tests"
)

func TestString(t *testing.T) {
	tests := []struct {
		role     user.Role
		expected string
		name     string
	}{
		{role: user.ADMINISTRATOR, expected: "admin", name: "Get string admin"},
		{role: user.BASIC, expected: "basic", name: "Get string basic"},
		{role: user.GUEST, expected: "guest", name: "Get string guest"},
		{role: user.UNKNOWN, expected: "unknown", name: "Get string unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.role.String()
			test.Assert(t, got, tt.expected)
		})
	}
}

func TestConvertToRole(t *testing.T) {
	// TODO: comment tester le fait que quelque soit le string send autre que ceux definis je dois avoir comme retour "UNKNOWN pour le role" ?
	tests := []struct {
		roleStr  string
		expected user.Role
		name     string
	}{
		{roleStr: "admin", expected: user.ADMINISTRATOR, name: "Convert to ADMINISTRATOR"},
		{roleStr: "basic", expected: user.BASIC, name: "Convert to BASIC"},
		{roleStr: "guest", expected: user.GUEST, name: "Convert to GUEST"},
		{roleStr: test.GenerateRandomString(5), expected: user.UNKNOWN, name: "Convert to UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := user.ConvertToRole(tt.roleStr)
			test.Assert(t, got, tt.expected)
		})
	}
}

// // TODO: find the equivalent to french "comparant" and "compared" in english
func TestIsSuperior(t *testing.T) {
	tests := []struct {
		comparing user.Role
		compared  user.Role
		expected  bool
		name      string
	}{
		{comparing: user.ADMINISTRATOR, compared: user.ADMINISTRATOR, name: "Administrator is superior to Administrator ?", expected: true},
		{comparing: user.ADMINISTRATOR, compared: user.GUEST, name: "Administration is superior to Guest ?", expected: true},
		{comparing: user.ADMINISTRATOR, compared: user.BASIC, name: "Administration is superior to Basic ?", expected: true},
		{comparing: user.ADMINISTRATOR, compared: user.UNKNOWN, name: "Administration is superior to Unknown ?", expected: false},
		{comparing: user.GUEST, compared: user.GUEST, name: "Guest is superior to Guest ?", expected: true},
		{comparing: user.GUEST, compared: user.ADMINISTRATOR, name: "Guest is superior to Administrator ?", expected: false},
		{comparing: user.GUEST, compared: user.BASIC, name: "Guest is superior to Basic ?", expected: false},
		{comparing: user.GUEST, compared: user.UNKNOWN, name: "Guest is superior to Basic ?", expected: false},
		{comparing: user.BASIC, compared: user.BASIC, name: "Basic is superior to Basic ?", expected: true},
		{comparing: user.BASIC, compared: user.ADMINISTRATOR, name: "Basic is superior to Administrator ?", expected: false},
		{comparing: user.BASIC, compared: user.GUEST, name: "Basic is superior to Guest ?", expected: false},
		{comparing: user.BASIC, compared: user.UNKNOWN, name: "Basic is superior to Unknown ?", expected: false},
		{comparing: user.UNKNOWN, compared: user.UNKNOWN, name: "Unknown is superior to Unknown ?", expected: false},
		{comparing: user.UNKNOWN, compared: user.ADMINISTRATOR, name: "Unknown is superior to Administrator ?", expected: false},
		{comparing: user.UNKNOWN, compared: user.GUEST, name: "Unknown is superior to Guest ?", expected: false},
		{comparing: user.UNKNOWN, compared: user.BASIC, name: "Unknown is superior to Basic ?", expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.comparing.IsSuperior(tt.compared)
			test.Assert(t, got, tt.expected)
		})
	}
}
