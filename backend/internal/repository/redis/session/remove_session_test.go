package sessionRepository_test

import (
	"context"
	"testing"

	sessionService "github.com/GaryHY/leviosa/internal/domain/session"
	miniredis "github.com/GaryHY/leviosa/internal/repository/redis"
	sessionRepository "github.com/GaryHY/leviosa/internal/repository/redis/session"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	assert "github.com/GaryHY/test-assert"
)

func TestRemoveSession(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
		initMap miniredis.InitMap[*sessionService.Values]
		name    string
	}{
		{id: factories.BaseSession.ID, wantErr: true, initMap: nil, name: "empty database"},
		{id: factories.RandomSessionID, wantErr: true, initMap: factories.InitSession, name: "id not in the database"},
		{id: factories.BaseSession.ID, wantErr: false, initMap: factories.InitSession, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := miniredis.SetupRepository(t, ctx, sessionRepository.SESSIONPREFIX, tt.initMap, sessionRepository.New)
			err := repo.RemoveSession(ctx, tt.id)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
