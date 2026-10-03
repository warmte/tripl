package handler

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
)

type CallbackHandler struct {
	ticketHandler  *TicketHandler
	bookingHandler *BookingHandler
	eventHandler   *EventHandler
	tripHandler    *TripHandler
}

func NewCallbackHandler(
	ticketHandler *TicketHandler,
	bookingHandler *BookingHandler,
	eventHandler *EventHandler,
	tripHandler *TripHandler,
) *CallbackHandler {
	return &CallbackHandler{
		ticketHandler:  ticketHandler,
		bookingHandler: bookingHandler,
		eventHandler:   eventHandler,
		tripHandler:    tripHandler,
	}
}

func (h *CallbackHandler) Handle(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	if callback == nil {
		return nil, nil
	}

	if strings.Count(callback.Data, ":") == 3 {
		extended, err := parseExtendedCallback(callback.Data)
		if err != nil {
			return nil, fmt.Errorf("parse extended callback: %w", err)
		}

		switch extended.Action {
		case TripTickets:
			r, err := h.ticketHandler.HandleListForTrip(ctx, callback, extended.V1, int(extended.V2), cfg)
			return []tgbotapi.Chattable{r}, err
		case TripBookings:
			r, err := h.bookingHandler.HandleListForTrip(ctx, callback, extended.V1, int(extended.V2), cfg)
			return []tgbotapi.Chattable{r}, err
		case TripEvents:
			r, err := h.eventHandler.HandleListForTrip(ctx, callback, extended.V1, int(extended.V2), cfg)
			return []tgbotapi.Chattable{r}, err
		default:
			switch extended.Label {
			case TicketLabel:
				return h.ticketHandler.SelectExtendedAction(extended, ctx, callback, cfg)
			case BookingLabel:
				return h.bookingHandler.SelectExtendedAction(extended, ctx, callback, cfg)
			case EventLabel:
				return h.eventHandler.SelectExtendedAction(extended, ctx, callback, cfg)
			case TripLabel:
				return h.tripHandler.SelectExtendedAction(extended, ctx, callback, cfg)
			}
		}
	}

	parsedCallback, err := parseCallback(callback.Data)
	if err != nil {
		return nil, fmt.Errorf("parse callback: %w", err)
	}

	unsupportedCallbackError := fmt.Errorf(
		"unsupported callback: %s:%s:%d",
		parsedCallback.Label,
		parsedCallback.Action,
		parsedCallback.V1,
	)

	switch parsedCallback.Action {
	case TripStats:
		r, err := h.HandleTripStats(ctx, callback, parsedCallback.V1, cfg)
		return []tgbotapi.Chattable{r}, err
	default:
		switch parsedCallback.Label {
		case TicketLabel:
			return h.ticketHandler.SelectAction(parsedCallback, ctx, callback, cfg)
		case BookingLabel:
			return h.bookingHandler.SelectAction(parsedCallback, ctx, callback, cfg)
		case EventLabel:
			return h.eventHandler.SelectAction(parsedCallback, ctx, callback, cfg)

		case TripLabel:
			switch parsedCallback.Action {
			case AttachSingleActual, AttachSingleArchived, DetachSingle:
				return nil, unsupportedCallbackError
			default:
				return h.tripHandler.SelectAction(parsedCallback, ctx, callback, cfg)
			}
		default:
			return nil, unsupportedCallbackError
		}
	}
}

func costsToString(costs map[string]float64) string {
	parts := make([]string, 0, len(costs))

	for currency, amount := range costs {
		parts = append(parts, fmt.Sprintf("%g %s", amount, currency))
	}

	return strings.Join(parts, ", ")
}

func spentFor(costs map[string]float64, cfg *lang.Cfg, objectCfg *lang.TripObjectCfg) string {
	return fmt.Sprintf("%s: %s\n", fmt.Sprintf(cfg.SpentFor, strings.ToLower(objectCfg.Many)), costsToString(costs))
}

func (h CallbackHandler) HandleTripStats(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	tripID int64,
	cfg *lang.Cfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	var text strings.Builder

	ticketsCount, err := h.ticketHandler.CountByTripID(ctx, tripID, callback.From.ID)
	if err == nil {
		text.WriteString(fmt.Sprintf("%s: %d\n", cfg.TripObjects.Ticket.Many, ticketsCount))
	}
	bookingsCount, err := h.bookingHandler.CountByTripID(ctx, tripID, callback.From.ID)
	if err == nil {
		text.WriteString(fmt.Sprintf("%s: %d\n", cfg.TripObjects.Booking.Many, bookingsCount))
	}
	eventsCount, err := h.eventHandler.CountByTripID(ctx, tripID, callback.From.ID)
	if err == nil {
		text.WriteString(fmt.Sprintf("%s: %d\n", cfg.TripObjects.Event.Many, eventsCount))
	}

	totalCosts := make(map[string]float64)
	ticketsCosts, err := h.ticketHandler.GetCostsByTripID(ctx, tripID, callback.From.ID)
	if err == nil && len(ticketsCosts) > 0 {
		text.WriteString(spentFor(ticketsCosts, cfg, &cfg.TripObjects.Ticket))
		for currency, amount := range ticketsCosts {
			totalCosts[currency] += amount
		}
	}
	bookingsCosts, err := h.bookingHandler.GetCostsByTripID(ctx, tripID, callback.From.ID)
	if err == nil && len(bookingsCosts) > 0 {
		text.WriteString(spentFor(bookingsCosts, cfg, &cfg.TripObjects.Booking))
		for currency, amount := range bookingsCosts {
			totalCosts[currency] += amount
		}
	}
	eventsCosts, err := h.eventHandler.GetCostsByTripID(ctx, tripID, callback.From.ID)
	if err == nil && len(eventsCosts) > 0 {
		text.WriteString(spentFor(eventsCosts, cfg, &cfg.TripObjects.Event))
		for currency, amount := range eventsCosts {
			totalCosts[currency] += amount
		}
	}
	if len(totalCosts) > 0 {
		text.WriteString(fmt.Sprintf("%s: %s", cfg.SpentTotal, costsToString(totalCosts)))
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				button(cfg.Buttons.Back, fmt.Sprintf("%s:%s:%d", TripLabel, GetSingle, tripID)),
			},
		},
	}

	r := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text.String(),
		keyboard,
	)

	return &r, nil
}
