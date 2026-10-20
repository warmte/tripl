package handler

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/document"
	"github.com/warmte/tripl/lang"
)

// Ticket

func (h TicketHandler) HandleMakeICS(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {

	object, err := h.service.GetByID(ctx, callback.From.ID, objectID)
	if err != nil {
		return nil, err
	}

	if object.DepartureTime == nil || object.ArrivalTime == nil || object.ArrivalTime.Before(*object.DepartureTime) {
		return nil, ErrDatetimeRestraint
	}

	keyboard, err := makeICSRegionsKeyboard(h.label, objectID, ICSChooseRegion, 0, cfg)
	if err != nil {
		return nil, err
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		cfg.ICS.ChooseStart,
		keyboard,
	)
	return []tgbotapi.Chattable{edit}, nil
}

func (h TicketHandler) HandleICSStartRegion(
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	regionIndex int,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	keyboard, err := makeICSZonesKeyboard(h.label, objectID, ICSChooseStart, regionIndex, 0, cfg)
	if err != nil {
		return nil, err
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		cfg.ICS.ChooseStart,
		keyboard,
	)
	return []tgbotapi.Chattable{edit}, nil
}

func (h TicketHandler) HandleICSStartZone(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	startCode int64,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	keyboard, err := makeICSRegionsKeyboard(h.label, objectID, ICSChooseEndRegion, startCode, cfg)
	if err != nil {
		return nil, err
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		cfg.ICS.ChooseEnd,
		keyboard,
	)
	return []tgbotapi.Chattable{edit}, nil
}

func (h TicketHandler) HandleICSEndRegion(
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	encoded int64,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	startCode, endRegion, err := decodeICSEndRegion(encoded)
	if err != nil {
		return nil, err
	}

	keyboard, err := makeICSZonesKeyboard(h.label, objectID, ICSChooseEnd, endRegion, startCode, cfg)
	if err != nil {
		return nil, err
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		cfg.ICS.ChooseEnd,
		keyboard,
	)
	return []tgbotapi.Chattable{edit}, nil
}

func (h TicketHandler) HandleICSEndZone(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	encoded int64,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	startCode, endCode, err := decodeICSZones(encoded)
	if err != nil {
		return nil, err
	}
	startRegion, startZoneIndex, err := decodeICSZone(startCode)
	if err != nil {
		return nil, err
	}
	endRegion, endZoneIndex, err := decodeICSZone(endCode)
	if err != nil {
		return nil, err
	}

	startZone, err := icsZoneByIndex(cfg, startRegion, startZoneIndex)
	if err != nil {
		return nil, err
	}
	endZone, err := icsZoneByIndex(cfg, endRegion, endZoneIndex)
	if err != nil {
		return nil, err
	}

	ics, err := h.sendICS(ctx, callback, objectID, startZone.ID, endZone.ID, cfg)
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(callback.From.ID, cfg.ICS.Success)

	edit, err := h.HandleGetSingle(ctx, callback, objectID, cfg)
	if err != nil {
		return nil, err
	}

	return []tgbotapi.Chattable{ics, msg, edit}, nil
}

// Booking

func (h BookingHandler) HandleMakeICS(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	object, err := h.service.GetByID(ctx, callback.From.ID, objectID)
	if err != nil {
		return nil, err
	}

	if object.CheckInDate == nil || object.CheckOutDate == nil || object.CheckOutDate.Before(*object.CheckInDate) {
		return nil, ErrDatetimeRestraint
	}

	ics, err := h.sendICS(ctx, callback, objectID, "", "", cfg)
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(callback.From.ID, cfg.ICS.Success)

	edit, err := h.HandleGetSingle(ctx, callback, objectID, cfg)
	if err != nil {
		return nil, err
	}
	return []tgbotapi.Chattable{ics, msg, edit}, nil
}

// Event

func (h EventHandler) HandleMakeICS(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {

	object, err := h.service.GetByID(ctx, callback.From.ID, objectID)
	if err != nil {
		return nil, err
	}

	if object.StartTime == nil || (object.EndTime != nil && object.EndTime.Before(*object.StartTime)) {
		return nil, ErrDatetimeRestraint
	}

	text := cfg.ICS.ChooseZone
	keyboard, err := makeICSRegionsKeyboard(h.label, objectID, ICSChooseRegion, 0, cfg)
	if err != nil {
		return nil, err
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	return []tgbotapi.Chattable{edit}, nil
}

func (h EventHandler) HandleICSRegion(
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	regionIndex int,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	keyboard, err := makeICSZonesKeyboard(h.label, objectID, ICSChooseStart, regionIndex, 0, cfg)
	if err != nil {
		return nil, err
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		cfg.ICS.ChooseZone,
		keyboard,
	)
	return []tgbotapi.Chattable{edit}, nil
}

func (h EventHandler) HandleICSZone(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	startCode int64,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	startRegion, startZoneIndex, err := decodeICSZone(startCode)
	if err != nil {
		return nil, err
	}

	startZone, err := icsZoneByIndex(cfg, startRegion, startZoneIndex)
	if err != nil {
		return nil, err
	}

	ics, err := h.sendICS(ctx, callback, objectID, startZone.ID, startZone.ID, cfg)

	if err != nil {
		return nil, err
	}

	edit, err := h.HandleGetSingle(ctx, callback, objectID, cfg)
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(callback.From.ID, cfg.ICS.Success)
	return []tgbotapi.Chattable{ics, msg, edit}, nil
}

// utils

const (
	icsZoneBase = 100
	icsCodeBase = 1000
)

func encodeICSZone(regionIndex, zoneIndex int) (int64, error) {
	if regionIndex < 0 || regionIndex >= 10 {
		return 0, fmt.Errorf("invalid ICS region index: %d", regionIndex)
	}
	if zoneIndex < 0 || zoneIndex >= 100 {
		return 0, fmt.Errorf("invalid ICS zone index: %d", zoneIndex)
	}

	return int64(regionIndex*icsZoneBase + zoneIndex), nil
}

func decodeICSZone(code int64) (regionIndex, zoneIndex int, err error) {
	if code < 0 || code >= 1000 {
		return 0, 0, fmt.Errorf("invalid ICS zone code: %d", code)
	}

	return int(code / icsZoneBase), int(code % icsZoneBase), nil
}

func encodeICSZones(startCode, endCode int64) (int64, error) {
	if startCode < 0 || startCode >= 1000 || endCode < 0 || endCode >= 1000 {
		return 0, fmt.Errorf("invalid ICS zone codes: %d, %d", startCode, endCode)
	}

	return startCode*icsCodeBase + endCode, nil
}

func decodeICSZones(code int64) (startCode, endCode int64, err error) {
	if code < 0 || code >= 1_000_000 {
		return 0, 0, fmt.Errorf("invalid ICS zones code: %d", code)
	}

	return code / icsCodeBase, code % icsCodeBase, nil
}

func encodeICSEndRegion(startCode int64, endRegion int) (int64, error) {
	if startCode < 0 || startCode >= 1000 {
		return 0, fmt.Errorf("invalid start zone code: %d", startCode)
	}
	if endRegion < 0 || endRegion >= 10 {
		return 0, fmt.Errorf("invalid end region index: %d", endRegion)
	}

	return startCode*icsCodeBase + int64(endRegion*icsZoneBase), nil
}

func decodeICSEndRegion(code int64) (startCode int64, endRegion int, err error) {
	if code < 0 || code >= 1_000_000 {
		return 0, 0, fmt.Errorf("invalid end-region code: %d", code)
	}

	startCode = code / icsCodeBase
	endRegion = int(code%icsCodeBase) / icsZoneBase
	return startCode, endRegion, nil
}

func icsZoneByIndex(cfg *lang.Cfg, regionIndex, zoneIndex int) (lang.ICSZone, error) {
	if regionIndex < 0 || regionIndex >= len(cfg.ICS.Regions) {
		return lang.ICSZone{}, fmt.Errorf("invalid region index: %d", regionIndex)
	}

	zones := cfg.ICS.Regions[regionIndex].Zones
	if zoneIndex < 0 || zoneIndex >= len(zones) {
		return lang.ICSZone{}, fmt.Errorf("invalid zone index: %d", zoneIndex)
	}

	return zones[zoneIndex], nil
}

func icsHasEndTime(obj TripObjectView) bool {
	switch v := any(obj).(type) {
	case Ticket:
		return v.ArrivalTime != nil
	case Event:
		return v.EndTime != nil
	default:
		return false
	}
}

func makeICSRegionsKeyboard(
	label Label,
	objectID int64,
	action CallbackAction,
	startCode int64,
	cfg *lang.Cfg,
) (tgbotapi.InlineKeyboardMarkup, error) {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(cfg.ICS.Regions)+1)

	for regionIndex, region := range cfg.ICS.Regions {
		callbackValue := int64(regionIndex)
		var err error
		if action == ICSChooseEndRegion {
			callbackValue, err = encodeICSEndRegion(startCode, regionIndex)
			if err != nil {
				return tgbotapi.InlineKeyboardMarkup{}, err
			}
		}
		buttonData := MakeExtendedCallback(label, action, objectID, callbackValue)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button(region.Name, buttonData)))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		button(cfg.Buttons.Cancel, MakeCallback(label, GetSingle, objectID)),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...), nil
}

func makeICSZonesKeyboard(
	label Label,
	objectID int64,
	action CallbackAction,
	regionIndex int,
	startZone int64,
	cfg *lang.Cfg,
) (tgbotapi.InlineKeyboardMarkup, error) {
	if regionIndex < 0 || regionIndex >= len(cfg.ICS.Regions) {
		return tgbotapi.InlineKeyboardMarkup{}, fmt.Errorf("invalid region index: %d", regionIndex)
	}

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(cfg.ICS.Regions[regionIndex].Zones)+1)

	for zoneIndex, zone := range cfg.ICS.Regions[regionIndex].Zones {
		var encoded int64

		switch action {
		case ICSChooseStart:
			encoded, _ = encodeICSZone(regionIndex, zoneIndex)
		case ICSChooseEnd:
			startCode := startZone
			endCode, err := encodeICSZone(regionIndex, zoneIndex)
			if err != nil {
				return tgbotapi.InlineKeyboardMarkup{}, err
			}
			encoded, err = encodeICSZones(startCode, endCode)
			if err != nil {
				return tgbotapi.InlineKeyboardMarkup{}, err
			}
		default:
			return tgbotapi.InlineKeyboardMarkup{}, fmt.Errorf("unsupported ICS zone action: %s", action)
		}

		buttonData := MakeExtendedCallback(label, action, objectID, encoded)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button(zone.Name, buttonData)))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		button(cfg.Buttons.Cancel, MakeCallback(label, GetSingle, objectID)),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...), nil
}

func (h *CommonHandler[T]) sendICS(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	startZone, endZone string,
	cfg *lang.Cfg,
) (tgbotapi.Chattable, error) {
	object, err := h.service.GetByID(ctx, callback.From.ID, objectID)
	if err != nil {
		return nil, err
	}

	content, err := document.ICSContent(*object, startZone, endZone, cfg)
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("%s-%d.ics", strings.ToLower(string(h.label)), objectID)
	documentConfig := tgbotapi.NewDocument(
		callback.Message.Chat.ID,
		tgbotapi.FileBytes{Name: fileName, Bytes: content},
	)

	return documentConfig, nil
}
