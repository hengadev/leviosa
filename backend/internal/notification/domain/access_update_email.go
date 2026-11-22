package domain

import (
	"fmt"
	"time"
)

// BookingAccessUpdateRequest tells a client with an upcoming confirmed
// booking that the practice access code changed, with the new access. Sent
// when the administrator changes the door code, or sends the current access
// again. Written in the locale the client booked in. No health data
// (ADR-0019).
type BookingAccessUpdateRequest struct {
	ToEmail     string
	ToFirstName string
	ToLastName  string
	Locale      Locale

	BookingID   string
	ProductName string
	PartnerName string
	// Address is the practice address.
	Address string

	StartTime time.Time
	EndTime   time.Time

	// ManageURL opens the booking token page. Empty when the booking has no
	// token.
	ManageURL string

	// DoorCode and AccessInstructions are the practice access, read when the
	// email is sent. Private: never logged.
	DoorCode           string
	AccessInstructions string
}

// BookingAccessUpdateTemplate is the template of the access update email
// (templates/booking_access_update.html).
const BookingAccessUpdateTemplate = "booking_access_update"

// accessUpdateStrings is the copy of the access update email.
type accessUpdateStrings struct {
	Subject           string // fmt: date
	Title             string
	Greeting          string
	Intro             string
	DetailsHeading    string
	ServiceLabel      string
	DateLabel         string
	TimeLabel         string
	PractitionerLabel string
	AddressLabel      string
	ReferenceLabel    string
	AccessHeading     string
	DoorCodeLabel     string
	InstructionsLabel string
	ManageText        string
	ManageButton      string
	SignOff           string
	Team              string
	Rights            string
}

var accessUpdateCopy = map[Locale]accessUpdateStrings{
	LocaleFR: {
		Subject:           "Nouveau code d'accès pour votre rendez-vous du %s",
		Title:             "Nouveau code d'accès",
		Greeting:          "Bonjour",
		Intro:             "Le code d'accès du cabinet a changé. Voici le nouveau code pour votre prochain rendez-vous.",
		DetailsHeading:    "Votre rendez-vous",
		ServiceLabel:      "Prestation",
		DateLabel:         "Date",
		TimeLabel:         "Heure",
		PractitionerLabel: "Avec",
		AddressLabel:      "Adresse",
		ReferenceLabel:    "Référence",
		AccessHeading:     "Accès au cabinet",
		DoorCodeLabel:     "Code d'accès",
		InstructionsLabel: "Indications",
		ManageText:        "Un empêchement ? Vous pouvez consulter ou annuler votre réservation depuis ce lien :",
		ManageButton:      "Gérer ma réservation",
		SignOff:           "À très bientôt,",
		Team:              "L'équipe Leviosa",
		Rights:            "Tous droits réservés.",
	},
	LocaleEN: {
		Subject:           "New access code for your appointment on %s",
		Title:             "New access code",
		Greeting:          "Hello",
		Intro:             "The practice access code has changed. Here is the new code for your upcoming appointment.",
		DetailsHeading:    "Your appointment",
		ServiceLabel:      "Treatment",
		DateLabel:         "Date",
		TimeLabel:         "Time",
		PractitionerLabel: "With",
		AddressLabel:      "Address",
		ReferenceLabel:    "Reference",
		AccessHeading:     "Getting in",
		DoorCodeLabel:     "Door code",
		InstructionsLabel: "Directions",
		ManageText:        "Can't make it? You can view or cancel your booking from this link:",
		ManageButton:      "Manage my booking",
		SignOff:           "See you soon,",
		Team:              "The Leviosa team",
		Rights:            "All rights reserved.",
	},
}

// BookingAccessUpdateView is the data of the booking_access_update template.
type BookingAccessUpdateView struct {
	Lang        string
	T           accessUpdateStrings
	FirstName   string
	ProductName string
	PartnerName string
	Address     string
	Date        string
	TimeRange   string
	ManageURL   string
	BookingID   string
	Year        int
	PracticeAccessView
}

// ComposeBookingAccessUpdate builds the access update email in the request's
// locale, dates and times in the practice's time zone. No attachment. now
// stamps the footer year.
func ComposeBookingAccessUpdate(req BookingAccessUpdateRequest, now time.Time) ComposedEmail {
	locale := ParseLocale(string(req.Locale))
	strs := accessUpdateCopy[locale]
	date := FormatDate(req.StartTime, locale)

	view := BookingAccessUpdateView{
		Lang:        string(locale),
		T:           strs,
		FirstName:   req.ToFirstName,
		ProductName: req.ProductName,
		PartnerName: req.PartnerName,
		Address:     req.Address,
		Date:        date,
		TimeRange:   FormatTimeRange(req.StartTime, req.EndTime, locale),
		ManageURL:   req.ManageURL,
		BookingID:   req.BookingID,
		Year:        now.In(practiceLocation).Year(),

		PracticeAccessView: newPracticeAccessView(req.DoorCode, req.AccessInstructions),
	}

	return ComposedEmail{
		Subject:  fmt.Sprintf(strs.Subject, date),
		Template: BookingAccessUpdateTemplate,
		Data:     view,
	}
}
