package cron

import (
	"context"
	"fmt"
)

// Une fonction pour realiser des actions des qu'un vote est termine
func checkCloseVote(ctx context.Context) func() error {
	return func() error {
		fmt.Println("Checking if I need to close the votes.")
		return nil
	}
}
