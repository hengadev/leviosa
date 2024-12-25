package main

import (
	"context"
	"database/sql"
	"fmt"

	// domain
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/mail"
	"github.com/GaryHY/event-reservation-app/internal/domain/photo"
	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/stripe"
	"github.com/GaryHY/event-reservation-app/internal/domain/throttler"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	"github.com/GaryHY/event-reservation-app/internal/server/app"

	// repositories
	"github.com/GaryHY/event-reservation-app/internal/repository/redis/session"
	"github.com/GaryHY/event-reservation-app/internal/repository/redis/throttler"
	"github.com/GaryHY/event-reservation-app/internal/repository/s3"
	"github.com/GaryHY/event-reservation-app/internal/repository/sqlite/event"
	"github.com/GaryHY/event-reservation-app/internal/repository/sqlite/register"
	"github.com/GaryHY/event-reservation-app/internal/repository/sqlite/user"
	"github.com/GaryHY/event-reservation-app/internal/repository/sqlite/vote"

	// config
	"github.com/GaryHY/event-reservation-app/pkg/config"

	// external packages
	rd "github.com/redis/go-redis/v9"
)

func makeServices(
	ctx context.Context,
	sqlitedb *sql.DB,
	redisdb *rd.Client,
	config *config.Config,
) (app.Services, app.Repos, error) {
	var appSvcs app.Services
	var appRepos app.Repos

	// user
	userRepo := userRepository.New(ctx, sqlitedb)
	userSvc := userService.New(userRepo, config.GetSecurity())
	// session
	sessionRepo := sessionRepository.New(ctx, redisdb)
	sessionSvc := sessionService.New(sessionRepo)
	// event
	eventRepo := eventRepository.New(ctx, sqlitedb)
	eventSvc := eventService.New(eventRepo)
	// vote
	voteRepo := voteRepository.New(ctx, sqlitedb)
	voteSvc := vote.NewService(voteRepo)
	// register
	registerRepo := registerRepository.New(ctx, sqlitedb)
	registerSvc := register.NewService(registerRepo)
	// stripe
	stripeSvc := stripeService.New()
	// photo
	photoRepo, err := mediaRepository.NewPhotoRepository(ctx)
	if err != nil {
		return appSvcs, appRepos, fmt.Errorf("create photo repository: %w", err)
	}
	photoSvc := photo.NewService(photoRepo)

	mailSvc := mailService.New()

	throttlerRepo := throttlerRepository.New(ctx, redisdb)
	throttlerSvc := throttlerService.New(throttlerRepo)

	// services
	appSvcs = app.Services{
		User:      userSvc,
		Event:     eventSvc,
		Vote:      voteSvc,
		Register:  registerSvc,
		Photo:     photoSvc,
		Session:   sessionSvc,
		Throttler: throttlerSvc,
		Mail:      mailSvc,
		Stripe:    stripeSvc,
	}
	// repos
	appRepos = app.Repos{
		User:      userRepo,
		Event:     eventRepo,
		Vote:      voteRepo,
		Register:  registerRepo,
		Photo:     photoRepo,
		Session:   sessionRepo,
		Throttler: throttlerRepo,
	}
	return appSvcs, appRepos, nil
}
