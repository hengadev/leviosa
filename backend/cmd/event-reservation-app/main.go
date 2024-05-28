package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/GaryHY/event-reservation-app/internal/api"
	"github.com/GaryHY/event-reservation-app/internal/database/s3"
	"github.com/GaryHY/event-reservation-app/internal/database/sqlite"

	// cron job import
	"github.com/GaryHY/event-reservation-app/internal/cron"
	"github.com/joho/godotenv"
	// "github.com/robfig/cron"
	"log"
	"net/http"
)

// TODO: Use that function when done with the app, and find a way to stop it using CTRL + C.
// get the function run that I am going to use in main from this link : https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("Failed to load the environment variables - %w", err)
	}
	// create the store and server
	store, err := sqlite.NewStore("db.sqlite")
	photostore := s3.NewPhotoStore()
	if err != nil {
		return fmt.Errorf("Cannot connect to the database - %w", err)
	}
	server := api.NewServer(store, photostore)

	// TODO: Make the crom function work
	// c := cron.New()
	// defer c.Stop()
	// c.AddFunc("* * * * * *", somefunc.SomefunctionForTHeCron)
	// c.Start()
	// select {}

	fmt.Println("Running server on port 5000...")
	if err := http.ListenAndServe(":5000", server); err != nil {
		return fmt.Errorf("Cannot launch the server - %w", err)
	}
	return nil
}
func main() {
	// TODO: Use that function when done with the app, and find a way to stop it using CTRL + C.
	// ctx := context.Background()
	// if err := run(ctx, os.Stdout, os.Args); err != nil {
	// 	fmt.Fprintf(os.Stderr, "%s\n", err)
	// 	os.Exit(1)
	// }

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load the environment variables - ", err)
	}
	// create the store and server
	store, err := sqlite.NewStore("db.sqlite")
	photostore := s3.NewPhotoStore()
	if err != nil {
		log.Fatal("Cannot connect to the database - ", err)
	}
	server := api.NewServer(store, photostore)

	cron.SetCron() // setting the cron job

	fmt.Println("Running server on port 5000...")
	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatal("Cannot launch the server - ", err)
	}
}
