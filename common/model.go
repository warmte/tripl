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

const (
	LocationName = "Europe/Moscow"
)

const PageSize int = 5

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
	GetID() int64
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

const (
	TicketCmd  string = "tt"
	BookingCmd string = "bk"
	EventCmd   string = "ev"
	TripCmd    string = "tr"
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
	Vehicle          *string
	DeparturePoint   *string
	ArrivalPoint     *string
	DepartureTime    *time.Time
	ArrivalTime      *time.Time
	Cost             *string
	Currency         *string
	DepartureAddress *string
	ArrivalAddress   *string
	Seat             *string
	Number           *string
	Comment          *string
}

type Booking struct {
	ID           int64
	UserID       int64
	TripID       *int64
	City         *string
	Place        *string
	CheckInDate  *time.Time
	CheckOutDate *time.Time
	Cost         *string
	Currency     *string
	Address      *string
	CheckInTime  *time.Time
	CheckOutTime *time.Time
	Number       *string
	Comment      *string
}

type Event struct {
	ID        int64
	UserID    int64
	TripID    *int64
	Name      *string
	Location  *string
	StartTime *time.Time
	EndTime   *time.Time
	Cost      *string
	Currency  *string
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

type User struct {
	ID        int64
	Language  string
	CreatedAt time.Time
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

	addField(
		&lines,
		cfg.MainFields[0],
		f.Vehicle,
	)

	addField(
		&lines,
		cfg.MainFields[1],
		f.DeparturePoint,
	)

	addField(
		&lines,
		cfg.MainFields[2],
		f.ArrivalPoint,
	)

	if f.DepartureTime != nil {
		addStringField(
			&lines,
			cfg.MainFields[3],
			f.DepartureTime.Format(DatetimeLayout),
		)
	}

	if f.ArrivalTime != nil {
		addStringField(
			&lines,
			cfg.MainFields[4],
			f.ArrivalTime.Format(DatetimeLayout),
		)
	}

	addField(
		&lines,
		cfg.OtherFields[0],
		f.Cost,
	)

	addField(
		&lines,
		cfg.OtherFields[1],
		f.Currency,
	)

	addField(
		&lines,
		cfg.OtherFields[2],
		f.DepartureAddress,
	)

	addField(
		&lines,
		cfg.OtherFields[3],
		f.ArrivalAddress,
	)

	addField(
		&lines,
		cfg.OtherFields[4],
		f.Seat,
	)

	addField(
		&lines,
		cfg.OtherFields[5],
		f.Number,
	)

	addField(
		&lines,
		cfg.OtherFields[6],
		f.Comment,
	)

	return strings.Join(lines, "\n")
}

func (f Booking) Show(cfg *lang.TripObjectCfg) string {
	var lines []string

	addField(
		&lines,
		cfg.MainFields[0],
		f.City,
	)

	addField(
		&lines,
		cfg.MainFields[1],
		f.Place,
	)

	if f.CheckInDate != nil {
		addStringField(
			&lines,
			cfg.MainFields[2],
			f.CheckInDate.Format(DateLayout),
		)
	}

	if f.CheckOutDate != nil {
		addStringField(
			&lines,
			cfg.MainFields[3],
			f.CheckOutDate.Format(DateLayout),
		)
	}

	addField(
		&lines,
		cfg.OtherFields[0],
		f.Cost,
	)

	addField(
		&lines,
		cfg.OtherFields[1],
		f.Currency,
	)

	addField(
		&lines,
		cfg.OtherFields[2],
		f.Address,
	)

	if f.CheckInTime != nil {
		addStringField(
			&lines,
			cfg.OtherFields[3],
			f.CheckInTime.Format(TimeLayout),
		)
	}

	if f.CheckOutTime != nil {
		addStringField(
			&lines,
			cfg.OtherFields[4],
			f.CheckOutTime.Format(TimeLayout),
		)
	}

	addField(
		&lines,
		cfg.OtherFields[5],
		f.Number,
	)

	addField(
		&lines,
		cfg.OtherFields[6],
		f.Comment,
	)

	return strings.Join(lines, "\n")
}

func (f Event) Show(cfg *lang.TripObjectCfg) string {
	var lines []string

	addField(
		&lines,
		cfg.MainFields[0],
		f.Name,
	)

	addField(
		&lines,
		cfg.MainFields[1],
		f.Location,
	)

	if f.StartTime != nil {
		addStringField(
			&lines,
			cfg.MainFields[2],
			f.StartTime.Format(DatetimeLayout),
		)
	}

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
		f.Currency,
	)

	addField(
		&lines,
		cfg.OtherFields[3],
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
	return fmt.Sprintf("/%s%d %s - %s\n%s%s", TicketCmd, f.ID, optStr(f.DeparturePoint), optStr(f.ArrivalPoint), optStr(f.Vehicle), optDateStr(f.DepartureTime, f.ArrivalTime, DatetimeLayout))
}

func (f Booking) Preview(cfg *lang.TripObjectCfg) string {
	return fmt.Sprintf("/%s%d %s\n%s%s", BookingCmd, f.ID, optStr(f.Place), optStr(f.City), optDateStr(f.CheckInDate, f.CheckOutDate, DateLayout))
}

func (f Event) Preview(cfg *lang.TripObjectCfg) string {
	return fmt.Sprintf("/%s%d %s\n%s%s", EventCmd, f.ID, optStr(f.Name), optStr(f.Location), optDateStr(f.StartTime, f.EndTime, DatetimeLayout))
}

func (f Trip) Preview(cfg *lang.TripObjectCfg) string {
	return fmt.Sprintf("/%s%d %s, %s - %s", TripCmd, f.ID, f.Name, f.StartDate.Format(DateLayout), f.EndDate.Format(DateLayout))
}

func (f Ticket) GetID() int64 {
	return f.ID
}

func (f Booking) GetID() int64 {
	return f.ID
}

func (f Event) GetID() int64 {
	return f.ID
}

func (f Trip) GetID() int64 {
	return f.ID
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
