package throttlerService

import (
	"time"
)

const MAXATTEMPT = 5
const THROTTLERSESSIONDURATION = 1 * time.Hour
const THROTTLINGDURATION = 15 * time.Minute

type Info struct {
	Email       string
	Attempts    int
	LastAttempt time.Time
	LockedUntil time.Time
}

func NewInfo(email string) *Info {
	return &Info{
		Email:       email,
		Attempts:    0,
		LastAttempt: time.Time{},
		LockedUntil: time.Time{},
	}
}
