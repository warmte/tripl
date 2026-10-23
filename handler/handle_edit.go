package handler

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
)

type editField struct {
	ID   int
	Name string
}

type editContext struct {
	Label    Label
	ObjectID int64
	FieldID  int
}

func makeEditCallback(label Label, objectID int64, fieldID int) string {
	return fmt.Sprintf(
		"%s:%s:%d:%d",
		label,
		EditField,
		objectID,
		fieldID,
	)
}

func ParseEditContext(text string, cfg *lang.Cfg) (editContext, error) {
	var editContextRe = regexp.MustCompile(
		fmt.Sprintf(
			`(?m)^%s\s*\n\[(\d+)\]\s*[^,\r\n]+,\s*(%s|%s|%s|%s)\s+№(\d+)`,
			regexp.QuoteMeta(cfg.OnEdit),
			regexp.QuoteMeta(cfg.TripObjects.Ticket.One),
			regexp.QuoteMeta(cfg.TripObjects.Booking.One),
			regexp.QuoteMeta(cfg.TripObjects.Event.One),
			regexp.QuoteMeta(cfg.TripObjects.Trip.One),
		),
	)

	matches := editContextRe.FindStringSubmatch(text)
	if len(matches) != 4 {
		return editContext{}, fmt.Errorf("invalid edit context")
	}

	fieldID, err := strconv.Atoi(matches[1])
	if err != nil || fieldID < 0 {
		return editContext{}, fmt.Errorf("invalid field id: %q", matches[1])
	}

	objectID, err := strconv.ParseInt(matches[3], 10, 64)
	if err != nil || objectID <= 0 {
		return editContext{}, fmt.Errorf("invalid object id: %q", matches[3])
	}

	var label Label

	switch matches[2] {
	case cfg.TripObjects.Ticket.One:
		label = TicketLabel
	case cfg.TripObjects.Booking.One:
		label = BookingLabel
	case cfg.TripObjects.Event.One:
		label = EventLabel
	case cfg.TripObjects.Trip.One:
		label = TripLabel
	default:
		return editContext{}, fmt.Errorf(
			"unknown object label: %q",
			matches[2],
		)
	}

	return editContext{
		Label:    label,
		ObjectID: objectID,
		FieldID:  fieldID - 1,
	}, nil
}

func editObjectName(cfg *lang.Cfg, label Label) string {
	switch label {
	case TicketLabel:
		return cfg.TripObjects.Ticket.One
	case BookingLabel:
		return cfg.TripObjects.Booking.One
	case EventLabel:
		return cfg.TripObjects.Event.One
	case TripLabel:
		return cfg.TripObjects.Trip.One
	default:
		return string(label)
	}
}

func editFields(cfg *lang.Cfg, label Label) []editField {
	switch label {
	case TicketLabel:
		return []editField{
			{0, cfg.TripObjects.Ticket.MainFields[0]},
			{1, cfg.TripObjects.Ticket.MainFields[1]},
			{2, cfg.TripObjects.Ticket.MainFields[2]},
			{3, cfg.TripObjects.Ticket.MainFields[3]},
			{4, cfg.TripObjects.Ticket.MainFields[4]},
			{5, cfg.TripObjects.Ticket.OtherFields[0]},
			{6, cfg.TripObjects.Ticket.OtherFields[1]},
			{7, cfg.TripObjects.Ticket.OtherFields[2]},
			{8, cfg.TripObjects.Ticket.OtherFields[3]},
			{9, cfg.TripObjects.Ticket.OtherFields[4]},
			{10, cfg.TripObjects.Ticket.OtherFields[5]},
			{11, cfg.TripObjects.Ticket.OtherFields[6]},
		}

	case BookingLabel:
		return []editField{
			{0, cfg.TripObjects.Booking.MainFields[0]},
			{1, cfg.TripObjects.Booking.MainFields[1]},
			{2, cfg.TripObjects.Booking.MainFields[2]},
			{3, cfg.TripObjects.Booking.MainFields[3]},
			{4, cfg.TripObjects.Booking.OtherFields[0]},
			{5, cfg.TripObjects.Booking.OtherFields[1]},
			{6, cfg.TripObjects.Booking.OtherFields[2]},
			{7, cfg.TripObjects.Booking.OtherFields[3]},
			{8, cfg.TripObjects.Booking.OtherFields[4]},
			{9, cfg.TripObjects.Booking.OtherFields[5]},
			{10, cfg.TripObjects.Booking.OtherFields[6]},
		}

	case EventLabel:
		return []editField{
			{0, cfg.TripObjects.Event.MainFields[0]},
			{1, cfg.TripObjects.Event.MainFields[1]},
			{2, cfg.TripObjects.Event.MainFields[2]},
			{3, cfg.TripObjects.Event.OtherFields[0]},
			{4, cfg.TripObjects.Event.OtherFields[1]},
			{5, cfg.TripObjects.Event.OtherFields[2]},
			{6, cfg.TripObjects.Event.OtherFields[3]},
		}

	case TripLabel:
		return []editField{
			{0, cfg.TripObjects.Trip.MainFields[0]},
			{1, cfg.TripObjects.Trip.MainFields[1]},
			{2, cfg.TripObjects.Trip.MainFields[2]},
			{3, cfg.TripObjects.Trip.OtherFields[0]},
		}

	default:
		return nil
	}
}

func (h CommonHandler[T]) HandleEditSingle(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	cfg *lang.Cfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	fields := editFields(cfg, h.label)
	if len(fields) == 0 {
		return nil, fmt.Errorf("no editable fields for %s", h.label)
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, field := range fields {
		rows = append(
			rows,
			[]tgbotapi.InlineKeyboardButton{
				button(fmt.Sprintf("%s", field.Name), makeEditCallback(h.label, objectID, field.ID)),
			},
		)
	}

	rows = append(
		rows,
		[]tgbotapi.InlineKeyboardButton{
			button(cfg.Buttons.Back, MakeCallback(h.label, GetSingle, objectID)),
		},
	)

	object, err := h.service.GetByID(ctx, callback.From.ID, objectID)
	if err != nil {
		return nil, err
	}

	objectCfg := TripObjectCfgByLabel(cfg, h.label)

	text := fmt.Sprintf(
		"%s.\n\n%s",
		cfg.WhatEdit,
		(*object).Show(&objectCfg),
	)

	keyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	return &edit, nil
}

func (h CommonHandler[T]) HandleEditField(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	fieldID int,
	cfg *lang.Cfg,
) (*tgbotapi.MessageConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	fields := editFields(cfg, h.label)

	var fieldName string

	for _, field := range fields {
		if field.ID == fieldID {
			fieldName = field.Name
			break
		}
	}

	if fieldName == "" {
		return nil, fmt.Errorf(
			"unknown edit field: label=%s field=%d",
			h.label,
			fieldID,
		)
	}

	objectName := editObjectName(cfg, h.label)

	text := fmt.Sprintf(
		"%s\n[%d] %s, %s №%d\n%s",
		cfg.OnEdit,
		fieldID+1,
		fieldName,
		objectName,
		objectID,
		cfg.OnEditReply,
	)

	msg := tgbotapi.NewMessage(
		callback.Message.Chat.ID,
		text,
	)

	msg.ReplyToMessageID = callback.Message.MessageID
	msg.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}

	return &msg, nil
}

func (h CommonHandler[T]) HandleEditReply(
	ctx context.Context,
	msg *tgbotapi.Message,
	cfg *lang.Cfg,
) (tgbotapi.Chattable, error) {
	if msg == nil {
		return nil, nil
	}

	if msg.ReplyToMessage == nil {
		return nil, nil
	}

	editCtx, err := ParseEditContext(msg.ReplyToMessage.Text, cfg)
	if err != nil {
		return nil, nil
	}

	if editCtx.Label != h.label {
		return nil, nil
	}

	value := strings.TrimSpace(lang.Clean(msg.Text))
	if value == "" {
		return nil, fmt.Errorf("empty edit value")
	}

	if utf8.RuneCountInString(value) > 256 {
		return nil, ErrTooLongField
	}

	field, err := parseEditValue(editCtx.Label, editCtx.FieldID, value)
	if err != nil {
		return nil, err
	}

	if err := h.service.UpdateField(ctx, editCtx.ObjectID, msg.From.ID, field.Name, field.Value); err != nil {
		return nil, err
	}
	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, cfg.Edited)
	return replyMsg, nil
}

type parsedEditValue struct {
	Name  string
	Value any
}

func parseEditValue(
	label Label,
	fieldID int,
	value string,
) (parsedEditValue, error) {
	switch label {
	case TicketLabel:
		return parseTicketEditValue(fieldID, value)
	case BookingLabel:
		return parseBookingEditValue(fieldID, value)
	case EventLabel:
		return parseEventEditValue(fieldID, value)
	case TripLabel:
		return parseTripEditValue(fieldID, value)
	default:
		return parsedEditValue{}, fmt.Errorf("unsupported label: %s", label)
	}
}

func parseTicketEditValue(fieldID int, value string) (parsedEditValue, error) {
	switch fieldID {
	case 0:
		return parsedEditValue{"vehicle", value}, nil

	case 1:
		return parsedEditValue{"departure_point", value}, nil

	case 2:
		return parsedEditValue{"arrival_point", value}, nil

	case 3:
		t, err := time.Parse(DatetimeLayout, value)
		if err != nil {
			return parsedEditValue{}, ErrIncorrectDateTime
		}
		return parsedEditValue{"departure_time", t}, nil

	case 4:
		t, err := time.Parse(DatetimeLayout, value)
		if err != nil {
			return parsedEditValue{}, ErrIncorrectDateTime
		}
		return parsedEditValue{"arrival_time", t}, nil

	case 5:
		if value == "-" {
			return parsedEditValue{"cost", nil}, nil
		}
		nValue := NormalizeCost(value)
		return parsedEditValue{"cost", &nValue}, nil

	case 6:
		if value == "-" {
			return parsedEditValue{"currency", nil}, nil
		}
		nValue := NormalizeCurrency(value)
		return parsedEditValue{"currency", &nValue}, nil

	case 7:
		if value == "-" {
			return parsedEditValue{"departure_address", nil}, nil
		}
		return parsedEditValue{"departure_address", &value}, nil

	case 8:
		if value == "-" {
			return parsedEditValue{"arrival_address", nil}, nil
		}
		return parsedEditValue{"arrival_address", &value}, nil

	case 9:
		if value == "-" {
			return parsedEditValue{"seat", nil}, nil
		}
		return parsedEditValue{"seat", &value}, nil

	case 10:
		if value == "-" {
			return parsedEditValue{"number", nil}, nil
		}
		return parsedEditValue{"number", &value}, nil

	case 11:
		if value == "-" {
			return parsedEditValue{"comment", nil}, nil
		}
		return parsedEditValue{"comment", &value}, nil

	default:
		return parsedEditValue{}, fmt.Errorf("unknown ticket field: %d", fieldID)
	}
}

func parseBookingEditValue(fieldID int, value string) (parsedEditValue, error) {
	switch fieldID {
	case 0:
		return parsedEditValue{"city", value}, nil

	case 1:
		return parsedEditValue{"place", value}, nil

	case 2:
		t, err := time.Parse(DateLayout, value)
		if err != nil {
			return parsedEditValue{}, ErrIncorrectDate
		}
		return parsedEditValue{"check_in_date", t}, nil

	case 3:
		t, err := time.Parse(DateLayout, value)
		if err != nil {
			return parsedEditValue{}, ErrIncorrectDate
		}
		return parsedEditValue{"check_out_date", t}, nil

	case 4:
		if value == "-" {
			return parsedEditValue{"cost", nil}, nil
		}
		nValue := NormalizeCost(value)
		return parsedEditValue{"cost", &nValue}, nil

	case 5:
		if value == "-" {
			return parsedEditValue{"currency", nil}, nil
		}
		nValue := NormalizeCurrency(value)
		return parsedEditValue{"currency", &nValue}, nil

	case 6:
		if value == "-" {
			return parsedEditValue{"address", nil}, nil
		}
		return parsedEditValue{"address", &value}, nil

	case 7:
		if value == "-" {
			return parsedEditValue{"check_in_time", nil}, nil
		}

		t, err := time.Parse(TimeLayout, value)
		if err != nil {
			return parsedEditValue{}, ErrIncorrectTime
		}

		return parsedEditValue{"check_in_time", t}, nil

	case 8:
		if value == "-" {
			return parsedEditValue{"check_out_time", nil}, nil
		}

		t, err := time.Parse(TimeLayout, value)
		if err != nil {
			return parsedEditValue{}, ErrIncorrectTime
		}

		return parsedEditValue{"check_out_time", t}, nil

	case 9:
		if value == "-" {
			return parsedEditValue{"number", nil}, nil
		}
		return parsedEditValue{"number", &value}, nil

	case 10:
		if value == "-" {
			return parsedEditValue{"comment", nil}, nil
		}
		return parsedEditValue{"comment", &value}, nil

	default:
		return parsedEditValue{}, fmt.Errorf("unknown booking field: %d", fieldID)
	}
}

func parseEventEditValue(fieldID int, value string) (parsedEditValue, error) {
	switch fieldID {
	case 0:
		return parsedEditValue{"name", value}, nil

	case 1:
		return parsedEditValue{"location", value}, nil

	case 2:
		t, err := time.Parse(DatetimeLayout, value)
		if err != nil {
			return parsedEditValue{}, ErrIncorrectDateTime
		}
		return parsedEditValue{"start_time", t}, nil

	case 3:
		if value == "-" {
			return parsedEditValue{"end_time", nil}, nil
		}

		t, err := time.Parse(DatetimeLayout, value)
		if err != nil {
			return parsedEditValue{}, ErrIncorrectDateTime
		}

		return parsedEditValue{"end_time", t}, nil

	case 4:
		if value == "-" {
			return parsedEditValue{"cost", nil}, nil
		}
		nValue := NormalizeCost(value)
		return parsedEditValue{"cost", &nValue}, nil

	case 5:
		if value == "-" {
			return parsedEditValue{"currency", nil}, nil
		}
		nValue := NormalizeCurrency(value)
		return parsedEditValue{"currency", &nValue}, nil

	case 6:
		if value == "-" {
			return parsedEditValue{"comment", nil}, nil
		}
		return parsedEditValue{"comment", &value}, nil

	default:
		return parsedEditValue{}, fmt.Errorf("unknown event field: %d", fieldID)
	}
}

func parseTripEditValue(fieldID int, value string) (parsedEditValue, error) {
	switch fieldID {
	case 0:
		return parsedEditValue{"name", value}, nil

	case 1:
		t, err := time.Parse(DateLayout, value)
		if err != nil {
			return parsedEditValue{}, ErrIncorrectDate
		}
		return parsedEditValue{"start_date", t}, nil

	case 2:
		t, err := time.Parse(DateLayout, value)
		if err != nil {
			return parsedEditValue{}, ErrIncorrectDate
		}
		return parsedEditValue{"end_date", t}, nil

	case 3:
		if value == "-" {
			return parsedEditValue{"comment", nil}, nil
		}
		return parsedEditValue{"comment", &value}, nil

	default:
		return parsedEditValue{}, fmt.Errorf("unknown trip field: %d", fieldID)
	}
}
