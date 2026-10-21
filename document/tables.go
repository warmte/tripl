package document

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
	"github.com/xuri/excelize/v2"
)

type exportSection struct {
	title   string
	headers []string
	rows    [][]string
}

func exportSections(
	tickets []common.Ticket,
	bookings []common.Booking,
	events []common.Event,
	cfg *lang.Cfg,
) []exportSection {
	str := func(v *string) string {
		if v == nil {
			return ""
		}
		return *v
	}
	date := func(v *time.Time, layout string) string {
		if v == nil {
			return ""
		}
		return v.Format(layout)
	}
	fields := func(main, other []string) []string {
		return append(append([]string{}, main...), other...)
	}

	sections := []exportSection{
		{
			title: cfg.TripObjects.Ticket.Many,
			headers: fields(
				cfg.TripObjects.Ticket.MainFields,
				cfg.TripObjects.Ticket.OtherFields,
			),
		},
		{
			title: cfg.TripObjects.Booking.Many,
			headers: fields(
				cfg.TripObjects.Booking.MainFields,
				cfg.TripObjects.Booking.OtherFields,
			),
		},
		{
			title: cfg.TripObjects.Event.Many,
			headers: fields(
				cfg.TripObjects.Event.MainFields,
				cfg.TripObjects.Event.OtherFields,
			),
		},
	}

	for _, t := range tickets {
		sections[0].rows = append(sections[0].rows, []string{
			str(t.Vehicle), str(t.DeparturePoint), str(t.ArrivalPoint),
			date(t.DepartureTime, common.DatetimeLayout),
			date(t.ArrivalTime, common.DatetimeLayout),
			str(t.Cost), str(t.Currency), str(t.DepartureAddress),
			str(t.ArrivalAddress), str(t.Seat), str(t.Number), str(t.Comment),
		})
	}

	for _, b := range bookings {
		sections[1].rows = append(sections[1].rows, []string{
			str(b.City), str(b.Place),
			date(b.CheckInDate, common.DateLayout),
			date(b.CheckOutDate, common.DateLayout),
			str(b.Cost), str(b.Currency), str(b.Address),
			date(b.CheckInTime, common.TimeLayout),
			date(b.CheckOutTime, common.TimeLayout),
			str(b.Number), str(b.Comment),
		})
	}

	for _, e := range events {
		sections[2].rows = append(sections[2].rows, []string{
			str(e.Name), str(e.Location),
			date(e.StartTime, common.DatetimeLayout),
			date(e.EndTime, common.DatetimeLayout),
			str(e.Cost), str(e.Currency), str(e.Comment),
		})
	}

	return sections
}

func CSVContent(
	tickets []common.Ticket,
	bookings []common.Booking,
	events []common.Event,
	cfg *lang.Cfg,
) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	for i, s := range exportSections(tickets, bookings, events, cfg) {
		if i > 0 {
			if err := w.Write([]string{}); err != nil {
				return nil, fmt.Errorf("write CSV separator: %w", err)
			}
		}
		rows := append([][]string{{s.title}, s.headers}, s.rows...)
		for _, row := range rows {
			if err := w.Write(row); err != nil {
				return nil, fmt.Errorf("write CSV section %q: %w", s.title, err)
			}
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush CSV: %w", err)
	}

	return append([]byte{0xEF, 0xBB, 0xBF}, buf.Bytes()...), nil
}

func XLSXContent(
	tickets []common.Ticket,
	bookings []common.Booking,
	events []common.Event,
	cfg *lang.Cfg,
) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	var sheet = cfg.TripObjects.Trip.One
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, fmt.Errorf("rename sheet: %w", err)
	}

	rowNum := 1
	for _, s := range exportSections(tickets, bookings, events, cfg) {
		rows := append([][]string{{s.title}, s.headers}, s.rows...)
		for _, row := range rows {
			cell, err := excelize.CoordinatesToCellName(1, rowNum)
			if err != nil {
				return nil, fmt.Errorf("get cell: %w", err)
			}

			values := make([]interface{}, len(row))
			for i, value := range row {
				values[i] = value
			}
			if err := f.SetSheetRow(sheet, cell, &values); err != nil {
				return nil, fmt.Errorf("write XLSX row %d: %w", rowNum, err)
			}
			rowNum++
		}
		rowNum++
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("write XLSX: %w", err)
	}
	return buf.Bytes(), nil
}
