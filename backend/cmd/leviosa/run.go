package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	// api
	"github.com/hengadev/leviosa/internal/server"
	"github.com/hengadev/leviosa/internal/server/app"

	// "github.com/hengadev/leviosa/internal/server/cron"
	"github.com/hengadev/leviosa/pkg/config"
	"github.com/hengadev/leviosa/pkg/flags"

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

	// set logger
	slogHandler, err := setLogger()
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	// set environment file (using [mode].env for specified mode)
	if opts.mode == mode.ModeDev {
		if err := godotenv.Load(fmt.Sprintf("%s.env", opts.mode.String())); err != nil {
			return fmt.Errorf("loading env variables: %w", err)
		}
	}
	// config
	conf := config.New(ctx, opts.mode.String(), "env")
	if errs := conf.Load(ctx, opts.mode); len(errs) > 0 {
		return fmt.Errorf("loading application configuration: %s", errs.Error())
	}

	sqlitedb, redisdb, err := setupDatabases(ctx, conf, opts.mode)
	if err != nil {
		return fmt.Errorf("setting up databases: %w", err)
	}

	appSvcs, appRepos, err := makeServices(ctx, sqlitedb, redisdb, conf)
	if err != nil {
		return fmt.Errorf("create services: %w", err)
	}
	appCtx := app.New(&appSvcs, &appRepos)
	srv := server.New(
		appCtx,
		opts.mode,
		slogHandler,
		server.WithPort(opts.server.port),
	)
	var srvErrCh = make(chan error)

	// setting cron jobs
	// go func() {
	// 	cronHandler := cron.New(handler, logger)
	// 	if err := cronHandler.Start(); err != nil {
	// 		srvErrCh <- fmt.Errorf("cron service failed: %w", err)
	// 		return
	// 	}
	// }()

	go func() {
		slog.InfoContext(ctx, fmt.Sprintf("Server running on port %d.", opts.server.port))
		if err := srv.ListenAndServe(); err != nil {
			srvErrCh <- fmt.Errorf("launch server: %w", err)
			return
		}
	}()

	select {
	case done := <-ctx.Done():
		return fmt.Errorf("ctx.Done: %v", done)
	case err := <-srvErrCh:
		return fmt.Errorf("server error: %w", err)
	}
}
