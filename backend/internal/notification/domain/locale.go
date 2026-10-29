package domain

import (
	"fmt"
	"strings"
	"time"

	// Embed the time zone database so Europe/Paris loads on images without
	// system tzdata.
	_ "time/tzdata"
)

// Locale is the language a client-facing email is written in. Booking emails
// follow the locale the client booked in; partner and staff emails are
// French.
type Locale string

const (
	LocaleFR Locale = "fr"
	LocaleEN Locale = "en"
)

// ParseLocale reads a stored locale. Anything but "en" is French, the
// practice's default language (bookings made before the locale was recorded
// have none).
func ParseLocale(value string) Locale {
	if Locale(strings.ToLower(strings.TrimSpace(value))) == LocaleEN {
		return LocaleEN
	}
	return LocaleFR
}

var practiceLocation = mustLoadLocation("Europe/Paris")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		// Unreachable: the tzdata database is embedded above.
		panic("notification: cannot load " + name + ": " + err.Error())
	}
	return loc
}

// PracticeLocation is the practice's time zone: every date and time in an
// email is shown in it, whatever the server's zone.
func PracticeLocation() *time.Location {
	return practiceLocation
}

var (
	frenchWeekdays = [...]string{"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi"}
	frenchMonths   = [...]string{"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"}
)

// FormatDate renders the calendar date of t in the practice's time zone:
// "mardi 3 novembre 2026" in French, "Tuesday 3 November 2026" in English.
func FormatDate(t time.Time, locale Locale) string {
	local := t.In(practiceLocation)
	if locale == LocaleEN {
		return local.Format("Monday 2 January 2006")
	}
	return fmt.Sprintf("%s %d %s %d",
		frenchWeekdays[local.Weekday()], local.Day(), frenchMonths[local.Month()-1], local.Year())
}

// FormatTime renders the time of day of t in the practice's time zone:
// "14h30" in French, "14:30" in English.
func FormatTime(t time.Time, locale Locale) string {
	local := t.In(practiceLocation)
	if locale == LocaleEN {
		return local.Format("15:04")
	}
	return local.Format("15h04")
}

// FormatTimeRange renders "14h30 – 15h30" (French) or "14:30 – 15:30"
// (English) in the practice's time zone.
func FormatTimeRange(start, end time.Time, locale Locale) string {
	return FormatTime(start, locale) + " – " + FormatTime(end, locale)
}

// FormatMoney renders an amount in cents: "45,00 €" in French, "€45.00" in
// English. Currencies without a known symbol show their code.
func FormatMoney(cents int, currency string, locale Locale) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	units, rest := cents/100, cents%100
	code := strings.ToUpper(currency)
	symbol, known := map[string]string{"EUR": "€", "USD": "$", "GBP": "£"}[code]

	if locale == LocaleEN {
		if known {
			return fmt.Sprintf("%s%s%d.%02d", sign, symbol, units, rest)
		}
		return fmt.Sprintf("%s%d.%02d %s", sign, units, rest, code)
	}
	if !known {
		symbol = code
	}
	return fmt.Sprintf("%s%d,%02d %s", sign, units, rest, symbol)
}
