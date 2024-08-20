package sqlite_test

import (
	"context"
	// "fmt"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"

	test "github.com/GaryHY/event-reservation-app/tests"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestDecreaseFreeplace(t *testing.T) {
	// TEST:
	// no id found
	// normal case
	// freeplace = zero and then call the function

	t.Setenv("TEST_MIGRATION_PATH", "./migrations/tests")
	tests := []struct {
		id                   string
		expectedRowsAffected int
		wantErr              bool
		version              int64
		name                 string
	}{
		{id: baseEvent2.ID, expectedRowsAffected: 0, wantErr: true, version: 20240820103230, name: "freeplace is equal to 0"},
		{id: test.GenerateRandomString(16), expectedRowsAffected: 0, wantErr: true, version: 20240820103230, name: "id not found"},
		{id: baseEvent.ID, expectedRowsAffected: 1, wantErr: false, version: 20240820103230, name: "normal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := setupRepo(ctx, tt.version, sqlite.NewEventRepository)
			if err != nil {
				t.Errorf("setup event domain stub repository: %s", err)
			}
			rowsAffected, err := repo.DecreaseFreeplace(ctx, tt.id)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, rowsAffected, tt.expectedRowsAffected)
		})
	}
}

func TestGetAllEvents(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "./migrations/tests")
	tests := []struct {
		expectedEvents []*event.Event
		wantErr        bool
		version        int64
		name           string
	}{
		{expectedEvents: nil, wantErr: false, version: 20240820013106, name: "no event in database"},
		{expectedEvents: []*event.Event{baseEvent, baseEvent1, baseEvent2}, wantErr: false, version: 20240820103230, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := setupRepo(ctx, tt.version, sqlite.NewEventRepository)
			if err != nil {
				t.Errorf("setup event domain stub repository: %s", err)
			}
			events, err := repo.GetAllEvents(ctx)
			assert.Equal(t, err != nil, tt.wantErr)
			for i, event := range events {
				assert.ReflectEqual(t, event, tt.expectedEvents[i])
			}
		})
	}
}

func TestGetEventByID(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "./migrations/tests")
	tests := []struct {
		id            string
		expectedEvent *event.Event
		wantErr       bool
		version       int64
		name          string
	}{
		{id: test.GenerateRandomString(12), expectedEvent: nil, wantErr: true, version: 20240820013106, name: "no event in database"},
		{id: "ea1d74e2-1612-47ec-aee9-c6a46b65640f", expectedEvent: baseEvent, wantErr: false, version: 20240820023513, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := setupRepo(ctx, tt.version, sqlite.NewEventRepository)
			if err != nil {
				t.Errorf("setup event domain stub repository: %s", err)
			}
			event, err := repo.GetEventByID(ctx, tt.id)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.ReflectEqual(t, event, tt.expectedEvent)
		})
	}
}

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
			assert.Equal(t, res, tt.expectedString)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}

func TestParseBeginAt(t *testing.T) {
	var zeroValue time.Time

	v := "07/12 07:25:17AM '98 -0700"
	vParsed, _ := time.Parse(time.Layout, v)
	expectedTime := vParsed.String()

	tests := []struct {
		hour         string
		day          int
		month        int
		year         int
		expectedTime string
		wantErr      bool
		name         string
	}{
		{hour: "07:25:17", day: 45, month: 7, year: 1998, expectedTime: zeroValue.String(), wantErr: true, name: "day > 30"},
		{hour: "07:25:17", day: -32, month: 7, year: 1998, expectedTime: zeroValue.String(), wantErr: true, name: "day < 0"},
		{hour: "07:25:17", day: 12, month: 32, year: 1998, expectedTime: zeroValue.String(), wantErr: true, name: "month > 12"},
		{hour: "07:25:17", day: 12, month: 0, year: 1998, expectedTime: zeroValue.String(), wantErr: true, name: "month < 1"},
		{hour: "07:25:17", day: 12, month: 7, year: 1998, expectedTime: expectedTime, wantErr: false, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := sqlite.ExportedParseBeginAt(tt.hour, tt.day, tt.month, tt.year)
			assert.Equal(t, res.String(), tt.expectedTime)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
