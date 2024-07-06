package vote

import "github.com/google/uuid"

const VoteSeparator = "|"

type Vote struct {
	UserID string `json:"userid,omitempty"`
	Day    int    `json:"day,omitempty"`
	Month  int    `json:"month"`
	Year   int    `json:"year"`
}

func NewVote(day, month, year int) *Vote {
	return &Vote{
		Day:   day,
		Month: month,
		Year:  year,
	}
}

func (v *Vote) Create() {
	v.UserID = uuid.NewString()
}

// NOTE: the way I am going to make the tables
// vote : month year days
// vote_august_2024 : userID days
