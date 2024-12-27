package vote_test

import (
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
)

// setup provides given a time.Time instance a month and year for the next valid votes. That function helps with the case where you want the next valid vote in December of some year for example.
func setup(now time.Time) (int, int) {
	var year int
	var month int
	if now.Month() == 12 {
		month = 1
		year = now.Year() + 1
	} else {
		month = int(now.Month()) + 1
		year = now.Year()
	}
	return month, year
}

// generate  a valid vote given a valid month and year with the day set to the day of the time.Time passed, use a random userID
func generateValidVote(now time.Time, month, year int) (string, *vote.Vote) {
	userID := "fwrg98wo2n3fh4wt"
	return userID, &vote.Vote{
		UserID: userID,
		Day:    now.Day(),
		Month:  month,
		Year:   year,
	}
}

// Get the formatted days as in the database given an array of int representing days.
func getFormattedDayFromIntArr(days []int) string {
	var formattedDays string
	for _, day := range days {
		formattedDays += fmt.Sprintf("%d%s", day, vote.VoteSeparator)
	}
	return formattedDays[:len(formattedDays)-1]
}

// Get the votes given an array of int representing days, userID, month and year.
func getVotesFromIntDaysArr(userID string, days []int, month, year int) []*vote.Vote {
	var votes []*vote.Vote
	for _, day := range days {
		votes = append(votes, &vote.Vote{
			UserID: userID,
			Day:    day,
			Month:  month,
			Year:   year,
		})
	}
	return votes
}
