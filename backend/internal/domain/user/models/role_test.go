package models_test

import (
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/tests"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestString(t *testing.T) {
	tests := []struct {
		role     models.Role
		expected string
		name     string
	}{
		{role: models.ADMINISTRATOR, expected: "admin", name: "Get string admin"},
		{role: models.BASIC, expected: "basic", name: "Get string basic"},
		{role: models.GUEST, expected: "guest", name: "Get string guest"},
		{role: models.UNKNOWN, expected: "unknown", name: "Get string unknown"},
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
		expected models.Role
		name     string
	}{
		{roleStr: "admin", expected: models.ADMINISTRATOR, name: "Convert to ADMINISTRATOR"},
		{roleStr: "basic", expected: models.BASIC, name: "Convert to BASIC"},
		{roleStr: "guest", expected: models.GUEST, name: "Convert to GUEST"},
		{roleStr: test.GenerateRandomString(5), expected: models.UNKNOWN, name: "Convert to UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.ConvertToRole(tt.roleStr)
			assert.Equal(t, got, tt.expected)
		})
	}
}

// // TODO: find the equivalent to french "comparant" and "compared" in english
func TestIsSuperior(t *testing.T) {
	tests := []struct {
		comparing models.Role
		compared  models.Role
		expected  bool
		name      string
	}{
		{comparing: models.ADMINISTRATOR, compared: models.ADMINISTRATOR, name: "Administrator is superior to Administrator ?", expected: true},
		{comparing: models.ADMINISTRATOR, compared: models.GUEST, name: "Administration is superior to Guest ?", expected: true},
		{comparing: models.ADMINISTRATOR, compared: models.BASIC, name: "Administration is superior to Basic ?", expected: true},
		{comparing: models.ADMINISTRATOR, compared: models.UNKNOWN, name: "Administration is superior to Unknown ?", expected: false},
		{comparing: models.GUEST, compared: models.GUEST, name: "Guest is superior to Guest ?", expected: true},
		{comparing: models.GUEST, compared: models.ADMINISTRATOR, name: "Guest is superior to Administrator ?", expected: false},
		{comparing: models.GUEST, compared: models.BASIC, name: "Guest is superior to Basic ?", expected: false},
		{comparing: models.GUEST, compared: models.UNKNOWN, name: "Guest is superior to Basic ?", expected: false},
		{comparing: models.BASIC, compared: models.BASIC, name: "Basic is superior to Basic ?", expected: true},
		{comparing: models.BASIC, compared: models.ADMINISTRATOR, name: "Basic is superior to Administrator ?", expected: false},
		{comparing: models.BASIC, compared: models.GUEST, name: "Basic is superior to Guest ?", expected: false},
		{comparing: models.BASIC, compared: models.UNKNOWN, name: "Basic is superior to Unknown ?", expected: false},
		{comparing: models.UNKNOWN, compared: models.UNKNOWN, name: "Unknown is superior to Unknown ?", expected: false},
		{comparing: models.UNKNOWN, compared: models.ADMINISTRATOR, name: "Unknown is superior to Administrator ?", expected: false},
		{comparing: models.UNKNOWN, compared: models.GUEST, name: "Unknown is superior to Guest ?", expected: false},
		{comparing: models.UNKNOWN, compared: models.BASIC, name: "Unknown is superior to Basic ?", expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.comparing.IsSuperior(tt.compared)
			assert.Equal(t, got, tt.expected)
		})
	}
}
