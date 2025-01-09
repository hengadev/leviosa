package eventRepository_test

import (
	"time"

	"github.com/GaryHY/leviosa/internal/domain/event"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/event"
)

var (
	beginAt, _  = eventRepository.ExportedParseBeginAt("08:00:00", 12, 7, 1998)
	beginAt1, _ = eventRepository.ExportedParseBeginAt("08:00:00", 13, 7, 1998)
	beginAt2, _ = eventRepository.ExportedParseBeginAt("08:00:00", 14, 7, 1998)

	baseEvent = &eventService.Event{
		ID:              "ea1d74e2-1612-47ec-aee9-c6a46b65640f",
		Location:        "Impasse Inconnue",
		PlaceCount:      16,
		FreePlace:       14,
		BeginAt:         beginAt,
		SessionDuration: time.Minute * 30,
		Day:             12,
		Month:           7,
		Year:            1998,
	}

	baseEvent1 = &eventService.Event{
		ID:              "b16a6f38-d2fb-428c-b97c-929b1010b951",
		Location:        "Impasse Inconnue",
		PlaceCount:      23,
		FreePlace:       19,
		BeginAt:         beginAt1,
		SessionDuration: time.Minute * 30,
		Day:             13,
		Month:           7,
		Year:            1998,
	}

	baseEvent2 = &eventService.Event{
		ID:              "9a676c5d-c9ec-4266-a426-24e5d4983caf",
		Location:        "Impasse Inconnue",
		PlaceCount:      22,
		FreePlace:       0,
		BeginAt:         beginAt2,
		SessionDuration: time.Minute * 30,
		Day:             14,
		Month:           7,
		Year:            1998,
	}

	baseEventWithPriceID = &eventService.Event{
		ID:              "ea1d74e2-1612-47ec-aee9-c6a46b65640f",
		Location:        "Impasse Inconnue",
		PlaceCount:      16,
		FreePlace:       14,
		BeginAt:         beginAt,
		SessionDuration: time.Minute * 30,
		PriceID:         "4fe0vuw3ef0223",
		Day:             12,
		Month:           7,
		Year:            1998,
	}
)
