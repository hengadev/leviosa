package sqlite_test

import (
	// "context"
	"fmt"
	"testing"
	// "time"

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

func TestFormatTime(t *testing.T) {
	tests := []struct {
		value          string
		expectedString string
		wantErr        bool
		name           string
	}{
		{value: "wrong format", expectedString: "", wantErr: true, name: "invalid time, not even letters"},
		{value: "08:12:57", expectedString: "08:12:57AM", wantErr: false, name: "valid morning time"},
		{value: "16:34:23", expectedString: "04:34:23PM", wantErr: false, name: "valid afternoon time"},
		{value: "33:34:23", expectedString: "", wantErr: true, name: "invalid time, hour > 24"},
		{value: "-23:34:23", expectedString: "", wantErr: true, name: "invalid time, hour < 0"},
		{value: "09:-23:23", expectedString: "", wantErr: true, name: "invalid time, minute < 0"},
		{value: "11:68:23", expectedString: "", wantErr: true, name: "invalid time, minute > 60"},
		{value: "18:15:-43", expectedString: "", wantErr: true, name: "invalid time, second < 0"},
		{value: "19:32:72", expectedString: "", wantErr: true, name: "invalid time, second > 60"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := sqlite.ExportedFormatTime(tt.value)
			fmt.Printf("the resulting string is : %q\n", res)
			fmt.Println("the err is:", err)
			assert.Equal(t, res, tt.expectedString)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
