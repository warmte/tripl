package handler

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/document"
	"github.com/warmte/tripl/lang"
)

func (h CommonHandler[T]) GetPageByTripID(ctx context.Context, tripID int64, userID int64, page int) (Page[T], error) {
	if h.label == TripLabel {
		return Page[T]{}, ErrInternalError
	}
	return h.service.GetPageByTripID(ctx, tripID, userID, page)
}

func (h CommonHandler[T]) GetAllByTripID(ctx context.Context, tripID int64, userID int64) ([]T, error) {
	if h.label == TripLabel {
		return []T{}, ErrInternalError
	}
	return h.service.GetAllByTripID(ctx, tripID, userID)
}

func (h CommonHandler[T]) GetCostsByTripID(ctx context.Context, tripID int64, userID int64) (map[string]float64, error) {
	if h.label == TripLabel {
		return nil, ErrInternalError
	}
	return h.service.GetCostsByTripID(ctx, tripID, userID)
}

func (h CommonHandler[T]) CountByTripID(ctx context.Context, tripID int64, userID int64) (int64, error) {
	if h.label == TripLabel {
		return 0, ErrInternalError
	}
	return h.service.CountByTripID(ctx, tripID, userID)
}

func costsToString(costs map[string]float64) string {
	parts := make([]string, 0, len(costs))

	for currency, amount := range costs {
		parts = append(parts, fmt.Sprintf("%s %s", strconv.FormatFloat(amount, 'f', -1, 64), currency))
	}

	return strings.Join(parts, ", ")
}

func spentFor(costs map[string]float64, cfg *lang.Cfg, objectCfg *lang.TripObjectCfg) string {
	slices.Sorted(maps.Keys(costs))
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

func (h CallbackHandler) HandleTable(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	tripID int64,
	action CallbackAction,
	cfg *lang.Cfg,
) (tgbotapi.Chattable, error) {
	if callback == nil || callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	userID := callback.From.ID

	tickets, err := h.ticketHandler.GetAllByTripID(ctx, tripID, userID)
	if err != nil {
		return nil, fmt.Errorf("get trip tickets: %w", err)
	}

	bookings, err := h.bookingHandler.GetAllByTripID(ctx, tripID, userID)
	if err != nil {
		return nil, fmt.Errorf("get trip bookings: %w", err)
	}

	events, err := h.eventHandler.GetAllByTripID(ctx, tripID, userID)
	if err != nil {
		return nil, fmt.Errorf("get trip events: %w", err)
	}

	switch action {
	case TripCSV:
		content, err := document.CSVContent(tickets, bookings, events, cfg)
		if err != nil {
			return nil, fmt.Errorf("generate trip CSV: %w", err)
		}
		return tgbotapi.NewDocument(callback.Message.Chat.ID, tgbotapi.FileBytes{
			Name:  fmt.Sprintf("trip-%d.csv", tripID),
			Bytes: content,
		}), nil

	case TripXLSX:
		content, err := document.XLSXContent(tickets, bookings, events, cfg)
		if err != nil {
			return nil, fmt.Errorf("generate trip XLSX: %w", err)
		}
		return tgbotapi.NewDocument(callback.Message.Chat.ID, tgbotapi.FileBytes{
			Name:  fmt.Sprintf("trip-%d.xlsx", tripID),
			Bytes: content,
		}), nil
	default:
		return nil, fmt.Errorf("unknown table callback: %w", err)
	}

}
