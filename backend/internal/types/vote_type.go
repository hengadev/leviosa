package types

import (
	"github.com/google/uuid"
)

type Vote struct {
	Id      string `json:"id"`
	Userid  string `json:"userid"`
	EventId string `json:"eventid"`
}

func NewVote(userid, eventid *string) *Vote {
	return &Vote{
		Id:      uuid.NewString(),
		Userid:  *userid,
		EventId: *eventid,
	}
}
