package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/warmte/tripl/lang"
)

const (
	DatetimeLayout = "02.01.2006 15:04"
	DateLayout     = "02.01.2006"
	TimeLayout     = "15:04"
)

const PageSize = 3

type Page[T any] struct {
	Items   []T
	HasNext bool
}

type ListType string

const (
	ActualListType  ListType = "actual"
	ArchiveListType ListType = "archive"
)

type TripObjectView interface {
	Show(cfg *lang.TripObjectCfg) string
	Preview(cfg *lang.TripObjectCfg) string
	GetTripID() *int64
}

type TripObject struct {
	Label Label
	Cfg   *lang.TripObjectCfg
}

type Label string

const (
	TicketLabel  Label = "ticket"
	BookingLabel Label = "booking"
	EventLabel   Label = "event"
	TripLabel    Label = "trip"
)

var Labels = []Label{TicketLabel, BookingLabel, EventLabel, TripLabel}

func TripObjects(text *lang.Cfg) []TripObject {
	return []TripObject{{TicketLabel, &text.TripObjects.Ticket}, {BookingLabel, &text.TripObjects.Booking}, {EventLabel, &text.TripObjects.Event}, {TripLabel, &text.TripObjects.Trip}}
}

func TripObjectCfgByLabel(text *lang.Cfg, label Label) lang.TripObjectCfg {
	switch label {
	case TicketLabel:
		return text.TripObjects.Ticket
	case BookingLabel:
		return text.TripObjects.Booking
	case EventLabel:
		return text.TripObjects.Event
	case TripLabel:
		return text.TripObjects.Trip
	default:
		return lang.TripObjectCfg{}
	}
}

type Ticket struct {
	ID               int64
	UserID           int64
	TripID           *int64
	Vehicle          string
	DeparturePoint   string
	ArrivalPoint     string
	DepartureTime    time.Time
	ArrivalTime      time.Time
	Cost             *string
	DepartureAddress *string
	ArrivalAddress   *string
	Seat             *string
	Comment          *string
}

type Booking struct {
	ID           int64
	UserID       int64
	TripID       *int64
	City         string
	Place        string
	CheckInDate  time.Time
	CheckOutDate time.Time
	Cost         *string
	Address      *string
	CheckInTime  *time.Time
	CheckOutTime *time.Time
	Comment      *string
}

type Event struct {
	ID        int64
	UserID    int64
	TripID    *int64
	Name      string
	Location  string
	StartTime time.Time
	EndTime   *time.Time
	Cost      *string
	Comment   *string
}

type Trip struct {
	ID        int64
	UserID    int64
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Comment   *string
}

func addStringField(lines *[]string, name, value string) {
	*lines = append(*lines, fmt.Sprintf("%s: %s", name, value))
}

func addField(lines *[]string, name string, value *string) {
	if value == nil {
		return
	}

	addStringField(lines, name, *value)
}

func (f Ticket) Show(cfg *lang.TripObjectCfg) string {
	var lines []string

	addStringField(
		&lines,
		cfg.MainFields[0],
		f.Vehicle,
	)

	addStringField(
		&lines,
		cfg.MainFields[1],
		f.DeparturePoint,
	)

	addStringField(
		&lines,
		cfg.MainFields[2],
		f.ArrivalPoint,
	)

	addStringField(
		&lines,
		cfg.MainFields[3],
		f.DepartureTime.Format(DatetimeLayout),
	)

	addStringField(
		&lines,
		cfg.MainFields[4],
		f.ArrivalTime.Format(DatetimeLayout),
	)

	addField(
		&lines,
		cfg.OtherFields[0],
		f.Cost,
	)

	addField(
		&lines,
		cfg.OtherFields[1],
		f.DepartureAddress,
	)

	addField(
		&lines,
		cfg.OtherFields[2],
		f.ArrivalAddress,
	)

	addField(
		&lines,
		cfg.OtherFields[3],
		f.Seat,
	)

	addField(
		&lines,
		cfg.OtherFields[4],
		f.Comment,
	)

	return strings.Join(lines, "\n")
}

func (f Booking) Show(cfg *lang.TripObjectCfg) string {
	var lines []string

	addStringField(
		&lines,
		cfg.MainFields[0],
		f.City,
	)

	addStringField(
		&lines,
		cfg.MainFields[1],
		f.Place,
	)

	addStringField(
		&lines,
		cfg.MainFields[2],
		f.CheckInDate.Format(DateLayout),
	)

	addStringField(
		&lines,
		cfg.MainFields[3],
		f.CheckOutDate.Format(DateLayout),
	)

	addField(
		&lines,
		cfg.OtherFields[0],
		f.Cost,
	)

	addField(
		&lines,
		cfg.OtherFields[1],
		f.Address,
	)

	if f.CheckInTime != nil {
		addStringField(
			&lines,
			cfg.OtherFields[2],
			f.CheckInTime.Format(TimeLayout),
		)
	}

	if f.CheckOutTime != nil {
		addStringField(
			&lines,
			cfg.OtherFields[3],
			f.CheckOutTime.Format(TimeLayout),
		)
	}

	addField(
		&lines,
		cfg.OtherFields[4],
		f.Comment,
	)

	return strings.Join(lines, "\n")
}

func (f Event) Show(cfg *lang.TripObjectCfg) string {
	var lines []string

	addStringField(
		&lines,
		cfg.MainFields[0],
		f.Name,
	)

	addStringField(
		&lines,
		cfg.MainFields[1],
		f.Location,
	)

	addStringField(
		&lines,
		cfg.MainFields[2],
		f.StartTime.Format(DatetimeLayout),
	)

	if f.EndTime != nil {
		addStringField(
			&lines,
			cfg.OtherFields[0],
			f.EndTime.Format(DatetimeLayout),
		)
	}

	addField(
		&lines,
		cfg.OtherFields[1],
		f.Cost,
	)

	addField(
		&lines,
		cfg.OtherFields[2],
		f.Comment,
	)

	return strings.Join(lines, "\n")
}

func (f Trip) Show(cfg *lang.TripObjectCfg) string {
	var lines []string

	addStringField(
		&lines,
		cfg.MainFields[0],
		f.Name,
	)

	addStringField(
		&lines,
		cfg.MainFields[1],
		f.StartDate.Format(DateLayout),
	)

	addStringField(
		&lines,
		cfg.MainFields[2],
		f.EndDate.Format(DateLayout),
	)

	addField(
		&lines,
		cfg.OtherFields[0],
		f.Comment,
	)

	return strings.Join(lines, "\n")
}

func (f Ticket) Preview(cfg *lang.TripObjectCfg) string {
	return fmt.Sprintf("/tt%d %s-%s\n%s, %s - %s", f.ID, f.DeparturePoint, f.ArrivalPoint, f.Vehicle, f.DepartureTime.Format(DatetimeLayout), f.ArrivalTime.Format(DatetimeLayout))
}

func (f Booking) Preview(cfg *lang.TripObjectCfg) string {
	return fmt.Sprintf("/bb%d %s\n%s, %s - %s", f.ID, f.Place, f.City, f.CheckInDate.Format(DateLayout), f.CheckOutDate.Format(DateLayout))
}

func (f Event) Preview(cfg *lang.TripObjectCfg) string {
	res := fmt.Sprintf("/ev%d %s\n%s, %s", f.ID, f.Name, f.Location, f.StartTime.Format(DatetimeLayout))

	if f.EndTime != nil {
		return fmt.Sprintf("%s - %s", res, f.EndTime.Format(DatetimeLayout))
	} else {
		return res
	}
}

func (f Trip) Preview(cfg *lang.TripObjectCfg) string {
	return fmt.Sprintf("/tr%d %s, %s-%s", f.ID, f.Name, f.StartDate.Format(DateLayout), f.EndDate.Format(DateLayout))
}

func (f Ticket) GetTripID() *int64 {
	return f.TripID
}

func (f Booking) GetTripID() *int64 {
	return f.TripID
}

func (f Event) GetTripID() *int64 {
	return f.TripID
}

func (f Trip) GetTripID() *int64 {
	return nil
}
