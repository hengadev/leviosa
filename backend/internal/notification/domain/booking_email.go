package domain

import (
	"fmt"
	"strings"
	"time"
)

// Booking emails (single-practitioner booking MVP). They carry no health
// data (ADR-0019): no notes, no free text from the client, only the
// appointment, its price and contact details.

// BookingConfirmationRequest is the client confirmation of a confirmed
// booking, written in the locale the client booked in.
type BookingConfirmationRequest struct {
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

	// AmountPaidOnlineCents is what the client paid online (the deposit, or
	// the full price); BalanceDueCents what remains to pay at the practice.
	AmountPaidOnlineCents int
	BalanceDueCents       int
	Currency              string

	// ManageURL opens the booking token page, where the client can see and
	// cancel the booking. Empty when the booking has no token.
	ManageURL string

	// DoorCode and AccessInstructions are the practice access, read when the
	// email is sent. Private: never in the calendar file, never logged.
	DoorCode           string
	AccessInstructions string
}

// BookingPartnerNoticeRequest tells the partner a new booking was confirmed.
// Partner emails are French.
type BookingPartnerNoticeRequest struct {
	ToEmail     string
	ToFirstName string

	BookingID   string
	ProductName string

	ClientName  string
	ClientEmail string
	ClientPhone string

	StartTime time.Time
	EndTime   time.Time
}

// ComposedEmail is an email ready for a transport: subject, template and its
// data, attachments.
type ComposedEmail struct {
	Subject     string
	Template    string
	Data        any
	Attachments []EmailAttachment
}

// Template names of the booking emails (templates/<name>.html).
const (
	BookingConfirmationTemplate  = "booking_confirmation"
	BookingPartnerNoticeTemplate = "booking_partner_notice"
)

// clientConfirmationStrings is the copy of the client confirmation email.
type clientConfirmationStrings struct {
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
	PaidOnlineLabel   string
	BalanceLabel      string
	ReferenceLabel    string
	AccessHeading     string
	DoorCodeLabel     string
	InstructionsLabel string
	ManageText        string
	ManageButton      string
	CalendarNote      string
	SignOff           string
	Team              string
	Rights            string
	ICSDescription    string // fmt: manage URL
	ICSFilename       string
}

var clientConfirmationCopy = map[Locale]clientConfirmationStrings{
	LocaleFR: {
		Subject:           "Votre rendez-vous du %s est confirmé",
		Title:             "Rendez-vous confirmé",
		Greeting:          "Bonjour",
		Intro:             "Votre réservation est confirmée. Voici le récapitulatif de votre rendez-vous.",
		DetailsHeading:    "Votre rendez-vous",
		ServiceLabel:      "Prestation",
		DateLabel:         "Date",
		TimeLabel:         "Heure",
		PractitionerLabel: "Avec",
		AddressLabel:      "Adresse",
		PaidOnlineLabel:   "Payé en ligne",
		BalanceLabel:      "Reste à régler sur place, par carte",
		ReferenceLabel:    "Référence",
		AccessHeading:     "Accès au cabinet",
		DoorCodeLabel:     "Code d'accès",
		InstructionsLabel: "Indications",
		ManageText:        "Vous pouvez consulter ou annuler votre réservation à tout moment depuis ce lien :",
		ManageButton:      "Gérer ma réservation",
		CalendarNote:      "Le fichier joint (.ics) ajoute ce rendez-vous à votre agenda.",
		SignOff:           "À très bientôt,",
		Team:              "L'équipe Leviosa",
		Rights:            "Tous droits réservés.",
		ICSDescription:    "Gérer ou annuler votre réservation : %s",
		ICSFilename:       "rendez-vous.ics",
	},
	LocaleEN: {
		Subject:           "Your appointment on %s is confirmed",
		Title:             "Appointment confirmed",
		Greeting:          "Hello",
		Intro:             "Your booking is confirmed. Here is a summary of your appointment.",
		DetailsHeading:    "Your appointment",
		ServiceLabel:      "Treatment",
		DateLabel:         "Date",
		TimeLabel:         "Time",
		PractitionerLabel: "With",
		AddressLabel:      "Address",
		PaidOnlineLabel:   "Paid online",
		BalanceLabel:      "Balance due by card at the practice",
		ReferenceLabel:    "Reference",
		AccessHeading:     "Getting in",
		DoorCodeLabel:     "Door code",
		InstructionsLabel: "Directions",
		ManageText:        "You can view or cancel your booking at any time from this link:",
		ManageButton:      "Manage my booking",
		CalendarNote:      "The attached file (.ics) adds this appointment to your calendar.",
		SignOff:           "See you soon,",
		Team:              "The Leviosa team",
		Rights:            "All rights reserved.",
		ICSDescription:    "Manage or cancel your booking: %s",
		ICSFilename:       "appointment.ics",
	},
}

// BookingConfirmationView is the data of the booking_confirmation template.
type BookingConfirmationView struct {
	Lang        string
	T           clientConfirmationStrings
	FirstName   string
	ProductName string
	PartnerName string
	Address     string
	Date        string
	TimeRange   string
	AmountPaid  string
	// BalanceDue is empty when nothing remains to pay at the practice.
	BalanceDue string
	ManageURL  string
	BookingID  string
	Year       int
	PracticeAccessView
}

// PracticeAccessView is the practice access block of the client confirmation
// and reminder: shown only when HasAccess, i.e. a door code or access
// instructions are set.
type PracticeAccessView struct {
	HasAccess bool
	DoorCode  string
	// AccessInstructions are the instruction lines, rendered with line breaks
	// between them and HTML-escaped like any template value.
	AccessInstructions []string
}

// newPracticeAccessView builds the access block from the door code and the
// access instructions.
func newPracticeAccessView(doorCode, instructions string) PracticeAccessView {
	view := PracticeAccessView{DoorCode: strings.TrimSpace(doorCode)}
	instructions = strings.TrimSpace(strings.ReplaceAll(instructions, "\r\n", "\n"))
	if instructions != "" {
		for _, line := range strings.Split(instructions, "\n") {
			view.AccessInstructions = append(view.AccessInstructions, strings.TrimRight(line, " \t\r"))
		}
	}
	view.HasAccess = view.DoorCode != "" || len(view.AccessInstructions) > 0
	return view
}

// BookingCalendarUID is the stable iCalendar UID of a booking.
func BookingCalendarUID(bookingID string) string {
	return "booking-" + bookingID + "@leviosa"
}

// ComposeBookingConfirmation builds the client confirmation email in the
// request's locale, with the appointment as an .ics attachment. now stamps
// the calendar file and the footer year.
func ComposeBookingConfirmation(req BookingConfirmationRequest, now time.Time) ComposedEmail {
	locale := ParseLocale(string(req.Locale))
	strs := clientConfirmationCopy[locale]
	date := FormatDate(req.StartTime, locale)

	view := BookingConfirmationView{
		Lang:        string(locale),
		T:           strs,
		FirstName:   req.ToFirstName,
		ProductName: req.ProductName,
		PartnerName: req.PartnerName,
		Address:     req.Address,
		Date:        date,
		TimeRange:   FormatTimeRange(req.StartTime, req.EndTime, locale),
		AmountPaid:  FormatMoney(req.AmountPaidOnlineCents, req.Currency, locale),
		ManageURL:   req.ManageURL,
		BookingID:   req.BookingID,
		Year:        now.In(practiceLocation).Year(),

		PracticeAccessView: newPracticeAccessView(req.DoorCode, req.AccessInstructions),
	}
	if req.BalanceDueCents > 0 {
		view.BalanceDue = FormatMoney(req.BalanceDueCents, req.Currency, locale)
	}

	summary := req.ProductName
	if req.PartnerName != "" {
		summary = fmt.Sprintf("%s – %s", req.ProductName, req.PartnerName)
	}
	// The calendar file carries the address and the manage link only, never
	// the door code or the access instructions.
	event := CalendarEvent{
		UID:      BookingCalendarUID(req.BookingID),
		Start:    req.StartTime,
		End:      req.EndTime,
		Summary:  summary,
		Location: req.Address,
		URL:      req.ManageURL,
		Stamp:    now,
	}
	if req.ManageURL != "" {
		event.Description = fmt.Sprintf(strs.ICSDescription, req.ManageURL)
	}

	return ComposedEmail{
		Subject:  fmt.Sprintf(strs.Subject, date),
		Template: BookingConfirmationTemplate,
		Data:     view,
		Attachments: []EmailAttachment{{
			Filename:    strs.ICSFilename,
			ContentType: "text/calendar; charset=utf-8; method=PUBLISH",
			Data:        event.ICS(),
		}},
	}
}

// BookingPartnerNoticeView is the data of the booking_partner_notice
// template.
type BookingPartnerNoticeView struct {
	FirstName   string
	ProductName string
	ClientName  string
	ClientEmail string
	ClientPhone string
	Date        string
	TimeRange   string
	BookingID   string
	Year        int
}

// ComposeBookingPartnerNotice builds the French new-booking notice for the
// partner.
func ComposeBookingPartnerNotice(req BookingPartnerNoticeRequest, now time.Time) ComposedEmail {
	date := FormatDate(req.StartTime, LocaleFR)
	view := BookingPartnerNoticeView{
		FirstName:   req.ToFirstName,
		ProductName: req.ProductName,
		ClientName:  req.ClientName,
		ClientEmail: req.ClientEmail,
		ClientPhone: req.ClientPhone,
		Date:        date,
		TimeRange:   FormatTimeRange(req.StartTime, req.EndTime, LocaleFR),
		BookingID:   req.BookingID,
		Year:        now.In(practiceLocation).Year(),
	}
	return ComposedEmail{
		Subject:  fmt.Sprintf("Nouvelle réservation : %s, %s à %s", req.ProductName, date, FormatTime(req.StartTime, LocaleFR)),
		Template: BookingPartnerNoticeTemplate,
		Data:     view,
	}
}

// BookingReminderTemplate is the template of the client reminder
// (templates/booking_reminder.html).
const BookingReminderTemplate = "booking_reminder"

// clientReminderStrings is the copy of the client reminder email.
type clientReminderStrings struct {
	Subject           string // fmt: date, time
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

var clientReminderCopy = map[Locale]clientReminderStrings{
	LocaleFR: {
		Subject:           "Rappel : votre rendez-vous du %s à %s",
		Title:             "Votre rendez-vous approche",
		Greeting:          "Bonjour",
		Intro:             "Petit rappel de votre rendez-vous à venir. Nous avons hâte de vous accueillir.",
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
		Subject:           "Reminder: your appointment on %s at %s",
		Title:             "Your appointment is coming up",
		Greeting:          "Hello",
		Intro:             "A quick reminder of your upcoming appointment. We look forward to seeing you.",
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

// BookingReminderView is the data of the booking_reminder template.
type BookingReminderView struct {
	Lang        string
	T           clientReminderStrings
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

// ComposeBookingReminder builds the client reminder email in the request's
// locale, dates and times in the practice's time zone. now stamps the footer
// year.
func ComposeBookingReminder(req BookingReminderRequest, now time.Time) ComposedEmail {
	locale := ParseLocale(string(req.Locale))
	strs := clientReminderCopy[locale]
	date := FormatDate(req.StartTime, locale)

	view := BookingReminderView{
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
		Subject:  fmt.Sprintf(strs.Subject, date, FormatTime(req.StartTime, locale)),
		Template: BookingReminderTemplate,
		Data:     view,
	}
}
