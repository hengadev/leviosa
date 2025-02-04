package sessionRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/repository/redis"
	"github.com/GaryHY/leviosa/pkg/testutil"

	"github.com/GaryHY/test-assert"
)

func TestRemoveSession(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
		init    miniredis.InitMap[*sessionService.Values]
		name    string
	}{
		{id: testutil.BaseSession.ID, wantErr: true, init: nil, name: "empty database"},
		{id: testutil.RandomSessionID, wantErr: true, init: testutil.InitSession, name: "id not in the database"},
		{id: testutil.BaseSession.ID, wantErr: false, init: testutil.InitSession, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := newTestRepository(t, ctx, tt.init)
			if err != nil {
				t.Errorf("setup repository: %s", err)
			}
			err = repo.RemoveSession(ctx, tt.id)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
