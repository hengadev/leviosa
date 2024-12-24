package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	// api
	"github.com/GaryHY/event-reservation-app/internal/server"
	"github.com/GaryHY/event-reservation-app/internal/server/app"

	// utils
	"github.com/GaryHY/event-reservation-app/pkg/config"
	"github.com/GaryHY/event-reservation-app/pkg/flags"

	// "github.com/GaryHY/event-reservation-app/internal/http/cron"

	// external packages
	"github.com/joho/godotenv"
)

var opts struct {
	mode   mode.EnvMode
	server struct {
		port int
	}
	logger struct {
		style string
		level string
	}
}

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// always load the .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("load env variables: %w", err)
	}

	// setup env variables
	if err := setupEnvVars(); err != nil {
		return fmt.Errorf("failed to get env variables: %w", err)
	}
	// set basic logger
	logger, err := setLogger()
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	// set environment file
	if opts.mode != mode.ModeProd {
		if err := godotenv.Load(fmt.Sprintf("%s.env", opts.mode.String())); err != nil {
			return fmt.Errorf("load env variables: %w", err)
		}
	}
	// config
	conf := config.New(ctx, opts.mode.String(), "env")
	if errs := conf.Load(ctx); len(errs) > 0 {
		return fmt.Errorf("load configuration: %s", errs.Error())
	}

	// auth, err := conf.NewAuthenticator()
	// if err != nil {
	// 	return fmt.Errorf("setup authenticator: %w", err)
	// }

	sqlitedb, redisdb, err := setupDatabases(ctx, conf)
	if err != nil {
		return fmt.Errorf("setup databases: %w", err)
	}

	appSvcs, appRepos, err := makeServices(
		ctx,
		sqlitedb,
		redisdb,
		conf,
	)
	if err != nil {
		return fmt.Errorf("create services: %w", err)
	}
	appCtx := app.New(&appSvcs, &appRepos)
	srv := server.New(
		appCtx,
		logger,
		server.WithPort(opts.server.port),
	)
	var srvErrCh = make(chan error)

	// setting cron jobs
	// go func() {
	// 	cronHandler := cron.NewHandler(handler)
	// 	srvErrCh <- cronHandler.Start()
	// }()

	go func() {
		logger.Info(fmt.Sprintf("Running server on port %d...\n", opts.server.port))
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
