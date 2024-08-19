package sqlite_test

import (
	// "context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestConvIntToStr(t *testing.T) {
	tests := []struct {
		value          int
		expectedString string
		wantErr        bool
		name           string
	}{
		{value: -8, expectedString: "", wantErr: true, name: "Invalid case with negative value"},
		{value: 3234, expectedString: "", wantErr: true, name: "Invalid case with value > 99 value"},
		{value: 5, expectedString: "05", wantErr: false, name: "Value inferior than 10"},
		{value: 12, expectedString: "12", wantErr: false, name: "Value superior than 10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			str, err := sqlite.ExportedConvIntToStr(tt.value)
			assert.Equal(t, str, tt.expectedString)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
