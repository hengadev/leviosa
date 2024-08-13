package main

import (
	"context"
	"database/sql"
	"fmt"

	// api
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/photo"
	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	"github.com/GaryHY/event-reservation-app/internal/server/service"

	// databases
	"github.com/GaryHY/event-reservation-app/internal/redis"
	"github.com/GaryHY/event-reservation-app/internal/s3"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"

	// external packages
	rd "github.com/redis/go-redis/v9"
)

func makeServices(
	ctx context.Context,
	sqlitedb *sql.DB,
	redisdb *rd.Client,
) (handler.Services, handler.Repos, error) {
	var appSvcs handler.Services
	var appRepos handler.Repos

	// user
	userRepo := sqlite.NewUserRepository(ctx, sqlitedb)
	userSvc := user.NewService(userRepo)
	// session
	sessionRepo := redis.NewSessionRepository(ctx, redisdb)
	sessionSvc := session.NewService(sessionRepo)
	// event
	eventRepo := sqlite.NewEventRepository(ctx, sqlitedb)
	eventSvc := event.NewService(eventRepo)
	// vote
	voteRepo := sqlite.NewVoteRepository(ctx, sqlitedb)
	voteSvc := vote.NewService(voteRepo)
	// register
	registerRepo := sqlite.NewRegisterRepository(ctx, sqlitedb)
	registerSvc := register.NewService(registerRepo)
	// photo
	photoRepo, err := s3.NewPhotoRepository(ctx)
	if err != nil {
		return appSvcs, appRepos, fmt.Errorf("create photo repository: %w", err)
	}
	photoSvc := photo.NewService(photoRepo)

	// services
	appSvcs = handler.Services{
		User:     userSvc,
		Event:    eventSvc,
		Vote:     voteSvc,
		Register: registerSvc,
		Photo:    photoSvc,
		Session:  sessionSvc,
	}
	// repos
	appRepos = handler.Repos{
		User:     userRepo,
		Event:    eventRepo,
		Vote:     voteRepo,
		Register: registerRepo,
		Photo:    photoRepo,
		Session:  sessionRepo,
	}
	return appSvcs, appRepos, nil
}
