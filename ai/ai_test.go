package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
)

func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain JSON",
			in:   `{"tickets":[]}`,
			want: `{"tickets":[]}`,
		},
		{
			name: "JSON code block",
			in:   "```json\n{\"tickets\":[]}\n```",
			want: `{"tickets":[]}`,
		},
		{
			name: "generic code block",
			in:   "```\n{\"tickets\":[]}\n```",
			want: `{"tickets":[]}`,
		},
		{
			name: "whitespace",
			in:   "  {\"tickets\":[]}  ",
			want: `{"tickets":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanJSONResponse(tt.in); got != tt.want {
				t.Fatalf("cleanJSONResponse() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRmNull(t *testing.T) {

	tests := []struct {
		name string
		in   *string
		want *string
	}{
		{name: "nil", in: nil, want: nil},
		{name: "null string", in: stringPtr("null"), want: nil},
		{name: "empty string", in: stringPtr(""), want: nil},
		{name: "normal string", in: stringPtr("100"), want: stringPtr("100")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.in)
			if !equalStringPointers(got, tt.want) {
				t.Fatalf("rmNull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckedCostCurrency(t *testing.T) {
	tests := []struct {
		name         string
		cost         *string
		currency     *string
		wantCost     *string
		wantCurrency *string
	}{
		{
			name:         "valid cost and currency",
			cost:         stringPtr("150"),
			currency:     stringPtr("RUB"),
			wantCost:     stringPtr("150"),
			wantCurrency: stringPtr("RUB"),
		},
		{
			name:         "zero cost",
			cost:         stringPtr("0"),
			currency:     stringPtr("RUB"),
			wantCost:     nil,
			wantCurrency: nil,
		},
		{
			name:         "null cost",
			cost:         stringPtr("null"),
			currency:     stringPtr("RUB"),
			wantCost:     nil,
			wantCurrency: nil,
		},
		{
			name:         "empty cost",
			cost:         stringPtr(""),
			currency:     stringPtr("RUB"),
			wantCost:     nil,
			wantCurrency: nil,
		},
		{
			name:         "nil cost",
			cost:         nil,
			currency:     stringPtr("RUB"),
			wantCost:     nil,
			wantCurrency: nil,
		},
		{
			name:         "nil currency",
			cost:         stringPtr("150"),
			currency:     nil,
			wantCost:     stringPtr("150"),
			wantCurrency: nil,
		},
		{
			name:         "null currency",
			cost:         stringPtr("150"),
			currency:     stringPtr("null"),
			wantCost:     stringPtr("150"),
			wantCurrency: nil,
		},
		{
			name:         "both nil",
			cost:         nil,
			currency:     nil,
			wantCost:     nil,
			wantCurrency: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCost, gotCurrency := checkedCostCurrency(tt.cost, tt.currency)

			if !equalStringPointers(gotCost, tt.wantCost) {
				t.Errorf("cost = %v, want %v", gotCost, tt.wantCost)
			}

			if !equalStringPointers(gotCurrency, tt.wantCurrency) {
				t.Errorf("currency = %v, want %v", gotCurrency, tt.wantCurrency)
			}
		})
	}
}

func TestParseOptionalDate(t *testing.T) {
	tests := []struct {
		name  string
		value *string
		want  bool
	}{
		{name: "nil", value: nil, want: false},
		{name: "empty", value: stringPtr(" "), want: false},
		{name: "invalid", value: stringPtr("2026-04-08"), want: false},
		{name: "valid", value: stringPtr("08.04.2026"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOptionalDate(tt.value)
			if (got != nil) != tt.want {
				t.Fatalf("parseOptionalDate() != expected result")
			}
			if got != nil && got.Format("02.01.2006") != *tt.value {
				t.Fatalf("unexpected date: %v", got)
			}
		})
	}
}

func TestParseOptionalTime(t *testing.T) {
	invalid := "2026-04-08T15:00:00Z"

	tests := []struct {
		name  string
		value *string
		want  bool
	}{
		{name: "nil", value: nil, want: false},
		{name: "empty", value: stringPtr(" "), want: false},
		{name: "invalid", value: &invalid, want: false},
		{name: "valid format", value: stringPtr("15:00"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOptionalTime(tt.value)
			if (got != nil) != tt.want {
				t.Fatalf("parseOptionalTime() != expected result")
			}
		})
	}
}

func TestParseOptionalDateTime(t *testing.T) {
	date := "08.04.2026"
	value := "15:00"

	got := parseOptionalDateTime(&date, &value)
	if got == nil {
		t.Fatal("expected a parsed date and time")
	}

	if got.Format("02.01.2006 15:04") != "08.04.2026 15:00" {
		t.Fatalf("unexpected datetime: %v", got)
	}

	invalid := "invalid"
	if got := parseOptionalDateTime(&date, &invalid); got != nil {
		t.Fatalf("expected nil for invalid datetime, got %v", got)
	}
}

func TestParseTickets(t *testing.T) {
	s := NewAIService("", "test-model")

	response := `{
		"tickets": [{
			"vehicle": "Train 560A",
			"departure_point": "Saint Petersburg",
			"arrival_point": "Samara",
			"departure_date": "08.04.2026",
			"departure_time": "15:00",
			"arrival_date": "09.04.2026",
			"arrival_time": "12:00",
			"cost": "6000",
			"currency": "RUB",
			"seat": "carriage 6, seat 15",
			"number": "ABC123"
		}]
	}`

	got, err := s.parseTickets(response, 42)
	if err != nil {
		t.Fatalf("parseTickets() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d tickets, want 1", len(got))
	}

	ticket := got[0]

	if ticket.UserID != 42 {
		t.Errorf("UserID = %v, want 42", ticket.UserID)
	}
	if ticket.Vehicle == nil || *ticket.Vehicle != "Train 560A" {
		t.Errorf("unexpected vehicle: %v", ticket.Vehicle)
	}
	if ticket.Cost == nil || *ticket.Cost != "6000" {
		t.Errorf("unexpected cost: %v", ticket.Cost)
	}
	if ticket.Currency == nil || *ticket.Currency != "RUB" {
		t.Errorf("unexpected currency: %v", ticket.Currency)
	}
	if ticket.DepartureTime == nil {
		t.Error("expected DepartureTime to be set")
	}
	if ticket.ArrivalTime == nil {
		t.Error("expected ArrivalTime to be set")
	}
}

func TestParseTicketsRejectsInvalidResults(t *testing.T) {
	s := NewAIService("", "test-model")

	tests := []struct {
		name string
		json string
	}{
		{name: "invalid JSON", json: `{`},
		{name: "empty list", json: `{"tickets":[]}`},
		{
			name: "too many tickets",
			json: `{"tickets":[{}, {}, {}, {}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.parseTickets(tt.json, 42)
			if !errors.Is(err, ErrAIInvalidJSON) {
				t.Fatalf("error = %v, want ErrAIInvalidJSON", err)
			}
		})
	}
}

func TestParseBooking(t *testing.T) {
	s := NewAIService("", "test-model")

	response := `{
		"city": "Seoul",
		"place": "Best Western Incheon",
		"check_in_date": "08.04.2026",
		"check_out_date": "14.04.2026",
		"cost": "500000",
		"currency": "KRW",
		"address": "Incheon, South Korea",
		"check_in_time": "15:00",
		"check_out_time": "12:00",
		"number": "BW12345"
	}`

	got, err := s.parseBooking(response, 42)
	if err != nil {
		t.Fatalf("parseBooking() error = %v", err)
	}

	if got.UserID != 42 {
		t.Errorf("UserID = %v, want 42", got.UserID)
	}
	if got.City == nil || *got.City != "Seoul" {
		t.Errorf("unexpected city: %v", got.City)
	}
	if got.Place == nil || *got.Place != "Best Western Incheon" {
		t.Errorf("unexpected place: %v", got.Place)
	}
	if got.CheckInDate == nil || got.CheckOutDate == nil {
		t.Error("expected check-in and check-out dates")
	}
	if got.Cost == nil || *got.Cost != "500000" {
		t.Errorf("unexpected cost: %v", got.Cost)
	}
}

func TestParseEvent(t *testing.T) {
	s := NewAIService("", "test-model")

	response := `{
		"name": "WCS Nights",
		"location": "Azimut",
		"start_date": "23.07.2026",
		"end_date": "27.07.2026",
		"start_time": "15:00",
		"end_time": "06:00",
		"cost": "20000",
		"currency": "RUB"
	}`

	got, err := s.parseEvent(response, 42)
	if err != nil {
		t.Fatalf("parseEvent() error = %v", err)
	}

	if got.UserID != 42 {
		t.Errorf("UserID = %v, want 42", got.UserID)
	}
	if got.Name == nil || *got.Name != "WCS Nights" {
		t.Errorf("unexpected name: %v", got.Name)
	}
	if got.StartTime == nil || got.EndTime == nil {
		t.Error("expected start and end times")
	}
	if got.Cost == nil || *got.Cost != "20000" {
		t.Errorf("unexpected cost: %v", got.Cost)
	}
}

func TestAsk(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantErr    bool
		wantResult string
	}{
		{
			name:       "success",
			status:     http.StatusOK,
			body:       `{"choices":[{"message":{"content":"{\"tickets\":[{}]}"}}]}`,
			wantResult: `{"tickets":[{}]}`,
		},
		{
			name:    "HTTP error",
			status:  http.StatusBadGateway,
			body:    "upstream error",
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			status:  http.StatusOK,
			body:    `{`,
			wantErr: true,
		},
		{
			name:    "no choices",
			status:  http.StatusOK,
			body:    `{"choices":[]}`,
			wantErr: true,
		},
		{
			name:    "empty content",
			status:  http.StatusOK,
			body:    `{"choices":[{"message":{"content":" "}}]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
				}

				var request aiChatRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decode request: %v", err)
				} else {
					if request.Model != "test-model" {
						t.Errorf("model = %q, want test-model", request.Model)
					}
					if len(request.Messages) != 2 {
						t.Errorf("got %d messages, want 2", len(request.Messages))
					}
				}

				w.WriteHeader(tt.status)
				_, _ = fmt.Fprint(w, tt.body)
			}))
			defer server.Close()

			s := NewAIService(server.URL, "test-model")
			got, err := s.ask(context.Background(), "system prompt", "user prompt")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ask() error = %v", err)
			}
			if got != tt.wantResult {
				t.Fatalf("ask() = %q, want %q", got, tt.wantResult)
			}
		})
	}
}

func TestExtractTicketRejectsEmptyInput(t *testing.T) {
	s := NewAIService("", "test-model")

	_, err := s.ExtractTicket(context.Background(), 42, " \n\t ", lang.LanguageEn)
	if !errors.Is(err, ErrAIEmptyResponse) {
		t.Fatalf("error = %v, want ErrAIEmptyResponse", err)
	}
}

func TestExtractBookingRejectsEmptyInput(t *testing.T) {
	s := NewAIService("", "test-model")

	_, err := s.ExtractBooking(context.Background(), 42, "", lang.LanguageEn)
	if !errors.Is(err, ErrAIEmptyResponse) {
		t.Fatalf("error = %v, want ErrAIEmptyResponse", err)
	}
}

func TestExtractEventRejectsEmptyInput(t *testing.T) {
	s := NewAIService("", "test-model")

	_, err := s.ExtractEvent(context.Background(), 42, "   ", lang.LanguageEn)
	if !errors.Is(err, ErrAIEmptyResponse) {
		t.Fatalf("error = %v, want ErrAIEmptyResponse", err)
	}
}

func stringPtr(value string) *string {
	return &value
}

func equalStringPointers(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
