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
