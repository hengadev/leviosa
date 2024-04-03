package types

import (
	"github.com/google/uuid"
)

// TODO: Add a field to know if the person paid or not within the time period acceptable
// TODO: Add test for the ispaid voting thing

type Vote struct {
	Id      string `json:"id"`
	UserId  string `json:"userid"`
	EventId string `json:"eventid"`
	// IsPaid  bool   `json:"ispaid"`
}

// NOTE: shoud I use *string or string ?
func NewVote(userid, eventid *string) *Vote {
	return &Vote{
		Id:      uuid.NewString(),
		UserId:  *userid,
		EventId: *eventid,
		// IsPaid  false
	}
}
