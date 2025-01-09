package main

import (
	"context"
	"database/sql"
	"fmt"

	// domain
	"github.com/GaryHY/leviosa/internal/domain/event"
	"github.com/GaryHY/leviosa/internal/domain/mail"
	"github.com/GaryHY/leviosa/internal/domain/media"
	"github.com/GaryHY/leviosa/internal/domain/otp"
	"github.com/GaryHY/leviosa/internal/domain/product"
	"github.com/GaryHY/leviosa/internal/domain/register"
	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/domain/stripe"
	"github.com/GaryHY/leviosa/internal/domain/throttler"
	"github.com/GaryHY/leviosa/internal/domain/user"
	"github.com/GaryHY/leviosa/internal/domain/vote"
	"github.com/GaryHY/leviosa/internal/server/app"

	// repositories
	"github.com/GaryHY/leviosa/internal/repository/redis/otp"
	"github.com/GaryHY/leviosa/internal/repository/redis/session"
	"github.com/GaryHY/leviosa/internal/repository/redis/throttler"
	"github.com/GaryHY/leviosa/internal/repository/s3"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/event"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/product"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/register"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/user"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/vote"

	// config
	"github.com/GaryHY/leviosa/pkg/config"

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
	registerSvc := registerService.NewService(registerRepo)
	// stripe
	stripeSvc := stripeService.New()
	// photo
	mediaRepo, err := mediaRepository.NewRepository(ctx)
	if err != nil {
		return appSvcs, appRepos, fmt.Errorf("create photo repository: %w", err)
	}
	mediaSvc := mediaService.New(mediaRepo)

	mailSvc := mailService.New()

	productRepo := productRepository.New(ctx, sqlitedb)
	productSvc := productService.New(productRepo)

	throttlerRepo := throttlerRepository.New(ctx, redisdb)
	throttlerSvc := throttlerService.New(throttlerRepo)

	otpRepo := otpRepository.New(ctx, redisdb)
	otpSvc := otpService.New(otpRepo)

	// services
	appSvcs = app.Services{
		User:      userSvc,
		Event:     eventSvc,
		Vote:      voteSvc,
		Register:  registerSvc,
		Media:     mediaSvc,
		Session:   sessionSvc,
		Throttler: throttlerSvc,
		Mail:      mailSvc,
		Stripe:    stripeSvc,
		Product:   productSvc,
		OTP:       otpSvc,
	}
	// repos
	appRepos = app.Repos{
		User:      userRepo,
		Event:     eventRepo,
		Vote:      voteRepo,
		Register:  registerRepo,
		Media:     mediaRepo,
		Session:   sessionRepo,
		Throttler: throttlerRepo,
		Product:   productRepo,
		OTP:       otpRepo,
	}
	return appSvcs, appRepos, nil
}
