package domain

import (
	"fmt"
	"time"
)

// Booking cancellation emails (issue 017): the client's, in the locale they
// booked in, stating the refund the cancellation tiers gave (ADR-0014); and
// the partner's notice, in French. No health data (ADR-0019), no free-text
// reason.

// Who cancelled a booking, as the cancellation emails tell it.
const (
	CancelledByClient       = "client"
	CancelledByPractitioner = "practitioner"
)

// BookingClientCancellationRequest tells a client their booking is cancelled
// and what was refunded.
type BookingClientCancellationRequest struct {
	ToEmail     string
	ToFirstName string
	ToLastName  string
	Locale      Locale

	BookingID   string
	ProductName string

	StartTime time.Time
	EndTime   time.Time

	// CancelledBy is CancelledByClient or CancelledByPractitioner; anything
	// else reads as a cancellation without a stated author.
	CancelledBy string

	// AmountPaidOnlineCents is what the client had paid online;
	// RefundedAmountCents what the cancellation refunded.
	AmountPaidOnlineCents int
	RefundedAmountCents   int
	Currency              string

	// BookAgainURL opens the practice page in the client's language. Empty
	// when the public site's origin is unknown.
	BookAgainURL string
}

// BookingPartnerCancellationRequest tells the partner a booking was
// cancelled. Partner emails are French.
type BookingPartnerCancellationRequest struct {
	ToEmail     string
	ToFirstName string

	BookingID   string
	ProductName string
	ClientName  string

	StartTime time.Time
	EndTime   time.Time

	CancelledBy         string
	RefundedAmountCents int
	Currency            string
}

// Template names of the cancellation emails (templates/<name>.html).
const (
	BookingClientCancellationTemplate  = "booking_cancellation"
	BookingPartnerCancellationTemplate = "booking_partner_cancellation"
)

// clientCancellationStrings is the copy of the client cancellation email.
type clientCancellationStrings struct {
	Subject             string
	Title               string
	Greeting            string
	IntroClient         string // fmt: date, time range
	IntroPractitioner   string // fmt: date, time range
	IntroOther          string // fmt: date, time range
	RefundFull          string // fmt: amount
	RefundPartial       string // fmt: refunded amount, amount paid
	NoRefund            string
	NothingPaid         string
	DelayNote           string
	DetailsHeading      string
	ServiceLabel        string
	DateLabel           string
	TimeLabel           string
	PaidLabel           string
	RefundLabel         string
	ReferenceLabel      string
	BookAgainText       string
	BookAgainLink       string
	SignOff             string
	SignOffPractitioner string
	Team                string
	Rights              string
}

var clientCancellationCopy = map[Locale]clientCancellationStrings{
	LocaleFR: {
		Subject:             "Votre rendez-vous est annulé",
		Title:               "Rendez-vous annulé",
		Greeting:            "Bonjour",
		IntroClient:         "Votre rendez-vous du %s (%s) est bien annulé.",
		IntroPractitioner:   "Votre praticien a dû annuler votre rendez-vous du %s (%s).",
		IntroOther:          "Votre rendez-vous du %s (%s) a été annulé.",
		RefundFull:          "Votre paiement de %s vous est intégralement remboursé.",
		RefundPartial:       "Selon les conditions d'annulation, %s vous sont remboursés sur les %s payés en ligne.",
		NoRefund:            "Selon les conditions d'annulation, cette annulation ne donne pas lieu à un remboursement.",
		NothingPaid:         "Aucun paiement en ligne n'avait été effectué : il n'y a rien à rembourser.",
		DelayNote:           "Le remboursement apparaît sur votre relevé sous 5 à 10 jours ouvrés, selon votre banque.",
		DetailsHeading:      "Rendez-vous annulé",
		ServiceLabel:        "Prestation",
		DateLabel:           "Date",
		TimeLabel:           "Heure",
		PaidLabel:           "Payé en ligne",
		RefundLabel:         "Remboursé",
		ReferenceLabel:      "Référence",
		BookAgainText:       "Vous pouvez réserver un autre créneau à tout moment :",
		BookAgainLink:       "Réserver un autre créneau",
		SignOff:             "À bientôt,",
		SignOffPractitioner: "Avec toutes nos excuses,",
		Team:                "L'équipe Leviosa",
		Rights:              "Tous droits réservés.",
	},
	LocaleEN: {
		Subject:             "Your appointment is cancelled",
		Title:               "Appointment cancelled",
		Greeting:            "Hello",
		IntroClient:         "Your appointment on %s (%s) is cancelled.",
		IntroPractitioner:   "Your practitioner had to cancel your appointment on %s (%s).",
		IntroOther:          "Your appointment on %s (%s) has been cancelled.",
		RefundFull:          "Your payment of %s is refunded in full.",
		RefundPartial:       "Under the cancellation terms, %s of the %s paid online is refunded to you.",
		NoRefund:            "Under the cancellation terms, this cancellation is not refunded.",
		NothingPaid:         "Nothing was paid online, so there is nothing to refund.",
		DelayNote:           "The refund shows on your statement within 5 to 10 business days, depending on your bank.",
		DetailsHeading:      "Cancelled appointment",
		ServiceLabel:        "Treatment",
		DateLabel:           "Date",
		TimeLabel:           "Time",
		PaidLabel:           "Paid online",
		RefundLabel:         "Refunded",
		ReferenceLabel:      "Reference",
		BookAgainText:       "You can book another slot at any time:",
		BookAgainLink:       "Book another slot",
		SignOff:             "See you soon,",
		SignOffPractitioner: "With our apologies,",
		Team:                "The Leviosa team",
		Rights:              "All rights reserved.",
	},
}

// BookingClientCancellationView is the data of the booking_cancellation
// template.
type BookingClientCancellationView struct {
	Lang         string
	T            clientCancellationStrings
	FirstName    string
	ProductName  string
	Date         string
	TimeRange    string
	Intro        string
	RefundText   string
	ShowDelay    bool
	ShowAmounts  bool
	PaidAmount   string
	RefundAmount string
	SignOff      string
	BookAgainURL string
	BookingID    string
	Year         int
}

// ComposeBookingClientCancellation builds the client cancellation email in
// the request's locale. now stamps the footer year.
func ComposeBookingClientCancellation(req BookingClientCancellationRequest, now time.Time) ComposedEmail {
	locale := ParseLocale(string(req.Locale))
	strs := clientCancellationCopy[locale]
	date := FormatDate(req.StartTime, locale)
	timeRange := FormatTimeRange(req.StartTime, req.EndTime, locale)

	intro := strs.IntroOther
	signOff := strs.SignOff
	switch req.CancelledBy {
	case CancelledByClient:
		intro = strs.IntroClient
	case CancelledByPractitioner:
		intro = strs.IntroPractitioner
		signOff = strs.SignOffPractitioner
	}

	paid := FormatMoney(req.AmountPaidOnlineCents, req.Currency, locale)
	refunded := FormatMoney(req.RefundedAmountCents, req.Currency, locale)
	var refundText string
	switch {
	case req.AmountPaidOnlineCents <= 0:
		refundText = strs.NothingPaid
	case req.RefundedAmountCents <= 0:
		refundText = strs.NoRefund
	case req.RefundedAmountCents >= req.AmountPaidOnlineCents:
		refundText = fmt.Sprintf(strs.RefundFull, refunded)
	default:
		refundText = fmt.Sprintf(strs.RefundPartial, refunded, paid)
	}

	return ComposedEmail{
		Subject:  strs.Subject,
		Template: BookingClientCancellationTemplate,
		Data: BookingClientCancellationView{
			Lang:         string(locale),
			T:            strs,
			FirstName:    req.ToFirstName,
			ProductName:  req.ProductName,
			Date:         date,
			TimeRange:    timeRange,
			Intro:        fmt.Sprintf(intro, date, timeRange),
			RefundText:   refundText,
			ShowDelay:    req.RefundedAmountCents > 0,
			ShowAmounts:  req.AmountPaidOnlineCents > 0,
			PaidAmount:   paid,
			RefundAmount: refunded,
			SignOff:      signOff,
			BookAgainURL: req.BookAgainURL,
			BookingID:    req.BookingID,
			Year:         now.In(practiceLocation).Year(),
		},
	}
}

// BookingPartnerCancellationView is the data of the
// booking_partner_cancellation template.
type BookingPartnerCancellationView struct {
	FirstName    string
	Intro        string
	ProductName  string
	ClientName   string
	Date         string
	TimeRange    string
	RefundAmount string
	BookingID    string
	Year         int
}

// ComposeBookingPartnerCancellation builds the French cancellation notice
// for the partner.
func ComposeBookingPartnerCancellation(req BookingPartnerCancellationRequest, now time.Time) ComposedEmail {
	date := FormatDate(req.StartTime, LocaleFR)
	intro := "Une réservation a été annulée."
	switch req.CancelledBy {
	case CancelledByClient:
		intro = "Une réservation a été annulée par le client. Le créneau est de nouveau proposé à la réservation."
	case CancelledByPractitioner:
		intro = "Vous avez annulé cette réservation. Le client est remboursé de ce qu'il a payé en ligne et prévenu par email."
	}
	return ComposedEmail{
		Subject:  fmt.Sprintf("Réservation annulée : %s, %s à %s", req.ProductName, date, FormatTime(req.StartTime, LocaleFR)),
		Template: BookingPartnerCancellationTemplate,
		Data: BookingPartnerCancellationView{
			FirstName:    req.ToFirstName,
			Intro:        intro,
			ProductName:  req.ProductName,
			ClientName:   req.ClientName,
			Date:         date,
			TimeRange:    FormatTimeRange(req.StartTime, req.EndTime, LocaleFR),
			RefundAmount: FormatMoney(req.RefundedAmountCents, req.Currency, LocaleFR),
			BookingID:    req.BookingID,
			Year:         now.In(practiceLocation).Year(),
		},
	}
}
