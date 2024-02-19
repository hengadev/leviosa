package types

import (
	"github.com/google/uuid"
)

const (
	EventFormat = "2006-01-02"
)


type Event struct {
	Id         string `json:"id"`
	Location   string `json:"location"`
	PlaceCount int    `json:"placecount"`
	Date       string `json:"date"`
}
}
