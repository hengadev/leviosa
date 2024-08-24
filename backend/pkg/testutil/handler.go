package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/redis"
	"github.com/GaryHY/event-reservation-app/internal/redis/session"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/user"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
)

func SetupUser(t testing.TB, ctx context.Context, version int64) (*userService.Service, *userRepository.Repository) {
	t.Helper()
	db, err := testdb.NewDatabase(ctx)
	if err != nil {
		t.Error(err)
	}
	if err := testdb.Setup(ctx, db, version); err != nil {
		t.Error(err)
	}
	userRepo := userRepository.New(ctx, db)
	userService := userService.New(userRepo)
	return userService, userRepo
}

type redisTeardownFunc func()

func SetupSession(t testing.TB, ctx context.Context, initMap miniredis.InitMap[*sessionService.Values]) (*sessionService.Service, *sessionRepository.Repository, redisTeardownFunc) {
	t.Helper()
	client, err := miniredis.Setup(t, ctx)
	if err != nil {
		t.Errorf("setup miniredis: %s", err)
	}
	if err := miniredis.Init(t, ctx, client, initMap); err != nil {
		t.Errorf("init miniredis: %s", err)
	}
	sessionRepo := sessionRepository.New(ctx, client)
	sessionService := sessionService.New(sessionRepo)
	teardown := func() {
		if err := client.Close(); err != nil {
			t.Errorf("closing miniredis: %s", err)
		}
	}
	return sessionService, sessionRepo, teardown
}
