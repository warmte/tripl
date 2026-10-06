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
)

const (
	defaultAIURL      = "http://localhost:8080/v1/chat/completions"
	aiRequestTimeout  = 60 * 5 * time.Second
	maxAIResponseSize = 1 * 1024 * 1024
	aiMaxTokens       = 1024
	aiTemperature     = 0.2
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
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
	Cost      *string `json:"cost"`
	Currency  *string `json:"currency"`
}

func (s *AIService) ExtractTicket(ctx context.Context, userID int64, text string) (*Ticket, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, ErrAIEmptyResponse
	}

	response, err := s.ask(ctx, ticketSystemPrompt, text)
	if err != nil {
		return nil, err
	}

	response = cleanJSONResponse(response)
	return s.parseTicket(response, userID)
}

func (s *AIService) ExtractBooking(ctx context.Context, userID int64, text string) (*Booking, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, ErrAIEmptyResponse
	}

	response, err := s.ask(ctx, bookingSystemPrompt, text)
	if err != nil {
		return nil, err
	}

	response = cleanJSONResponse(response)
	return s.parseBooking(response, userID)
}

func (s *AIService) ExtractEvent(ctx context.Context, userID int64, text string) (*Event, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, ErrAIEmptyResponse
	}

	response, err := s.ask(ctx, bookingSystemPrompt, text)
	if err != nil {
		return nil, err
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

func (s *AIService) parseTicket(response string, userID int64) (*Ticket, error) {
	var data ticketAIResponse

	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return nil, fmt.Errorf("%w: ticket: %v", ErrAIInvalidJSON, err)
	}

	departureTime := parseOptionalDateTime(data.DepartureDate, data.DepartureTime, "departure")

	arrivalTime := parseOptionalDateTime(data.ArrivalDate, data.ArrivalTime, "arrival")

	return &Ticket{
		UserID:           userID,
		Vehicle:          data.Vehicle,
		DeparturePoint:   data.DeparturePoint,
		ArrivalPoint:     data.ArrivalPoint,
		DepartureTime:    departureTime,
		ArrivalTime:      arrivalTime,
		Cost:             data.Cost,
		Currency:         data.Currency,
		DepartureAddress: data.DepartureAddress,
		ArrivalAddress:   data.ArrivalAddress,
		Seat:             data.Seat,
		Number:           data.Number,
	}, nil
}

func (s *AIService) parseBooking(response string, userID int64) (*Booking, error) {
	var data bookingAIResponse

	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return nil, fmt.Errorf("%w: booking: %v", ErrAIInvalidJSON, err)
	}

	checkInDate := parseOptionalDate(
		data.CheckInDate,
		"check_in_date",
	)

	checkOutDate := parseOptionalDate(
		data.CheckOutDate,
		"check_out_date",
	)

	checkInTime := parseOptionalTime(
		data.CheckInTime,
		"check_in_time",
	)

	checkOutTime := parseOptionalTime(
		data.CheckOutTime,
		"check_out_time",
	)

	return &Booking{
		UserID:       userID,
		City:         data.City,
		Place:        data.Place,
		CheckInDate:  checkInDate,
		CheckOutDate: checkOutDate,
		Cost:         data.Cost,
		Currency:     data.Currency,
		Address:      data.Address,
		CheckInTime:  checkInTime,
		CheckOutTime: checkOutTime,
		Number:       data.Number,
	}, nil
}

func (s *AIService) parseEvent(response string, userID int64) (*Event, error) {
	var data eventAIResponse

	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return nil, fmt.Errorf("%w: event: %v", ErrAIInvalidJSON, err)
	}

	startTime := parseOptionalTime(
		data.StartTime,
		"start_time",
	)

	endTime := parseOptionalTime(
		data.EndTime,
		"end_time",
	)

	return &Event{
		UserID:    userID,
		Name:      data.Name,
		Location:  data.Location,
		StartTime: startTime,
		EndTime:   endTime,
		Cost:      data.Cost,
		Currency:  data.Currency,
	}, nil
}

func parseOptionalDate(value *string, field string) *time.Time {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}

	result, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return nil
	}

	return &result
}

func parseOptionalTime(value *string, field string) *time.Time {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}

	result, err := time.Parse(
		time.RFC3339,
		strings.TrimSpace(*value),
	)
	if err != nil {
		return nil
	}
	return &result
}

func parseOptionalDateTime(date *string, value *string, field string) *time.Time {
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

const ticketSystemPrompt = `
Extract one travel ticket from the document.

Return ONLY valid JSON, no Markdown or explanations. Never invent data. Missing nullable fields = null.

{
  "vehicle": "string",
  "departure_point": "city only or null",
  "arrival_point": "city only or null",
  "departure_time": "RFC3339 datetime or null",
  "arrival_time": "RFC3339 datetime or null",
  "cost": "string or null",
  "currency": "string or null",
  "departure_address": "string or null",
  "arrival_address": "string or null",
  "seat": "string or null",
  "number": "string or null"
}

Vehicle: add flight number / train number to name of vehicle if available

Location:
- departure_point, arrival_point contain cities names only.
- Put stations, airports, terminals, addresses, platforms, gates and other location details in the corresponding *_address field.
- Airplane: address include airport and terminal.
- Train: address include railway stations and other departure/arrival details. seat = "carriage <carriage>, seat <seat>".
- Other transport: same city-only rule; put location details in addresses.

Cost:
- Numeric value only: 1250, 1250.5, 1250.50.
- No currency symbols, codes, spaces or words.

Datetime:
- Return departure_time and arrival_time as RFC3339 when date and time are available.
- Combine a date and time stated separately in the document: arrival with arrival, departure with departure.
- Never invent a date or timezone.
- If the document contains only a time without a date or date without a time, return null.

"number" = eticket number, not flight number, not booking number
`

const bookingSystemPrompt = `
You extract one hotel/accommodation booking from the supplied document.

Return ONLY valid JSON.
Do not use Markdown.
Do not add explanations.
Do not invent information.

JSON fields:
{
  "city": "string" or null,
  "place": "string" or null,
  "check_in_date": "YYYY-MM-DD or null",
  "check_out_date": "YYYY-MM-DD or null",
  "cost": "string or null",
  "currency": "string or null",
  "address": "string or null",
  "check_in_time": "RFC3339 datetime or null",
  "check_out_time": "RFC3339 datetime or null",
  "number": "string or null"
}

"cost" must be a number containing only the numeric value.
Do not include currency symbols, currency codes, spaces, words, or other characters.
Examples of valid values: 1250, 1250.50, 99.99.
Examples of invalid values: "1250 RUB", "1 250", "₽1250", "1250€".
If the cost is not present, return null.
"number" = booking number.

Never invent missing information.
`

const eventSystemPrompt = `
You extract one event from the supplied travel document.

Return ONLY valid JSON.
Do not use Markdown.
Do not add explanations.
Do not invent information.

JSON fields:
{
  "name": "string or null",
  "location": "string or null",
  "start_time": "RFC3339 datetime or null",
  "end_time": "RFC3339 datetime or null",
  "cost": "string or null",
  "currency": "string or null"
}

"cost" must be a number containing only the numeric value.
Do not include currency symbols, currency codes, spaces, words, or other characters.
Examples of valid values: 1250, 1250.50, 99.99.
Examples of invalid values: "1250 RUB", "1 250", "₽1250", "1250€".
If the cost is not present, return null.

Never invent missing information.
`
