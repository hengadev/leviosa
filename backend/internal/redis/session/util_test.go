package sessionRepository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/redis"
	"github.com/GaryHY/event-reservation-app/internal/redis/session"
)

const sessionID = "a0rg34tWfQ33009_K"

var sessionTime, _ = time.Parse(time.Layout, "07/12 11:00:00AM '98 -0700")

var baseSession = session.Session{
	ID:         sessionID,
	UserID:     1,
	Role:       "basic",
	LoggedInAt: sessionTime,
	CreatedAt:  sessionTime,
	ExpiresAt:  sessionTime.Add(time.Hour),
}

var initSession = map[string]any{
	baseSession.ID: baseSession.Values(),
}

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
