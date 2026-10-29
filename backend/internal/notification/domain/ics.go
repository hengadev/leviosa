package domain

import (
	"strings"
	"time"
)

// CalendarEvent is one appointment exported as an iCalendar file (RFC 5545),
// attached to booking emails so the client can add it to a calendar.
type CalendarEvent struct {
	// UID identifies the event across exports: the same booking always gets
	// the same UID, so a later export updates the event instead of adding a
	// second one.
	UID         string
	Start       time.Time
	End         time.Time
	Summary     string
	Location    string
	Description string
	URL         string
	// Stamp is when the file was produced (DTSTAMP).
	Stamp time.Time
}

// icsMaxLineOctets is the RFC 5545 limit for a content line, CRLF excluded.
const icsMaxLineOctets = 75

const icsDateTimeUTC = "20060102T150405Z"

// ICS renders the event as an iCalendar object. Times are written in UTC
// ("Z" form), which every calendar application converts to the viewer's
// local time, so no VTIMEZONE is needed. Text values are escaped, lines end
// with CRLF and lines longer than 75 octets are folded.
func (e CalendarEvent) ICS() []byte {
	var b strings.Builder
	write := func(line string) {
		b.WriteString(foldICSLine(line))
		b.WriteString("\r\n")
	}

	write("BEGIN:VCALENDAR")
	write("VERSION:2.0")
	write("PRODID:-//Leviosa//Booking//FR")
	write("CALSCALE:GREGORIAN")
	write("METHOD:PUBLISH")
	write("BEGIN:VEVENT")
	write("UID:" + escapeICSText(e.UID))
	write("DTSTAMP:" + e.Stamp.UTC().Format(icsDateTimeUTC))
	write("DTSTART:" + e.Start.UTC().Format(icsDateTimeUTC))
	write("DTEND:" + e.End.UTC().Format(icsDateTimeUTC))
	write("SUMMARY:" + escapeICSText(e.Summary))
	if e.Location != "" {
		write("LOCATION:" + escapeICSText(e.Location))
	}
	if e.Description != "" {
		write("DESCRIPTION:" + escapeICSText(e.Description))
	}
	if e.URL != "" {
		// URI value: not a TEXT value, so not escaped.
		write("URL:" + e.URL)
	}
	write("STATUS:CONFIRMED")
	write("TRANSP:OPAQUE")
	write("END:VEVENT")
	write("END:VCALENDAR")

	return []byte(b.String())
}

// escapeICSText escapes a TEXT value (RFC 5545 §3.3.11): backslash,
// semicolon and comma are backslash-escaped, line breaks become "\n".
func escapeICSText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.NewReplacer(
		`\`, `\\`,
		`;`, `\;`,
		`,`, `\,`,
		"\n", `\n`,
	).Replace(s)
}

// foldICSLine folds a content line longer than 75 octets (RFC 5545 §3.1):
// each continuation line starts with one space, which counts toward its 75
// octets. Lines are only split between UTF-8 characters, never inside one.
func foldICSLine(line string) string {
	if len(line) <= icsMaxLineOctets {
		return line
	}

	var b strings.Builder
	lineOctets := 0
	for _, r := range line {
		size := len(string(r))
		if lineOctets+size > icsMaxLineOctets {
			b.WriteString("\r\n ")
			lineOctets = 1
		}
		b.WriteRune(r)
		lineOctets += size
	}
	return b.String()
}
