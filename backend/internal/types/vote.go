package types

// TODO: Add test for the ispaid voting thing

const (
	VoteSeparator = "|"
)

// NOTE: Do I need an Id for that thing ?
type Vote struct {
	UserId string `json:"userid"`
	Month  int    `json:"month"`
	Year   int    `json:"year"`
	Day    int    `json:"day"`
}

type VoteSent struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

type VoteStored struct {
	Days  string `json:"day"`
	Month int    `json:"month"`
	Year  int    `json:"year"`
}

type NextVote struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}

// NOTE: remove that function since I do not use it anymore
func NewVote(userid, eventid *string) *Vote {
	return &Vote{
		UserId: *userid,
		Month:  0,
		Year:   0,
		Day:    0,
	}
}
