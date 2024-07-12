package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	// "github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/server"
	"github.com/GaryHY/event-reservation-app/pkg/config"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"

	// "github.com/GaryHY/event-reservation-app/internal/redis"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"

	// "github.com/GaryHY/event-reservation-app/internal/http/cron"
	handler "github.com/GaryHY/event-reservation-app/internal/server/service"
	"github.com/joho/godotenv"
)

var opts struct {
	mode   string
	server struct {
		port int
	}
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	// flags
	flag.IntVar(&opts.server.port, "port", 5000, "the port the server listens to")
	flag.StringVar(&opts.mode, "mode", "dev", "the mode environment for the project")
	flag.Parse()

	// set environment file
	err := godotenv.Load(fmt.Sprintf("%s.env", opts.mode))
	if err != nil {
		return fmt.Errorf("load env variables: %w", err)
	}

	// setting cron jobs
	// c.SetCron()

	// config
	conf := config.New(ctx, opts.mode, "env")
	if err := conf.Load(ctx); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	sqliteConf := conf.GetSQLITE()

	// datanases setup
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
	// 	log.Fatal("session repo err: ", err)
	// }
	// sessionSvc := session.NewService(sessionRepo)

	// services
	appSvcs := handler.Services{
		User: userSvc,
		// Session: sessionSvc,
	}
	// repos
	appRepos := handler.Repos{
		User: userRepo,
		// Session: sessionRepo,
	}
	handler := handler.NewHandler(&appSvcs, &appRepos)
	srv := server.New(handler, server.WithPort(opts.server.port))

	var srvErrCh = make(chan error)
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
