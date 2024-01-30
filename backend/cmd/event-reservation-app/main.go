package main

import (
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/api"
	"net/http"
)

func main() {
	fmt.Println("Running server on port 5000...")
	http.ListenAndServe(":5000", http.HandlerFunc(api.PlayerServer))
}
