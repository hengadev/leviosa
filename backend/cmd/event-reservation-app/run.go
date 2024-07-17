package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	// api
	// "github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/photo"
	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	"github.com/GaryHY/event-reservation-app/internal/server"
	"github.com/GaryHY/event-reservation-app/internal/server/service"

	// utils
	"github.com/GaryHY/event-reservation-app/pkg/config"
	"github.com/GaryHY/event-reservation-app/pkg/flags"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"
	// "github.com/GaryHY/event-reservation-app/internal/http/cron"

	// databases
	// "github.com/GaryHY/event-reservation-app/internal/redis"
	"github.com/GaryHY/event-reservation-app/internal/s3"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"

	// external packages
	"github.com/joho/godotenv"
)

var opts struct {
	mode   mode.EnvMode
	server struct {
		port int
	}
}

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	// flags
	flag.IntVar(&opts.server.port, "port", 5000, "the port the server listens to")
	flag.Var(&opts.mode, "mode", "the mode environment for the project")
	flag.Parse()

	// set environment file
	err := godotenv.Load(fmt.Sprintf("%s.env", opts.mode.String()))
	if err != nil {
		return fmt.Errorf("load env variables: %w", err)
	}

	// config
	conf := config.New(ctx, opts.mode.String(), "env")
	if err := conf.Load(ctx); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	sqliteConf := conf.GetSQLITE()

	// databases setup
	sqlitedb, err := sqliteutil.Connect(ctx, sqliteutil.BuildDSN(sqliteConf.Filename))
	if err != nil {
		return fmt.Errorf("create connection to sqlite : %w", err)
	}

	// user
	userRepo := sqlite.NewUserRepository(ctx, sqlitedb)
	userSvc := user.NewService(userRepo)
	// session
	// sessionRepo, err := redis.NewSessionRepository(ctx)
	// if err != nil {
	// 	return fmt.Errorf("create session repo : %w", err)
	// }
	// sessionSvc := session.NewService(sessionRepo)
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
		return fmt.Errorf("create photo repository: %w", err)
	}
	photoSvc := photo.NewService(photoRepo)

	// services
	appSvcs := handler.Services{
		// Session: sessionSvc,
		User:     userSvc,
		Event:    eventSvc,
		Vote:     voteSvc,
		Register: registerSvc,
		Photo:    photoSvc,
	}
	// repos
	appRepos := handler.Repos{
		// Session: sessionRepo,
		User:     userRepo,
		Event:    eventRepo,
		Vote:     voteRepo,
		Register: registerRepo,
		Photo:    photoRepo,
	}

	handler := handler.NewHandler(&appSvcs, &appRepos)
	srv := server.New(
		handler,
		server.WithPort(opts.server.port),
	)
	var srvErrCh = make(chan error)

	// setting cron jobs
	// go func() {
	// 	cronHandler := cron.NewHandler(handler)
	// 	srvErrCh <- cronHandler.Start()
	// }()

	go func() {
		fmt.Fprintf(w, "Running server on port %d...\n", opts.server.port)
		if err := srv.ListenAndServe(); err != nil {
			srvErrCh <- fmt.Errorf("launch server: %w", err)

		}
	}()

	select {
	case done := <-ctx.Done():
		return fmt.Errorf("ctx.Done: %v", done)
	case err := <-srvErrCh:
		return fmt.Errorf("server error: %w", err)
	}
}
