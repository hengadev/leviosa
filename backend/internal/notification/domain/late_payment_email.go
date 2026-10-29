package domain

import (
	"fmt"
	"time"
)

// LatePaymentRefundRequest tells a client their payment was refunded in
// full: it arrived after the payment hold elapsed and the slot had been
// booked by someone else meanwhile, so the booking could not be confirmed
// (ADR-0018). Written in the locale the client booked in. No health data
// (ADR-0019).
type LatePaymentRefundRequest struct {
	ToEmail     string
	ToFirstName string
	ToLastName  string
	Locale      Locale

	BookingID   string
	ProductName string

	StartTime time.Time
	EndTime   time.Time

	RefundedAmountCents int
	Currency            string

	// BookAgainURL opens the practice page in the client's language, to pick
	// another slot. Empty when the public site's origin is unknown.
	BookAgainURL string
}

// LatePaymentRefundTemplate is the template of the late-payment refund email.
const LatePaymentRefundTemplate = "booking_late_payment_refund"

// latePaymentRefundStrings is the copy of the late-payment refund email.
type latePaymentRefundStrings struct {
	Subject        string
	Title          string
	Greeting       string
	Intro          string // fmt: date, time range
	RefundText     string // fmt: amount
	DelayNote      string
	DetailsHeading string
	ServiceLabel   string
	DateLabel      string
	TimeLabel      string
	RefundLabel    string
	ReferenceLabel string
	BookAgainText  string
	BookAgainLink  string
	SignOff        string
	Team           string
	Rights         string
}

var latePaymentRefundCopy = map[Locale]latePaymentRefundStrings{
	LocaleFR: {
		Subject:        "Votre paiement a été remboursé",
		Title:          "Paiement remboursé",
		Greeting:       "Bonjour",
		Intro:          "Votre paiement pour le rendez-vous du %s (%s) nous est parvenu après la fin du délai de réservation, et ce créneau n'était plus disponible. Nous n'avons donc pas pu confirmer ce rendez-vous.",
		RefundText:     "Votre paiement de %s vous a été intégralement remboursé.",
		DelayNote:      "Le remboursement apparaît sur votre relevé sous 5 à 10 jours ouvrés, selon votre banque.",
		DetailsHeading: "Créneau demandé",
		ServiceLabel:   "Prestation",
		DateLabel:      "Date",
		TimeLabel:      "Heure",
		RefundLabel:    "Montant remboursé",
		ReferenceLabel: "Référence",
		BookAgainText:  "Vous pouvez choisir un autre créneau à tout moment :",
		BookAgainLink:  "Choisir un autre créneau",
		SignOff:        "Avec toutes nos excuses,",
		Team:           "L'équipe Leviosa",
		Rights:         "Tous droits réservés.",
	},
	LocaleEN: {
		Subject:        "Your payment has been refunded",
		Title:          "Payment refunded",
		Greeting:       "Hello",
		Intro:          "Your payment for the appointment on %s (%s) reached us after the booking time limit had passed, and the slot was no longer available. We could therefore not confirm this appointment.",
		RefundText:     "Your payment of %s has been refunded in full.",
		DelayNote:      "The refund shows on your statement within 5 to 10 business days, depending on your bank.",
		DetailsHeading: "Requested slot",
		ServiceLabel:   "Treatment",
		DateLabel:      "Date",
		TimeLabel:      "Time",
		RefundLabel:    "Amount refunded",
		ReferenceLabel: "Reference",
		BookAgainText:  "You can pick another slot at any time:",
		BookAgainLink:  "Pick another slot",
		SignOff:        "With our apologies,",
		Team:           "The Leviosa team",
		Rights:         "All rights reserved.",
	},
}

// LatePaymentRefundView is the data of the booking_late_payment_refund
// template.
type LatePaymentRefundView struct {
	Lang         string
	T            latePaymentRefundStrings
	FirstName    string
	ProductName  string
	Date         string
	TimeRange    string
	Intro        string
	RefundText   string
	Amount       string
	BookAgainURL string
	BookingID    string
	Year         int
}

// ComposeLatePaymentRefund builds the late-payment refund email in the
// request's locale. now stamps the footer year.
func ComposeLatePaymentRefund(req LatePaymentRefundRequest, now time.Time) ComposedEmail {
	locale := ParseLocale(string(req.Locale))
	strs := latePaymentRefundCopy[locale]
	date := FormatDate(req.StartTime, locale)
	timeRange := FormatTimeRange(req.StartTime, req.EndTime, locale)
	amount := FormatMoney(req.RefundedAmountCents, req.Currency, locale)

	return ComposedEmail{
		Subject:  strs.Subject,
		Template: LatePaymentRefundTemplate,
		Data: LatePaymentRefundView{
			Lang:         string(locale),
			T:            strs,
			FirstName:    req.ToFirstName,
			ProductName:  req.ProductName,
			Date:         date,
			TimeRange:    timeRange,
			Intro:        fmt.Sprintf(strs.Intro, date, timeRange),
			RefundText:   fmt.Sprintf(strs.RefundText, amount),
			Amount:       amount,
			BookAgainURL: req.BookAgainURL,
			BookingID:    req.BookingID,
			Year:         now.In(practiceLocation).Year(),
		},
	}
}
