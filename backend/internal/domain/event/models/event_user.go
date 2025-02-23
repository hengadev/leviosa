package models

type EventUser struct {
	// PastEvents     []*Event `json:"pastEvents"`
	// NextEvents     []*Event `json:"nextEvents"`
	// IncomingEvents []*Event `json:"incomingEvents"`

	PastEvents     []*Event `json:"past_events"`
	NextEvents     []*Event `json:"next_events"`
	IncomingEvents []*Event `json:"incoming_events"`
}
