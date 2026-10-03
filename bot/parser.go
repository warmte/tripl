package bot

import (
	"log"
	"strings"
	"time"

	. "github.com/warmte/tripl/common"
)

func parseDatetime(input string, layout string, name string) (*time.Time, error) {
	t, err := time.Parse(layout, input)
	if err != nil {
		log.Printf("Unable parse %s datetime:\n%s", name, err.Error())
		return nil, err
	}
	return &t, nil
}

func parseOptionalDatetime(input string, layout string, name string) (*time.Time, error) {
	if input == "-" {
		return nil, nil
	}
	return parseDatetime(input, layout, name)
}

func opt(input string) *string {
	if input == "-" {
		return nil
	} else {
		return &input
	}
}

func hasTooLongField(fields ...string) bool {
	for _, field := range fields {
		if field != "-" && len([]rune(field)) > 256 {
			return true
		}
	}
	return false
}

func parseTicketMainFields(input string, userID int64) (Ticket, error) {
	lines := strings.Split(input, "\n")
	if len(lines) != 5 {
		return Ticket{}, ErrIncorrectNumberOfFields
	}
	if hasTooLongField(lines...) {
		return Ticket{}, ErrTooLongField
	}

	departureTime, err := parseDatetime(lines[3], DatetimeLayout, "ticket departure")
	if err != nil {
		return Ticket{}, ErrIncorrectDateTime
	}
	arrivalTime, err := parseDatetime(lines[4], DatetimeLayout, "ticket arrival")
	if err != nil {
		return Ticket{}, ErrIncorrectDateTime
	}

	return Ticket{
		UserID:         userID,
		Vehicle:        lines[0],
		DeparturePoint: lines[1],
		ArrivalPoint:   lines[2],
		DepartureTime:  *departureTime,
		ArrivalTime:    *arrivalTime,
	}, nil
}

func parseTicketOtherFields(input string, ticket *Ticket) (Ticket, error) {
	lines := strings.Split(input, "\n")
	if len(lines) != 7 {
		return Ticket{}, ErrIncorrectNumberOfFields
	}
	if hasTooLongField(lines...) {
		return Ticket{}, ErrTooLongField
	}

	return Ticket{
		UserID:           ticket.UserID,
		Vehicle:          ticket.Vehicle,
		DeparturePoint:   ticket.DeparturePoint,
		ArrivalPoint:     ticket.ArrivalPoint,
		DepartureTime:    ticket.DepartureTime,
		ArrivalTime:      ticket.ArrivalTime,
		Cost:             opt(lines[0]),
		Currency:         opt(lines[1]),
		DepartureAddress: opt(lines[2]),
		ArrivalAddress:   opt(lines[3]),
		Seat:             opt(lines[4]),
		Number:           opt(lines[5]),
		Comment:          opt(lines[6]),
	}, nil
}

func parseBookingMainFields(input string, userID int64) (Booking, error) {
	lines := strings.Split(input, "\n")
	if len(lines) != 4 {
		return Booking{}, ErrIncorrectNumberOfFields
	}
	if hasTooLongField(lines...) {
		return Booking{}, ErrTooLongField
	}

	checkInDate, err := parseDatetime(lines[2], DateLayout, "booking check-in date")
	if err != nil {
		return Booking{}, ErrIncorrectDate
	}
	checkOutDate, err := parseDatetime(lines[3], DateLayout, "booking check-out date")
	if err != nil {
		return Booking{}, ErrIncorrectDate
	}

	return Booking{
		UserID:       userID,
		City:         lines[0],
		Place:        lines[1],
		CheckInDate:  *checkInDate,
		CheckOutDate: *checkOutDate,
	}, nil
}

func parseBookingOtherFields(input string, booking *Booking) (Booking, error) {
	lines := strings.Split(input, "\n")
	if len(lines) != 7 {
		return Booking{}, ErrIncorrectNumberOfFields
	}
	if hasTooLongField(lines...) {
		return Booking{}, ErrTooLongField
	}

	checkInTime, err := parseOptionalDatetime(lines[3], TimeLayout, "booking check-in time")
	if err != nil {
		return Booking{}, ErrIncorrectTime
	}
	checkOutTime, err := parseOptionalDatetime(lines[4], TimeLayout, "booking check-out time")
	if err != nil {
		return Booking{}, ErrIncorrectTime
	}

	return Booking{
		UserID:       booking.UserID,
		City:         booking.City,
		Place:        booking.Place,
		CheckInDate:  booking.CheckInDate,
		CheckOutDate: booking.CheckOutDate,
		Cost:         opt(lines[0]),
		Currency:     opt(lines[1]),
		Address:      opt(lines[2]),
		CheckInTime:  checkInTime,
		CheckOutTime: checkOutTime,
		Number:       opt(lines[5]),
		Comment:      opt(lines[6]),
	}, nil
}

func parseEventMainFields(input string, userID int64) (Event, error) {
	lines := strings.Split(input, "\n")
	if len(lines) != 3 {
		return Event{}, ErrIncorrectNumberOfFields
	}
	if hasTooLongField(lines...) {
		return Event{}, ErrTooLongField
	}

	startTime, err := parseDatetime(lines[2], DatetimeLayout, "event start")
	if err != nil {
		return Event{}, ErrIncorrectDateTime
	}

	return Event{
		UserID:    userID,
		Name:      lines[0],
		Location:  lines[1],
		StartTime: *startTime,
	}, nil
}

func parseEventOtherFields(input string, event *Event) (Event, error) {
	lines := strings.Split(input, "\n")
	if len(lines) != 4 {
		return Event{}, ErrIncorrectNumberOfFields
	}
	if hasTooLongField(lines...) {
		return Event{}, ErrTooLongField
	}

	endTime, err := parseOptionalDatetime(lines[0], DatetimeLayout, "event end")
	if err != nil {
		return Event{}, ErrIncorrectDateTime
	}

	return Event{
		UserID:    event.UserID,
		Name:      event.Name,
		Location:  event.Location,
		StartTime: event.StartTime,
		EndTime:   endTime,
		Cost:      opt(lines[1]),
		Currency:  opt(lines[2]),
		Comment:   opt(lines[3]),
	}, nil
}

func parseTripMainFields(input string, userID int64) (Trip, error) {
	lines := strings.Split(input, "\n")
	if len(lines) != 3 {
		return Trip{}, ErrIncorrectNumberOfFields
	}
	if hasTooLongField(lines...) {
		return Trip{}, ErrTooLongField
	}

	startDate, err := parseDatetime(lines[1], DateLayout, "trip start")
	if err != nil {
		return Trip{}, ErrIncorrectDate
	}
	endDate, err := parseDatetime(lines[2], DateLayout, "trip end")
	if err != nil {
		return Trip{}, ErrIncorrectDate
	}

	return Trip{
		UserID:    userID,
		Name:      lines[0],
		StartDate: *startDate,
		EndDate:   *endDate,
	}, nil
}

func parseTripOtherFields(input string, trip *Trip) (Trip, error) {
	lines := strings.Split(input, "\n")
	if len(lines) != 1 {
		return Trip{}, ErrIncorrectNumberOfFields
	}
	if hasTooLongField(lines...) {
		return Trip{}, ErrTooLongField
	}

	return Trip{
		UserID:    trip.UserID,
		Name:      trip.Name,
		StartDate: trip.StartDate,
		EndDate:   trip.EndDate,
		Comment:   opt(lines[0]),
	}, nil
}

func selectMainFieldsParser(input string, label Label, userID int64) (TripObjectView, error) {
	switch label {
	case TicketLabel:
		return parseTicketMainFields(input, userID)

	case BookingLabel:
		return parseBookingMainFields(input, userID)

	case EventLabel:
		return parseEventMainFields(input, userID)

	case TripLabel:
		return parseTripMainFields(input, userID)

	default:
		return nil, ErrIncorrectInput
	}
}

func selectOtherFieldsParser(input string, label Label, current TripObjectView) (TripObjectView, error) {
	switch label {
	case TicketLabel:
		value, ok := current.(Ticket)
		if !ok {
			return nil, ErrInternalError
		}
		return parseTicketOtherFields(input, &value)

	case BookingLabel:
		value, ok := current.(Booking)
		if !ok {
			return nil, ErrInternalError
		}
		return parseBookingOtherFields(input, &value)

	case EventLabel:
		value, ok := current.(Event)
		if !ok {
			return nil, ErrInternalError
		}
		return parseEventOtherFields(input, &value)

	case TripLabel:
		value, ok := current.(Trip)
		if !ok {
			return nil, ErrInternalError
		}
		return parseTripOtherFields(input, &value)

	default:
		return nil, ErrIncorrectInput
	}
}
