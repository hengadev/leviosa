package main

import (
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/api"
	"github.com/GaryHY/event-reservation-app/internal/database"
	"log"
	"net/http"
)

func main() {
	store, err := sqlite.NewStore("db.sqlite")
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}
	server := api.NewServer(store)

	fmt.Println("Running server on port 5000...")
	http.ListenAndServe(":5000", server)
}
