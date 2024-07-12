package main

import (
	"context"
	"flag"
	"fmt"

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

	"log"
)

var opts struct {
	mode   string
	server struct {
		port int
	}
}

func main() {
	// TODO: Use that function when done with the app, and find a way to stop it using CTRL + C.
	// ctx := context.Background()
	// if err := run(ctx, os.Stdout, os.Args); err != nil {
	// 	fmt.Fprintf(os.Stderr, "%s\n", err)
	// 	os.Exit(1)
	// }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// use flags to do configuration for some part of code
	flag.IntVar(&opts.server.port, "port", 5000, "the port the server listens to")
	flag.StringVar(&opts.mode, "mode", "dev", "the mode environment for the project")
	flag.Parse()

	// the version classic without sauce
	// err := env.Load()
	envFile := fmt.Sprintf(".env.%s", opts.mode)
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatal("Failed to load the environment variables - ", err)
	}
	// setting the cron job
	// c.SetCron()

	// config
	conf := config.New(ctx)
	if err := conf.Load(ctx); err != nil {
		log.Fatal("failed to load configuration:", err)
	}
	sqliteConf := conf.GetSQLITE()

	sqlitedb, err := sqliteutil.Connect(ctx, sqliteutil.BuildDSN(sqliteConf.Filename))
	if err != nil {
		log.Fatal("failed to create connection to sqlite :", err)
	}

	// user
	userRepo := sqlite.NewUserRepository(ctx, sqlitedb)
	if err != nil {
		log.Fatal("user repo err: ", err)
	}
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
	appRepos := handler.Repos{
		User: userRepo,
		// Session: sessionRepo,
	}
	handler := handler.NewHandler(&appSvcs, &appRepos)
	srv := server.New(handler, server.WithPort(opts.server.port))

	fmt.Printf("Running server on port %d...\n", opts.server.port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Cannot launch the server - ", err)
	}
}

// TODO: Use that function when done with the app, and find a way to stop it using CTRL + C.
// get the function run that I am going to use in main from this link : https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
// func run(ctx context.Context, w io.Writer, args []string) error {
// 	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
// 	defer cancel()
// 	err := env.Load()
// 	if err != nil {
// 		return fmt.Errorf("Failed to load the environment variables - %w", err)
// 	}
// 	// create the instances of database that I am going to use with their configuration
//
// 	// create the stores and server brother
// 	store, err := sqlite.NewStore("db.sqlite")
// 	photostore := s3.NewPhotoStore()
// 	if err != nil {
// 		return fmt.Errorf("Cannot connect to the database - %w", err)
// 	}
// 	server := api.NewServer(store, photostore)
//
// 	fmt.Println("Running server on port 5000...")
// 	if err := http.ListenAndServe(":5000", server); err != nil {
// 		return fmt.Errorf("Cannot launch the server - %w", err)
// 	}
// 	return nil
// }
