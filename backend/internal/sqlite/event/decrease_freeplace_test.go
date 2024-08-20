package eventRepository_test

// import (
// 	"context"
// 	"testing"
//
// 	"github.com/GaryHY/event-reservation-app/internal/sqlite/event"
// 	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
// 	"github.com/GaryHY/event-reservation-app/tests"
// 	"github.com/GaryHY/event-reservation-app/tests/assert"
// )

// func TestDecreaseFreeplace(t *testing.T) {
// 	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
// 	tests := []struct {
// 		id                   string
// 		expectedRowsAffected int
// 		wantErr              bool
// 		version              int64
// 		name                 string
// 	}{
// 		{id: baseEvent2.ID, expectedRowsAffected: 0, wantErr: true, version: 20240820103230, name: "freeplace is equal to 0"},
// 		{id: test.GenerateRandomString(16), expectedRowsAffected: 0, wantErr: true, version: 20240820103230, name: "id not found"},
// 		{id: baseEvent.ID, expectedRowsAffected: 1, wantErr: false, version: 20240820103230, name: "normal case"},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			ctx := context.Background()
// 			repo, err := testdb.SetupRepo(ctx, tt.version, eventRepository.New)
// 			defer testdb.Teardown(repo.DB)
// 			if err != nil {
// 				t.Errorf("setup event domain stub repository: %s", err)
// 			}
// 			rowsAffected, err := repo.DecreaseFreeplace(ctx, tt.id)
// 			assert.Equal(t, err != nil, tt.wantErr)
// 			assert.Equal(t, rowsAffected, tt.expectedRowsAffected)
// 		})
// 	}
// }
