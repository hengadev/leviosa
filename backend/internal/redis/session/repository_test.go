package sessionRepository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/redis"
	"github.com/GaryHY/event-reservation-app/internal/redis/session"
)

func newTestRepository(t testing.TB, ctx context.Context, initMap miniredis.InitMap) (*sessionRepository.Repository, error) {
	t.Helper()
	client, err := miniredis.Setup(t, ctx)
	if err != nil {
		return nil, fmt.Errorf("setup miniredis: %w", err)
	}

	if err := miniredis.Init(t, ctx, client, initMap); err != nil {
		return nil, fmt.Errorf("setup miniredis: %w", err)
	}
	return sessionRepository.New(ctx, client), nil
}
