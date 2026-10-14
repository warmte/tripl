package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
	"github.com/warmte/tripl/service"
)

func assertContainsTicketFields(t *testing.T, text string, ticket Ticket) {
	t.Helper()

	for _, field := range []*string{
		ticket.Vehicle,
		ticket.DeparturePoint,
		ticket.ArrivalPoint,
		ticket.Cost,
		ticket.Currency,
		ticket.DepartureAddress,
		ticket.ArrivalAddress,
		ticket.Seat,
		ticket.Number,
		ticket.Comment,
	} {
		if field != nil && *field != "-" {
			require.Contains(t, text, lang.EscapeMarkdownV2(*field))
		}
	}

	for _, field := range []*time.Time{ticket.DepartureTime, ticket.ArrivalTime} {
		if field != nil {
			require.Contains(t, text, lang.EscapeMarkdownV2(field.Format(DatetimeLayout)))
		}
	}
}

func assertContainsBookingFields(t *testing.T, text string, booking Booking) {
	t.Helper()

	for _, field := range []*string{
		booking.City,
		booking.Place,
		booking.Cost,
		booking.Currency,
		booking.Address,
		booking.Number,
		booking.Comment,
	} {
		if field != nil && *field != "-" {
			require.Contains(t, text, lang.EscapeMarkdownV2(*field))
		}
	}

	for _, field := range []*time.Time{booking.CheckInDate, booking.CheckOutDate} {
		if field != nil {
			require.Contains(t, text, lang.EscapeMarkdownV2(field.Format(DateLayout)))
		}
	}

	for _, field := range []*time.Time{booking.CheckInTime, booking.CheckOutTime} {
		if field != nil {
			require.Contains(t, text, lang.EscapeMarkdownV2(field.Format(TimeLayout)))
		}
	}
}

func assertContainsEventFields(t *testing.T, text string, event Event) {
	t.Helper()

	for _, field := range []*string{
		event.Name,
		event.Location,
		event.Cost,
		event.Currency,
		event.Comment,
	} {
		if field != nil && *field != "-" {
			require.Contains(t, text, lang.EscapeMarkdownV2(*field))
		}
	}

	for _, field := range []*time.Time{event.StartTime, event.EndTime} {
		if field != nil {
			require.Contains(t, text, lang.EscapeMarkdownV2(field.Format(DatetimeLayout)))
		}
	}
}

func assertContainsTripFields(t *testing.T, text string, trip Trip) {
	t.Helper()

	require.Contains(t, text, lang.EscapeMarkdownV2(trip.Name))
	require.Contains(t, text, lang.EscapeMarkdownV2(trip.StartDate.Format(DateLayout)))
	require.Contains(t, text, lang.EscapeMarkdownV2(trip.EndDate.Format(DateLayout)))

	if trip.Comment != nil && *trip.Comment != "-" {
		require.Contains(t, text, lang.EscapeMarkdownV2(*trip.Comment))
	}
}

func assertKeyboardHasButtons(
	t *testing.T,
	markup *tgbotapi.InlineKeyboardMarkup,
	expected []string,
	unexpected []string,
) {
	t.Helper()

	require.NotNil(t, markup)

	var texts []string
	for _, row := range markup.InlineKeyboard {
		for _, button := range row {
			texts = append(texts, button.Text)
		}
	}

	for _, text := range expected {
		require.Contains(t, texts, text)
	}

	for _, text := range unexpected {
		require.NotContains(t, texts, text)
	}
}

func testTripID() *int64 {
	id := int64(202)
	return &id
}

func stringPtr(value string) *string {
	return &value
}

func TestShowSingle_Ticket(t *testing.T) {
	ctx := context.Background()
	const userID int64 = 1
	const objectID int64 = 101

	ticket := NewTestTicket()
	ticket.ID = objectID
	ticket.UserID = userID
	ticket.TripID = nil

	svc := new(service.MockService[Ticket])
	svc.On("GetByID", mock.Anything, userID, objectID).Return(&ticket, nil).Once()

	cfg := lang.Get(lang.LanguageEn)

	h := CommonHandler[Ticket]{
		label:   TicketLabel,
		service: svc,
	}

	msg, err := h.ShowSingle(ctx, userID, objectID, cfg)

	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, userID, int64(msg.ChatID))
	require.Equal(t, tgbotapi.ModeMarkdownV2, msg.ParseMode)
	require.NotEmpty(t, msg.Text)
	require.Contains(t, msg.Text, lang.EscapeMarkdownV2(cfg.TripObjects.Ticket.One))
	require.NotContains(t, msg.Text, "\n-\n")

	assertContainsTicketFields(t, msg.Text, ticket)
	assertKeyboardHasButtons(t, msg.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup),
		[]string{cfg.Buttons.Attach, cfg.Buttons.Edit, cfg.Buttons.Delete},
		[]string{cfg.Buttons.Detach, cfg.Buttons.Stats},
	)

	svc.AssertExpectations(t)
}

func TestShowSingle_GetByIDError(t *testing.T) {
	ctx := context.Background()
	const userID int64 = 1
	const objectID int64 = 101

	expectedErr := errors.New("database error")

	svc := new(service.MockService[Ticket])
	svc.On("GetByID", mock.Anything, userID, objectID).Return((*Ticket)(nil), expectedErr).Once()

	h := CommonHandler[Ticket]{
		label:   TicketLabel,
		service: svc,
	}

	msg, err := h.ShowSingle(ctx, userID, objectID, &lang.Cfg{})

	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "get ticket")
	require.Nil(t, msg)

	svc.AssertExpectations(t)
}

func TestShowSingle_TicketWithTripID(t *testing.T) {
	ctx := context.Background()
	const userID int64 = 1
	const objectID int64 = 101

	ticket := NewTestTicket()
	ticket.ID = objectID
	ticket.UserID = userID
	ticket.TripID = testTripID()

	trip := NewTestTrip()

	svc := new(service.MockService[Ticket])
	svc.On("GetByID", mock.Anything, userID, objectID).Return(&ticket, nil).Once()

	tripSvc := new(service.MockService[Trip])
	tripSvc.On("GetByID", mock.Anything, userID, *ticket.TripID).Return(&trip, nil).Once()

	cfg := lang.Get(lang.LanguageEn)

	h := CommonHandler[Ticket]{
		label:       TicketLabel,
		service:     svc,
		tripService: tripSvc,
	}

	msg, err := h.ShowSingle(ctx, userID, objectID, cfg)

	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Contains(t, msg.Text, lang.EscapeMarkdownV2(trip.Preview(&cfg.TripObjects.Trip)))

	assertKeyboardHasButtons(t, msg.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup),
		[]string{cfg.Buttons.Detach, cfg.Buttons.Edit, cfg.Buttons.Delete},
		[]string{cfg.Buttons.Attach, cfg.Buttons.Stats},
	)

	svc.AssertExpectations(t)
	tripSvc.AssertExpectations(t)
}

func TestShowSingle_Booking(t *testing.T) {
	ctx := context.Background()
	const userID int64 = 1
	const objectID int64 = 102

	booking := Booking{
		ID:           objectID,
		UserID:       userID,
		TripID:       nil,
		City:         stringPtr("Helsinki"),
		Place:        stringPtr("Test hotel"),
		CheckInDate:  timePtr(time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)),
		CheckOutDate: timePtr(time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)),
		Cost:         stringPtr("300"),
		Currency:     stringPtr("EUR"),
		Address:      stringPtr("Test street"),
		CheckInTime:  timePtr(time.Date(2026, 1, 1, 14, 0, 0, 0, time.UTC)),
		CheckOutTime: timePtr(time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)),
		Number:       stringPtr("B-102"),
		Comment:      stringPtr("Breakfast included"),
	}

	svc := new(service.MockService[Booking])
	svc.On("GetByID", mock.Anything, userID, objectID).Return(&booking, nil).Once()

	cfg := lang.Get(lang.LanguageEn)

	h := CommonHandler[Booking]{
		label:   BookingLabel,
		service: svc,
	}

	msg, err := h.ShowSingle(ctx, userID, objectID, cfg)

	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Contains(t, msg.Text, lang.EscapeMarkdownV2(cfg.TripObjects.Booking.One))
	assertContainsBookingFields(t, msg.Text, booking)

	assertKeyboardHasButtons(t, msg.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup),
		[]string{cfg.Buttons.Attach, cfg.Buttons.Edit, cfg.Buttons.Delete},
		[]string{cfg.Buttons.Detach, cfg.Buttons.Stats},
	)

	svc.AssertExpectations(t)
}

func TestShowSingle_Event(t *testing.T) {
	ctx := context.Background()
	const userID int64 = 1
	const objectID int64 = 103

	event := Event{
		ID:        objectID,
		UserID:    userID,
		TripID:    nil,
		Name:      stringPtr("Museum visit"),
		Location:  stringPtr("Helsinki"),
		StartTime: timePtr(time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)),
		EndTime:   timePtr(time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)),
		Cost:      stringPtr("20"),
		Currency:  stringPtr("EUR"),
		Comment:   stringPtr("Guided tour"),
	}

	svc := new(service.MockService[Event])
	svc.On("GetByID", mock.Anything, userID, objectID).Return(&event, nil).Once()

	cfg := lang.Get(lang.LanguageEn)

	h := CommonHandler[Event]{
		label:   EventLabel,
		service: svc,
	}

	msg, err := h.ShowSingle(ctx, userID, objectID, cfg)

	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Contains(t, msg.Text, lang.EscapeMarkdownV2(cfg.TripObjects.Event.One))
	assertContainsEventFields(t, msg.Text, event)

	assertKeyboardHasButtons(t, msg.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup),
		[]string{cfg.Buttons.Attach, cfg.Buttons.Edit, cfg.Buttons.Delete},
		[]string{cfg.Buttons.Detach, cfg.Buttons.Stats},
	)

	svc.AssertExpectations(t)
}

func TestShowSingle_Trip(t *testing.T) {
	ctx := context.Background()
	const userID int64 = 1
	const objectID int64 = 202

	trip := NewTestTrip()
	trip.ID = objectID
	trip.UserID = userID

	svc := new(service.MockService[Trip])
	svc.On("GetByID", mock.Anything, userID, objectID).Return(&trip, nil).Once()

	cfg := lang.Get(lang.LanguageEn)

	h := CommonHandler[Trip]{
		label:   TripLabel,
		service: svc,
	}

	msg, err := h.ShowSingle(ctx, userID, objectID, cfg)

	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Contains(t, msg.Text, lang.EscapeMarkdownV2(cfg.TripObjects.Trip.One))
	assertContainsTripFields(t, msg.Text, trip)

	assertKeyboardHasButtons(t, msg.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup),
		[]string{
			cfg.Buttons.Stats,
			cfg.Buttons.Edit,
			cfg.Buttons.Delete,
			cfg.TripObjects.Ticket.Many,
			cfg.TripObjects.Booking.Many,
			cfg.TripObjects.Event.Many,
		},
		[]string{cfg.Buttons.Attach, cfg.Buttons.Detach},
	)

	svc.AssertExpectations(t)
}

func timePtr(value time.Time) *time.Time {
	return &value
}
