package cron

import (
	"fmt"
	// needs to use the database or the api thing to make some request
	// "github.com/GaryHY/event-reservation-app/internal/database/sqlite"
	"github.com/robfig/cron"
	"time"
)

// A function to just set the cron job friend
func SetCron() {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		fmt.Println("Failed to load the location with the time module - ", err)
	}
	c := cron.NewWithLocation(loc)
	defer c.Stop()
	c.AddFunc("0 0 8 * * *", parseEvent)
	c.AddFunc("0 0 8 * * *", checkCloseVote)
	c.AddFunc("0 0 8 * * *", backupDatabase)
	c.Start()
	select {}
}

// TODO: Fonctions a implementer, il me faut de la database du mail etc..

// Une fonction pour realiser des actions a l'approche de certaines dates.
func parseEvent() {
	// TODO:
	// 1. get the list of events
	// 2. send the emails with the right templates
	fmt.Println("Parse ther event man !.")
}

// Une fonction pour realiser des actions des qu'un vote est termine
func checkCloseVote() {
	fmt.Println("Checking if I need to close the votes.")
}

// Une fonction pour backer mon fichier sqlite ailleurs
func backupDatabase() {
	fmt.Println("Ici on bien dormi en fait.")
}

// TODO: The CRON jobs that I need on the server.
// - A CRON job to fetch the database daily to see if the limit date for a vote is reached.
// - The cron jobs to send messages to the users for reminders on the events.
