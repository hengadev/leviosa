package voteRepository_test

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
)

var (
	baseVote = &vote.Vote{
		UserID: 1,
		Day:    23,
		Month:  4,
		Year:   2025,
	}
	baseVote2 = &vote.Vote{
		UserID: 1,
		Day:    12,
		Month:  4,
		Year:   2025,
	}
	baseVote3 = &vote.Vote{
		UserID: 1,
		Day:    6,
		Month:  4,
		Year:   2025,
	}
)
