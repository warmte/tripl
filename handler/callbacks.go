package handler

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/warmte/tripl/common"
)

type CallbackAction string

const (
	ArchiveList           CallbackAction = "archive"
	ActualList            CallbackAction = "actual"
	TomorrowList          CallbackAction = "tomorrow"
	DeleteSingle          CallbackAction = "delete"
	DeleteSingleConfirmed CallbackAction = "delete_confirmed"
	EditSingle            CallbackAction = "edit"
	EditField             CallbackAction = "edit_field"
	AttachSingleActual    CallbackAction = "attach_actual"
	AttachSingleArchived  CallbackAction = "attach_archived"
	AttachTrip            CallbackAction = "attach_trip"
	DetachSingle          CallbackAction = "detach"
	GetSingle             CallbackAction = "get"
	TripStats             CallbackAction = "trip_stats"
	TripTickets           CallbackAction = "trips_tickets"
	TripBookings          CallbackAction = "trips_bookings"
	TripEvents            CallbackAction = "trips_events"
)

type Callback struct {
	Label  Label
	Action CallbackAction
	V1     int64
}

type ExtendedCallback struct {
	Label  Label
	Action CallbackAction
	V1     int64
	V2     int64
}

func MakeCallback(label Label, action CallbackAction, id int64) string {
	return fmt.Sprintf("%s:%s:%d", label, action, id)
}

func MakeExtendedCallback(label Label, action CallbackAction, id int64, value int64) string {
	return fmt.Sprintf("%s:%s:%d:%d", label, action, id, value)
}

func parseCallback(data string) (Callback, error) {
	parts := strings.Split(data, ":")

	if len(parts) != 3 {
		return Callback{}, fmt.Errorf(
			"invalid callback format: %q",
			data,
		)
	}

	label := Label(parts[0])

	switch label {
	case TicketLabel, BookingLabel, EventLabel, TripLabel:
	default:
		return Callback{}, fmt.Errorf(
			"unknown label: %q",
			label,
		)
	}

	action := CallbackAction(parts[1])
	switch action {
	case ActualList, ArchiveList, EditSingle, DeleteSingle, GetSingle, DeleteSingleConfirmed, DetachSingle, TripStats, TomorrowList:
	default:
		return Callback{}, fmt.Errorf(
			"unknown action: %q",
			action,
		)
	}

	v1, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || v1 < 0 {
		return Callback{}, fmt.Errorf(
			"invalid callback v1: %q",
			parts[2],
		)
	}

	return Callback{label, action, v1}, nil
}

func parseExtendedCallback(data string) (ExtendedCallback, error) {
	parts := strings.Split(data, ":")

	if len(parts) != 4 {
		return ExtendedCallback{}, fmt.Errorf(
			"invalid extended callback format: %q",
			data,
		)
	}

	label := Label(parts[0])

	switch label {
	case TicketLabel, BookingLabel, EventLabel, TripLabel:
	default:
		return ExtendedCallback{}, fmt.Errorf(
			"invalid extended callback label: %q",
			label,
		)
	}

	v1, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || v1 < 0 {
		return ExtendedCallback{}, fmt.Errorf(
			"invalid callback v1: %q",
			parts[2],
		)
	}

	v2, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil || v2 < 0 {
		return ExtendedCallback{}, fmt.Errorf(
			"invalid callback v2: %q",
			parts[3],
		)
	}

	action := CallbackAction(parts[1])
	switch action {
	case AttachSingleActual, AttachSingleArchived, AttachTrip, EditField, TripTickets, TripBookings, TripEvents:
	default:
		return ExtendedCallback{}, fmt.Errorf(
			"unknown action: %q",
			action,
		)
	}

	return ExtendedCallback{label, action, v1, v2}, nil
}
