package document

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
)

const icsDateTimeLayout = "20060102T150405Z"

func ICSContent(obj TripObjectView, startZone, endZone string, cfg *lang.Cfg) ([]byte, error) {
	var (
		uid, summary, location, description string
		start, end                          time.Time
		allDay                              bool
		err                                 error
	)

	switch v := obj.(type) {
	case Ticket:
		if v.DepartureTime == nil {
			return nil, ErrDatetimeRestraint
		}

		uid = fmt.Sprintf("ticket-%d@tripl", v.ID)
		summary = icsTicketSummary(v, cfg)

		start, err = icsLocalTime(*v.DepartureTime, startZone)
		if err != nil {
			return nil, err
		}
		if v.ArrivalTime != nil {
			end, err = icsLocalTime(*v.ArrivalTime, endZone)
			if err != nil {
				return nil, err
			}
		}

		description = obj.Show(&cfg.TripObjects.Ticket)

	case Booking:
		if v.CheckInDate == nil {
			return nil, ErrDatetimeRestraint
		}

		uid, summary, allDay = fmt.Sprintf("booking-%d@tripl", v.ID), icsBookingSummary(v, cfg), true
		start = *v.CheckInDate
		end = start.AddDate(0, 0, 1)
		if v.CheckOutDate != nil {
			end = *v.CheckOutDate
		}
		end = end.AddDate(0, 0, 1)

		if v.Address != nil {
			location = *v.Address
		}

		description = obj.Show(&cfg.TripObjects.Booking)

	case Event:
		if v.StartTime == nil {
			return nil, fmt.Errorf("event %d: missing start time", v.ID)
		}

		uid = fmt.Sprintf("event-%d@tripl", v.ID)
		if v.Name != nil {
			summary = *v.Name
		}
		if v.Location != nil {
			location = *v.Location
		}

		start, err = icsLocalTime(*v.StartTime, startZone)
		if err != nil {
			return nil, err
		}
		end = start.Add(1 * time.Hour)
		if v.EndTime != nil {
			end = *v.EndTime
			end, err = icsLocalTime(*v.EndTime, endZone)
			if err != nil {
				return nil, err
			}
		}

		description = obj.Show(&cfg.TripObjects.Event)

	default:
		return nil, fmt.Errorf("unsupported ICS object type %T", obj)
	}

	if !end.After(start) {
		return nil, ErrDatetimeRestraint
	}

	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//Tripl//Calendar Export//RU",
		"CALSCALE:GREGORIAN",
		"BEGIN:VEVENT",
		"UID:" + icsEscape(uid),
		"DTSTAMP:" + time.Now().UTC().Format(icsDateTimeLayout),
	}

	if allDay {
		lines = append(lines,
			"DTSTART;VALUE=DATE:"+start.Format("20060102"),
			"DTEND;VALUE=DATE:"+end.Format("20060102"),
		)
	} else {
		lines = append(lines, "DTSTART:"+start.UTC().Format(icsDateTimeLayout))
		if !end.IsZero() {
			lines = append(lines, "DTEND:"+end.UTC().Format(icsDateTimeLayout))
		}
	}

	lines = append(lines, "SUMMARY:"+icsEscape(summary))
	if location != "" {
		lines = append(lines, "LOCATION:"+icsEscape(location))
	}
	if description != "" {
		lines = append(lines, "DESCRIPTION:"+icsEscape(description))
	}
	lines = append(lines, "END:VEVENT", "END:VCALENDAR")

	var b strings.Builder
	for _, line := range lines {
		b.WriteString(icsFoldLine(line))
		b.WriteString("\r\n")
	}
	return []byte(b.String()), nil
}

func icsLocalTime(value time.Time, zone string) (time.Time, error) {
	if zone == "" {
		return time.Time{}, fmt.Errorf("time zone is required")
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time zone %q: %w", zone, err)
	}
	return time.Date(
		value.Year(), value.Month(), value.Day(),
		value.Hour(), value.Minute(), value.Second(), value.Nanosecond(),
		loc,
	).UTC(), nil
}

func icsTicketSummary(v Ticket, cfg *lang.Cfg) string {
	var parts []string
	for _, s := range []*string{v.Vehicle, v.DeparturePoint, v.ArrivalPoint} {
		if s != nil && *s != "" {
			parts = append(parts, *s)
		}
	}
	if len(parts) == 0 {
		return fmt.Sprintf("%s №%d", cfg.TripObjects.Ticket.One, v.ID)
	}
	return strings.Join(parts, " — ")
}

func icsBookingSummary(v Booking, cfg *lang.Cfg) string {
	var parts []string
	if v.Place != nil && *v.Place != "" {
		parts = append(parts, *v.Place)
	} else {
		parts = append(parts, fmt.Sprintf("%s №%d", cfg.TripObjects.Booking.One, v.ID))
	}
	if v.City != nil && *v.City != "" {
		parts = append(parts, *v.City)
	}
	return strings.Join(parts, " — ")
}

func icsEscape(s string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		";", `\;`,
		",", `\,`,
		"\r\n", `\n`,
		"\n", `\n`,
		"\r", `\n`,
	).Replace(s)
}

func icsFoldLine(s string) string {
	const limit = 75
	var b strings.Builder
	col := 0

	for len(s) > 0 {
		_, size := utf8.DecodeRuneInString(s)
		if col+size > limit {
			b.WriteString("\r\n ")
			col = 1
		}
		b.WriteString(s[:size])
		s = s[size:]
		col += size
	}
	return b.String()
}
