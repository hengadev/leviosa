package throttlerService

import (
	"time"
)

const THROTTLERSESSIONDURATION = 1 * time.Hour
const THROTTLINGDURATION = 15 * time.Minute

// NOTE: here we should try something different for the throttling
var testduration = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
	15 * time.Second,
	30 * time.Second,
	60 * time.Second,
	120 * time.Second,
	300 * time.Second,
}

var MAXATTEMPT = len(testduration)

// TODO: need to hash the email in redis
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
