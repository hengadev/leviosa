package userHandler_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	userhandler "github.com/GaryHY/event-reservation-app/internal/server/handler/user"
	handler "github.com/GaryHY/event-reservation-app/internal/server/service"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/user"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
)

func Setup(t testing.TB, ctx context.Context, version int64) *userhandler.Handler {
	t.Helper()
	sqlitedb, err := testdb.NewDatabase(ctx)
	if err != nil {
		t.Error(err)
	}
	if err := testdb.Setup(ctx, sqlitedb, version); err != nil {
		t.Error(err)
	}
	// readerRepo := userRepository.NewReaderRepository(ctx, db)
	// userRepo := userRepository.New(ctx, readerRepo)
	userRepo := userRepository.New(ctx, sqlitedb)
	userService := userService.New(userRepo)
	appsvc := handler.Services{User: userService}
	// apprepo := handler.Repos{User: readerRepo}
	apprepo := handler.Repos{User: userRepo}
	h := handler.NewHandler(&appsvc, &apprepo)
	return userhandler.New(h)
}
