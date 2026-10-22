package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
)

const (
	defaultAIURL      = "http://llama-server:8080/v1/chat/completions"
	aiRequestTimeout  = 6 * 60 * time.Second
	maxAIResponseSize = 1 * 1024 * 1024
	aiMaxTokens       = 600
	aiTemperature     = 0.1
)

var (
	ErrAIEmptyResponse    = errors.New("AI returned empty response")
	ErrAIInvalidJSON      = errors.New("AI returned invalid JSON")
	ErrAIUnsupportedLabel = errors.New("unsupported label")
	ErrAIInvalidResult    = errors.New("AI returned invalid object")
)

type AIService struct {
	client  *http.Client
	baseURL string
	model   string
}

func NewAIService(baseURL string, model string) *AIService {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultAIURL
	}

	return &AIService{
		client: &http.Client{
			Timeout: aiRequestTimeout,
		},
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
	}
}

type aiChatRequest struct {
	Model              string            `json:"model"`
	Messages           []aiChatMessage   `json:"messages"`
	Temperature        float64           `json:"temperature"`
	MaxTokens          int               `json:"max_tokens"`
	ResponseFormat     *aiResponseFormat `json:"response_format,omitempty"`
	ChatTemplateKwargs map[string]bool   `json:"chat_template_kwargs,omitempty"`
}

type aiResponseFormat struct {
	Type string `json:"type"`
}

type aiChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type aiChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type ticketAIResponse struct {
	Vehicle          *string `json:"vehicle"`
	DeparturePoint   *string `json:"departure_point"`
	ArrivalPoint     *string `json:"arrival_point"`
	DepartureTime    *string `json:"departure_time"`
	ArrivalTime      *string `json:"arrival_time"`
	DepartureDate    *string `json:"departure_date"`
	ArrivalDate      *string `json:"arrival_date"`
	Cost             *string `json:"cost"`
	Currency         *string `json:"currency"`
	DepartureAddress *string `json:"departure_address"`
	ArrivalAddress   *string `json:"arrival_address"`
	Seat             *string `json:"seat"`
	Number           *string `json:"number"`
}

type ticketsAIResponse struct {
	Tickets []ticketAIResponse `json:"tickets"`
}

type bookingAIResponse struct {
	City         *string `json:"city"`
	Place        *string `json:"place"`
	CheckInDate  *string `json:"check_in_date"`
	CheckOutDate *string `json:"check_out_date"`
	Cost         *string `json:"cost"`
	Currency     *string `json:"currency"`
	Address      *string `json:"address"`
	CheckInTime  *string `json:"check_in_time"`
	CheckOutTime *string `json:"check_out_time"`
	Number       *string `json:"number"`
}

type eventAIResponse struct {
	Name      *string `json:"name"`
	Location  *string `json:"location"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
	Cost      *string `json:"cost"`
	Currency  *string `json:"currency"`
}

func withLangPrompt(l lang.Language, text string) string {
	langPrompt := russianPrompt
	if l == lang.LanguageEn {
		langPrompt = englishPrompt
	}
	return fmt.Sprintf("%s\n\n%s", langPrompt, text)
}

func (s *AIService) ExtractTicket(ctx context.Context, userID int64, text string, l lang.Language) ([]Ticket, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, ErrAIEmptyResponse
	}

	response, err := s.ask(ctx, withLangPrompt(l, ticketSystemPrompt), text)
	if err != nil {
		return nil, err
	}

	response = cleanJSONResponse(response)
	return s.parseTickets(response, userID)
}

func (s *AIService) ExtractBooking(ctx context.Context, userID int64, text string, l lang.Language) (Booking, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Booking{}, ErrAIEmptyResponse
	}

	response, err := s.ask(ctx, withLangPrompt(l, bookingSystemPrompt), text)
	if err != nil {
		return Booking{}, err
	}

	response = cleanJSONResponse(response)
	return s.parseBooking(response, userID)
}

func (s *AIService) ExtractEvent(ctx context.Context, userID int64, text string, l lang.Language) (Event, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Event{}, ErrAIEmptyResponse
	}

	response, err := s.ask(ctx, withLangPrompt(l, eventSystemPrompt), text)
	if err != nil {
		return Event{}, err
	}

	response = cleanJSONResponse(response)
	return s.parseEvent(response, userID)
}

func (s *AIService) ask(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	request := aiChatRequest{
		Model: s.model,
		Messages: []aiChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: aiTemperature,
		MaxTokens:   aiMaxTokens,
		ResponseFormat: &aiResponseFormat{
			Type: "json_object",
		},
		ChatTemplateKwargs: map[string]bool{
			"enable_thinking": false,
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("marshal AI request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create AI request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send AI request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(
			io.LimitReader(resp.Body, maxAIResponseSize),
		)

		return "", fmt.Errorf(
			"AI returned status %s: %s",
			resp.Status,
			strings.TrimSpace(string(responseBody)),
		)
	}

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxAIResponseSize))
	if err != nil {
		return "", fmt.Errorf("read AI response: %w", err)
	}

	var response aiChatResponse

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("decode AI response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", errors.New("AI response contains no choices")
	}

	result := strings.TrimSpace(response.Choices[0].Message.Content)
	if result == "" {
		return "", ErrAIEmptyResponse
	}

	return result, nil
}

func (s *AIService) parseTickets(response string, userID int64) ([]Ticket, error) {
	var data ticketsAIResponse

	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return nil, fmt.Errorf("%w: tickets: %v", ErrAIInvalidJSON, err)
	}

	if len(data.Tickets) == 0 {
		return nil, fmt.Errorf("%w: no tickets found", ErrAIInvalidJSON)
	}

	if len(data.Tickets) > 3 {
		return nil, fmt.Errorf("%w: too many tickets: %d", ErrAIInvalidJSON, len(data.Tickets))
	}

	tickets := make([]Ticket, 0, len(data.Tickets))

	for _, data := range data.Tickets {
		cost, currency := checkedCostCurrency(data.Cost, data.Currency)
		tickets = append(tickets, Ticket{
			UserID:           userID,
			Vehicle:          Normalize(data.Vehicle),
			DeparturePoint:   Normalize(data.DeparturePoint),
			ArrivalPoint:     Normalize(data.ArrivalPoint),
			DepartureTime:    parseOptionalDateTime(data.DepartureDate, data.DepartureTime),
			ArrivalTime:      parseOptionalDateTime(data.ArrivalDate, data.ArrivalTime),
			Cost:             cost,
			Currency:         currency,
			DepartureAddress: Normalize(data.DepartureAddress),
			ArrivalAddress:   Normalize(data.ArrivalAddress),
			Seat:             Normalize(data.Seat),
			Number:           Normalize(data.Number),
		})
	}

	return tickets, nil
}

func (s *AIService) parseBooking(response string, userID int64) (Booking, error) {
	var data bookingAIResponse

	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return Booking{}, fmt.Errorf("%w: booking: %v", ErrAIInvalidJSON, err)
	}

	checkInDate := parseOptionalDate(data.CheckInDate)

	checkOutDate := parseOptionalDate(data.CheckOutDate)

	checkInTime := parseOptionalTime(data.CheckInTime)

	checkOutTime := parseOptionalTime(data.CheckOutTime)

	cost, currency := checkedCostCurrency(data.Cost, data.Currency)

	return Booking{
		UserID:       userID,
		City:         Normalize(data.City),
		Place:        Normalize(data.Place),
		CheckInDate:  checkInDate,
		CheckOutDate: checkOutDate,
		Cost:         cost,
		Currency:     currency,
		Address:      Normalize(data.Address),
		CheckInTime:  checkInTime,
		CheckOutTime: checkOutTime,
		Number:       Normalize(data.Number),
	}, nil
}

func (s *AIService) parseEvent(response string, userID int64) (Event, error) {
	var data eventAIResponse

	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return Event{}, fmt.Errorf("%w: event: %v", ErrAIInvalidJSON, err)
	}

	startTime := parseOptionalDateTime(
		data.StartDate,
		data.StartTime,
	)

	endTime := parseOptionalDateTime(
		data.EndDate,
		data.EndTime,
	)

	cost, currency := checkedCostCurrency(data.Cost, data.Currency)

	return Event{
		UserID:    userID,
		Name:      Normalize(data.Name),
		Location:  Normalize(data.Location),
		StartTime: startTime,
		EndTime:   endTime,
		Cost:      cost,
		Currency:  currency,
	}, nil
}

func parseOptionalDate(value *string) *time.Time {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}

	result, err := time.Parse("02.01.2006", *value)
	if err != nil {
		return nil
	}

	return &result
}

func parseOptionalTime(value *string) *time.Time {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}

	result, err := time.Parse("15:04", strings.TrimSpace(*value))
	if err != nil {
		return nil
	}
	return &result
}

func parseOptionalDateTime(date *string, value *string) *time.Time {
	if date == nil || value == nil ||
		strings.TrimSpace(*date) == "" ||
		strings.TrimSpace(*value) == "" {
		return nil
	}

	t, err := time.Parse(
		"02.01.2006 15:04",
		strings.TrimSpace(*date)+" "+strings.TrimSpace(*value),
	)
	if err != nil {
		return nil
	}

	return &t
}

func cleanJSONResponse(value string) string {
	value = strings.TrimSpace(value)

	if strings.HasPrefix(value, "```json") {
		value = strings.TrimPrefix(value, "```json")
		value = strings.TrimSpace(value)
	}

	if strings.HasPrefix(value, "```") {
		value = strings.TrimPrefix(value, "```")
		value = strings.TrimSpace(value)
	}

	if strings.HasSuffix(value, "```") {
		value = strings.TrimSuffix(value, "```")
		value = strings.TrimSpace(value)
	}

	return value
}

func checkedCostCurrency(cost *string, currency *string) (*string, *string) {
	if (cost != nil && (*(cost) == "0" || *(cost) == "null" || *(cost) == "")) || cost == nil {
		return nil, nil
	}
	nCost := NormalizeCost(*cost)
	nCurrency := Normalize(currency)
	if nCurrency != nil {
		r := NormalizeCurrency(*nCurrency)
		nCurrency = &r
	}
	return &nCost, nCurrency
}

const englishPrompt = "Use English for all textual values: airplane, train, bus, carriage, seat, etc."
const russianPrompt = "Use Russian for all textual values: самолет, поезд, автобус, вагон, место, etc. instead of airplane, train, bus, carriage, seat, etc."

const ticketSystemPrompt = `
Extract 1–3 tickets/travel segments from the document.

Return ONLY valid JSON: {"tickets":[...]}. No Markdown or explanations. Never invent data.
Each journey segment must be a separate ticket, ordered chronologically.

Each ticket must contain:
- vehicle: string or null
- departure_point: string or null
- arrival_point: string or null
- departure_date: DD.MM.YYYY or null
- departure_time: HH:MM or null
- arrival_date: DD.MM.YYYY or null
- arrival_time: HH:MM or null
- cost: string or null
- currency: string or null
- departure_address: string or null
- arrival_address: string or null
- seat: string or null
- number: ticket/e-ticket number or null

Rules:
- Use JSON null (without quotes) for missing values, never the string "null".
- vehicle = transport type + flight/train/bus number, e.g. "Airplane F0101", "Train 560A", "Bus 43". Translate to language specified before.
- departure_point/arrival_point = city name only. For airports, append the 3-letter IATA code only if explicitly shown in the document, e.g. "Ufa, UFA". Never guess codes or use airport names as cities.
- *_address = airport/station and any shown terminal, platform or gate, e.g. "Ufa Airport, terminal 1".
- seat = actual passenger seat only. For trains use "carriage X, seat Y". Never use terminal, gate, luggage or other numbers as a seat.
- Dates: DD.MM.YYYY. Times: HH:MM. Match them to the correct segment; never invent them.
- cost = amount only, without currency. Currency = stated currency. If there are only one cost for multiple segments, place it in first segments, others = null.
- number = ticket/e-ticket number, not a flight, train, bus or booking number. If multiple segments belong to the same e-ticket, use the same ticket number for each.
- Extract every available field independently. Missing information in one field must not make other fields null.
`

const bookingSystemPrompt = `
Extract one hotel/accommodation booking from the document.

Return ONLY valid JSON. No Markdown or explanations. Never invent data.
Use null (without quotes) for missing values, never the string "null".

Fields:
{
  "city": "string or null",
  "place": "string or null",
  "check_in_date": "DD.MM.YYYY or null",
  "check_out_date": "DD.MM.YYYY or null",
  "cost": "string or null",
  "currency": "string or null",
  "address": "string or null",
  "check_in_time": "HH:MM or null",
  "check_out_time": "HH:MM  or null",
  "number": "booking number or null"
}

Rules:
- city = city where the accommodation is located.
- place = hotel, apartment or accommodation name.
- address = full address if shown.
- Dates: DD.MM.YYYY.
- Times: HH:M.
- cost = numeric amount only, without currency symbols or codes.
- currency = stated currency code or name.
- number = booking/reservation number only.
- Extract each field independently. Missing information in one field must not make other fields null.
`

const eventSystemPrompt = `
Extract one event from the document.

Return ONLY valid JSON. No Markdown or explanations. Never invent data.
Use null (without quotes) for missing values, never the string "null".

Fields:
{
  "name": "string or null",
  "location": "string or null",
	"start_date": "DD.MM.YYYY or null",
  "end_date": "DD.MM.YYYY or null",
  "start_time": "HH:MM or null",
  "end_time": "HH:MM or null",
  "cost": "string or null",
  "currency": "string or null"
}

Rules:
- name = event name or title.
- location = event venue and address, if shown.
- Dates: DD.MM.YYYY. Times: HH:MM. If one-day event, return start_date = end_date.
- cost = numeric amount only, without currency symbols or codes.
- currency = stated currency code or name.
- Extract each field independently. Missing information in one field must not make other fields null.
`
