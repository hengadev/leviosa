package main

import (
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/api"
	"github.com/GaryHY/event-reservation-app/internal/database"
	// "github.com/GaryHY/event-reservation-app/internal/somefunc"
	"github.com/joho/godotenv"
	// "github.com/robfig/cron"
	"log"
	"net/http"
)

func main() {
	// load the env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	// create the store and server
	store, err := sqlite.NewStore("db.sqlite")
	photostore := sqlite.NewPhotoStore()
	if err != nil {
		log.Fatal("Cannot connect to the database")
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
		log.Fatal("Cannot launch the server - ", err)
	}

}
