package bot

import (
	"errors"
	"strings"
	"testing"
	"time"

	. "github.com/warmte/tripl/common"
)

func TestParseDatetime(t *testing.T) {
	t.Run("valid datetime", func(t *testing.T) {
		got, err := parseDatetime("25.12.2026 15:30", DatetimeLayout, "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Format(DatetimeLayout) != "25.12.2026 15:30" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("invalid datetime", func(t *testing.T) {
		_, err := parseDatetime("invalid", DatetimeLayout, "test")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestParseOptionalDatetime(t *testing.T) {
	got, err := parseOptionalDatetime("-", DatetimeLayout, "test")
	if err != nil || got != nil {
		t.Fatalf("got (%v, %v), want (nil, nil)", got, err)
	}

	got, err = parseOptionalDatetime("invalid", DatetimeLayout, "test")
	if err == nil || got != nil {
		t.Fatalf("got (%v, %v), want (nil, error)", got, err)
	}
}

func TestParseTicketMainFields(t *testing.T) {
	got, err := parseTicketMainFields(
		"Train\nMoscow\nSaint Petersburg\n25.12.2026 15:30\n26.12.2026 08:00",
		123,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.UserID != 123 {
		t.Errorf("UserID = %d, want 123", got.UserID)
	}
	if got.Vehicle == nil || *got.Vehicle != "Train" {
		t.Errorf("Vehicle = %v, want Train", got.Vehicle)
	}
	if got.DeparturePoint == nil || *got.DeparturePoint != "Moscow" {
		t.Errorf("DeparturePoint = %v, want Moscow", got.DeparturePoint)
	}
	if got.ArrivalPoint == nil || *got.ArrivalPoint != "Saint Petersburg" {
		t.Errorf("ArrivalPoint = %v, want Saint Petersburg", got.ArrivalPoint)
	}
	if got.DepartureTime == nil || got.DepartureTime.Format(DatetimeLayout) != "25.12.2026 15:30" {
		t.Errorf("unexpected DepartureTime: %v", got.DepartureTime)
	}
	if got.ArrivalTime == nil || got.ArrivalTime.Format(DatetimeLayout) != "26.12.2026 08:00" {
		t.Errorf("unexpected ArrivalTime: %v", got.ArrivalTime)
	}
}

func TestParseTicketMainFieldsErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  error
	}{
		{
			name:  "wrong number of fields",
			input: "Train\nMoscow",
			want:  ErrIncorrectNumberOfFields,
		},
		{
			name: "too long field",
			input: strings.Repeat("x", 257) +
				"\nMoscow\nPiter\n25.12.2026 15:30\n26.12.2026 08:00",
			want: ErrTooLongField,
		},
		{
			name:  "invalid departure time",
			input: "Train\nMoscow\nPiter\ninvalid\n26.12.2026 08:00",
			want:  ErrIncorrectDateTime,
		},
		{
			name:  "invalid arrival time",
			input: "Train\nMoscow\nPiter\n25.12.2026 15:30\ninvalid",
			want:  ErrIncorrectDateTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTicketMainFields(tt.input, 123)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestParseTicketOtherFields(t *testing.T) {
	main, err := parseTicketMainFields(
		"Train\nMoscow\nPiter\n25.12.2026 15:30\n26.12.2026 08:00",
		123,
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := parseTicketOtherFields(
		"1500\nRUB\nStation A\nStation B\n12A\nAB123\nWindow seat",
		&main,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Cost == nil || *got.Cost != "1500" {
		t.Errorf("Cost = %v, want 1500", got.Cost)
	}
	if got.Currency == nil || *got.Currency != "RUB" {
		t.Errorf("Currency = %v, want RUB", got.Currency)
	}
	if got.DepartureAddress == nil || *got.DepartureAddress != "Station A" {
		t.Errorf("DepartureAddress = %v", got.DepartureAddress)
	}
	if got.ArrivalAddress == nil || *got.ArrivalAddress != "Station B" {
		t.Errorf("ArrivalAddress = %v", got.ArrivalAddress)
	}
	if got.Seat == nil || *got.Seat != "12A" {
		t.Errorf("Seat = %v", got.Seat)
	}
	if got.Number == nil || *got.Number != "AB123" {
		t.Errorf("Number = %v", got.Number)
	}
	if got.Comment == nil || *got.Comment != "Window seat" {
		t.Errorf("Comment = %v", got.Comment)
	}
	if got.UserID != main.UserID {
		t.Errorf("UserID = %d, want %d", got.UserID, main.UserID)
	}
}

func TestParseTicketOtherFieldsOptionalValues(t *testing.T) {
	main := Ticket{UserID: 123}

	got, err := parseTicketOtherFields("-\n-\n-\n-\n-\n-\n-", &main)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Cost != nil || got.Currency != nil ||
		got.DepartureAddress != nil || got.ArrivalAddress != nil ||
		got.Seat != nil || got.Number != nil || got.Comment != nil {
		t.Fatal("expected all optional fields to be nil")
	}
}

func TestParseBookingMainFields(t *testing.T) {
	got, err := parseBookingMainFields("Paris\nHotel Central\n25.12.2026\n28.12.2026", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.UserID != 42 {
		t.Errorf("UserID = %d, want 42", got.UserID)
	}
	if got.City == nil || *got.City != "Paris" {
		t.Errorf("City = %v", got.City)
	}
	if got.Place == nil || *got.Place != "Hotel Central" {
		t.Errorf("Place = %v", got.Place)
	}
	if got.CheckInDate == nil || got.CheckInDate.Format(DateLayout) != "25.12.2026" {
		t.Errorf("CheckInDate = %v", got.CheckInDate)
	}
	if got.CheckOutDate == nil || got.CheckOutDate.Format(DateLayout) != "28.12.2026" {
		t.Errorf("CheckOutDate = %v", got.CheckOutDate)
	}
}

func TestParseBookingMainFieldsErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  error
	}{
		{"wrong number of fields", "Paris\nHotel", ErrIncorrectNumberOfFields},
		{"invalid check-in date", "Paris\nHotel\ninvalid\n28.12.2026", ErrIncorrectDate},
		{"invalid check-out date", "Paris\nHotel\n25.12.2026\ninvalid", ErrIncorrectDate},
		{
			"too long field",
			strings.Repeat("x", 257) + "\nHotel\n25.12.2026\n28.12.2026",
			ErrTooLongField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseBookingMainFields(tt.input, 42)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestParseBookingOtherFields(t *testing.T) {
	main, err := parseBookingMainFields("Paris\nHotel\n25.12.2026\n28.12.2026", 42)
	if err != nil {
		t.Fatal(err)
	}

	got, err := parseBookingOtherFields(
		"500\nEUR\nMain Street\n14:00\n11:00\nBK123\nBreakfast included",
		&main,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Cost == nil || *got.Cost != "500" {
		t.Errorf("Cost = %v", got.Cost)
	}
	if got.Currency == nil || *got.Currency != "EUR" {
		t.Errorf("Currency = %v", got.Currency)
	}
	if got.Address == nil || *got.Address != "Main Street" {
		t.Errorf("Address = %v", got.Address)
	}
	if got.CheckInTime == nil || got.CheckInTime.Format(TimeLayout) != "14:00" {
		t.Errorf("CheckInTime = %v", got.CheckInTime)
	}
	if got.CheckOutTime == nil || got.CheckOutTime.Format(TimeLayout) != "11:00" {
		t.Errorf("CheckOutTime = %v", got.CheckOutTime)
	}
	if got.Number == nil || *got.Number != "BK123" {
		t.Errorf("Number = %v", got.Number)
	}
	if got.Comment == nil || *got.Comment != "Breakfast included" {
		t.Errorf("Comment = %v", got.Comment)
	}
}

func TestParseBookingOtherFieldsErrors(t *testing.T) {
	main := Booking{}

	tests := []struct {
		name  string
		input string
		want  error
	}{
		{"wrong number of fields", "500\nEUR", ErrIncorrectNumberOfFields},
		{"invalid check-in time", "500\nEUR\nAddress\ninvalid\n11:00\nBK123\nComment", ErrIncorrectTime},
		{"invalid check-out time", "500\nEUR\nAddress\n14:00\ninvalid\nBK123\nComment", ErrIncorrectTime},
		{
			"too long field",
			"500\nEUR\n" + strings.Repeat("x", 257) + "\n14:00\n11:00\nBK123\nComment",
			ErrTooLongField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseBookingOtherFields(tt.input, &main)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestParseEventMainFields(t *testing.T) {
	got, err := parseEventMainFields("Concert\nCentral Park\n25.12.2026 19:00", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.UserID != 7 {
		t.Errorf("UserID = %d, want 7", got.UserID)
	}
	if got.Name == nil || *got.Name != "Concert" {
		t.Errorf("Name = %v", got.Name)
	}
	if got.Location == nil || *got.Location != "Central Park" {
		t.Errorf("Location = %v", got.Location)
	}
	if got.StartTime == nil || got.StartTime.Format(DatetimeLayout) != "25.12.2026 19:00" {
		t.Errorf("StartTime = %v", got.StartTime)
	}
}

func TestParseEventOtherFields(t *testing.T) {
	main, err := parseEventMainFields("Concert\nCentral Park\n25.12.2026 19:00", 7)
	if err != nil {
		t.Fatal(err)
	}

	got, err := parseEventOtherFields(
		"-\n20\nUSD\nLive music",
		&main,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.EndTime != nil {
		t.Errorf("EndTime = %v, want nil", got.EndTime)
	}
	if got.Cost == nil || *got.Cost != "20" {
		t.Errorf("Cost = %v", got.Cost)
	}
	if got.Currency == nil || *got.Currency != "USD" {
		t.Errorf("Currency = %v", got.Currency)
	}
	if got.Comment == nil || *got.Comment != "Live music" {
		t.Errorf("Comment = %v", got.Comment)
	}
}

func TestParseTripMainFields(t *testing.T) {
	got, err := parseTripMainFields("Summer trip\n25.12.2026\n28.12.2026", 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.UserID != 99 || got.Name != "Summer trip" {
		t.Errorf("unexpected trip: %+v", got)
	}
	if got.StartDate.Format(DateLayout) != "25.12.2026" {
		t.Errorf("StartDate = %v", got.StartDate)
	}
	if got.EndDate.Format(DateLayout) != "28.12.2026" {
		t.Errorf("EndDate = %v", got.EndDate)
	}
}

func TestParseTripOtherFields(t *testing.T) {
	main := Trip{
		UserID:    99,
		Name:      "Summer trip",
		StartDate: time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 12, 28, 0, 0, 0, 0, time.UTC),
	}

	got, err := parseTripOtherFields("Trip notes", &main)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Comment == nil || *got.Comment != "Trip notes" {
		t.Errorf("Comment = %v", got.Comment)
	}
	if got.Name != main.Name || got.UserID != main.UserID {
		t.Errorf("main fields were not preserved: %+v", got)
	}
}

func TestSelectMainFieldsParser(t *testing.T) {
	tests := []struct {
		name  string
		label Label
		input string
	}{
		{"ticket", TicketLabel, "Train\nMoscow\nPiter\n25.12.2026 15:30\n26.12.2026 08:00"},
		{"booking", BookingLabel, "Paris\nHotel\n25.12.2026\n28.12.2026"},
		{"event", EventLabel, "Concert\nPark\n25.12.2026 19:00"},
		{"trip", TripLabel, "Summer trip\n25.06.2026\n28.06.2026"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectMainFieldsParser(tt.input, tt.label, 1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected parsed object")
			}
		})
	}
}

func TestSelectMainFieldsParserInvalidLabel(t *testing.T) {
	_, err := selectMainFieldsParser("data", Label("unknown"), 1)
	if !errors.Is(err, ErrIncorrectInput) {
		t.Fatalf("error = %v, want %v", err, ErrIncorrectInput)
	}
}

func TestSelectOtherFieldsParserInvalidType(t *testing.T) {
	_, err := selectOtherFieldsParser("data", TicketLabel, Booking{})
	if !errors.Is(err, ErrInternalError) {
		t.Fatalf("error = %v, want %v", err, ErrInternalError)
	}
}

func TestSelectOtherFieldsParserInvalidLabel(t *testing.T) {
	_, err := selectOtherFieldsParser("data", Label("unknown"), Ticket{})
	if !errors.Is(err, ErrIncorrectInput) {
		t.Fatalf("error = %v, want %v", err, ErrIncorrectInput)
	}
}
