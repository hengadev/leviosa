package factories

import (
	"github.com/GaryHY/leviosa/internal/domain/event/models"
)

func NewBasicEvent(overrides map[string]any) *models.Event {
	event := &models.Event{
		ID:               "ea1d74e2-1612-47ec-aee9-c6a46b65640f",
		Title:            "First event for Leviosa",
		Description:      "First description for Leviosa",
		City:             "Paris",
		PostalCode:       "postalCode",
		Address1:         "address1",
		Address2:         "",
		PlaceCount:       16,
		FreePlace:        14,
		EncryptedBeginAt: "08:00:00",
		EncryptedEndAt:   "20:00:00",
		PriceID:          "179cf8f1-81ad-4ec1-b8bb-8f48abf9ef80",
		Day:              22,
		Month:            4,
		Year:             2025,
	}
	for key, value := range overrides {
		switch key {
		case "ID":
			event.ID = value.(string)
		}
		// TODO: complete other cases
	}
	return event
}

func NewBasicEventList() []*models.Event {
	events := []*models.Event{
		NewBasicEvent(nil),
		NewBasicEvent(map[string]any{
			"ID":               "43391431-984f-4b06-8fcc-88040215430b",
			"Title":            "Second event for Leviosa",
			"Description":      "Second description for Leviosa",
			"City":             "Marseille",
			"PostalCode":       "postalCode2",
			"Address1":         "address1 - 2",
			"Address2":         "",
			"PlaceCount":       6,
			"FreePlace":        3,
			"EncryptedBeginAt": "09:00:00",
			"EncryptedEndAt":   "20:00:00",
			"PriceID":          "bdab8511-875a-46d5-a228-6db7aecb42a2",
			"Day":              17,
			"Month":            5,
			"Year":             2025,
		}),
		NewBasicEvent(map[string]any{
			"ID":               "9a676c5d-c9ec-4266-a426-24e5d4983caf",
			"Title":            "Third event for Leviosa",
			"Description":      "Third description for Leviosa",
			"City":             "Lyon",
			"PostalCode":       "postalCode3",
			"Address1":         "address1 - 3",
			"Address2":         "",
			"PlaceCount":       18,
			"FreePlace":        4,
			"EncryptedBeginAt": "10:00:00",
			"EncryptedEndAt":   "21:00:00",
			"PriceID":          "ef55b80d-6eb2-4e22-9b68-ea219c202c71",
			"Day":              3,
			"Month":            6,
			"Year":             2025,
		}),
	}

	return events
}
