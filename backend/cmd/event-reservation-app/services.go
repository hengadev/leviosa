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
	"github.com/GaryHY/event-reservation-app/internal/redis/session"
	"github.com/GaryHY/event-reservation-app/internal/s3"

	// repositories
	"github.com/GaryHY/event-reservation-app/internal/sqlite/event"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/register"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/user"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/vote"

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
	userRepo := userRepository.New(ctx, sqlitedb)
	userSvc := userService.New(userRepo)
	// session
	sessionRepo := sessionRepository.New(ctx, redisdb)
	sessionSvc := session.NewService(sessionRepo)
	// event
	eventRepo := eventRepository.New(ctx, sqlitedb)
	eventSvc := event.NewService(eventRepo)
	// vote
	voteRepo := voteRepository.New(ctx, sqlitedb)
	voteSvc := vote.NewService(voteRepo)
	// register
	registerRepo := registerRepository.New(ctx, sqlitedb)
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
