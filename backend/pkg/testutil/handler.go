package testutil

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/redis/session"
	// "github.com/GaryHY/event-reservation-app/internal/server/handler/session"
	"github.com/GaryHY/event-reservation-app/internal/redis"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/user"
	redistest "github.com/GaryHY/event-reservation-app/pkg/redisutil/testdatabase"
	sqlitetest "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
)

// func SetupUser(t testing.TB, ctx context.Context, version int64) *userHandler.Handler {
func SetupUser(t testing.TB, ctx context.Context, version int64) (*userService.Service, *userRepository.Repository) {
	t.Helper()
	sqlitedb, err := sqlitetest.NewDatabase(ctx)
	if err != nil {
		t.Error(err)
	}
	if err := sqlitetest.Setup(ctx, sqlitedb, version); err != nil {
		t.Error(err)
	}
	userRepo := userRepository.New(ctx, sqlitedb)
	userService := userService.New(userRepo)
	return userService, userRepo
}

// func SetupSession(t testing.TB, ctx context.Context, version int64) *userHandler.Handler {
type redisTeardownFunc func()

func SetupSession(t testing.TB, ctx context.Context, version int64) (*sessionService.Service, *sessionRepository.Repository, redisTeardownFunc) {
	t.Helper()
	redisdb, err := redistest.NewTestDatabase(ctx)
	if err != nil {
		t.Error(err)
	}

	sessionRepo := sessionRepository.New(ctx, redisdb.DB)
	sessionService := sessionService.New(sessionRepo)

	teardown := func() {
		redisdb.TearDown()
	}
	return sessionService, sessionRepo, teardown
}

type otherRedisTeardownFunc func() error

// NOTE: the version of the setup that uses miniredis
func OtherSetupSession(t testing.TB, ctx context.Context, initMap miniredis.InitMap[*sessionService.Values]) (*sessionService.Service, *sessionRepository.Repository, redisTeardownFunc) {
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
