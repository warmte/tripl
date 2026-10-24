package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
)

func (h CommonHandler[T]) ShowList(
	ctx context.Context,
	userID int64,
	listType ListType,
	page int,
	cfg *lang.Cfg,
) (*tgbotapi.MessageConfig, error) {

	objectCfg := TripObjectCfgByLabel(cfg, h.label)

	otherName := cfg.Buttons.Archive
	otherAction := ArchiveList
	title := objectCfg.Actual
	action := ActualList
	switch listType {
	case ActualListType:
	case ArchiveListType:
		action = ArchiveList
		otherName = objectCfg.Many
		otherAction = ActualList
		title = objectCfg.Archive
	default:
		return nil, fmt.Errorf("unknown list type: %s", listType)
	}

	result, err := h.service.GetPage(ctx, userID, listType, page)
	if err != nil {
		return nil, err
	}

	keyboard := makeListKeyboard(h.label, action, page, result.HasNext, true, otherName, otherAction)
	message := tgbotapi.NewMessage(userID, fmt.Sprintf("%s. %s\n\n%s", title, cfg.Info, h.RenderObjects(result.Items, objectCfg)))
	message.ReplyMarkup = keyboard

	return &message, nil
}

func (h CommonHandler[T]) HandleList(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	listCallback Callback,
	cfg *lang.Cfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	listType := ActualListType
	if listCallback.Action == ArchiveList {
		listType = ArchiveListType
	} else if listCallback.Action != ActualList {
		return nil, fmt.Errorf("unknown list action: %s", listCallback.Action)
	}

	showList, err := h.ShowList(ctx, callback.From.ID, listType, int(listCallback.V1), cfg)
	if err != nil {
		return nil, err
	}

	keyboard, ok := showList.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup)

	var edit tgbotapi.EditMessageTextConfig

	if ok {
		edit = tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			showList.Text,
			*keyboard,
		)
	} else {
		edit = tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			showList.Text,
		)
	}

	return &edit, nil
}

func (h CommonHandler[T]) RenderObjects(items []T, cfg lang.TripObjectCfg) string {
	if len(items) == 0 {
		return fmt.Sprintf("%s.", cfg.Empty)
	}

	var text strings.Builder

	for i := range items {
		if i > 0 {
			text.WriteString("\n\n")
		}

		text.WriteString(items[i].Preview(&cfg))
	}

	return text.String()
}

func (h CommonHandler[T]) renderTrips(items []Trip, cfg lang.TripObjectCfg) string {
	if len(items) == 0 {
		return fmt.Sprintf("%s.", cfg.Empty)
	}

	var text strings.Builder

	for i, item := range items {
		if i > 0 {
			text.WriteString("\n\n")
		}

		text.WriteString(fmt.Sprintf("[%d] %s\n%s-%s", item.ID, item.Name, item.StartDate.Format(DateLayout), item.EndDate.Format(DateLayout)))
	}

	return text.String()
}

func tripListAction(label Label) CallbackAction {
	switch label {
	case BookingLabel:
		return TripBookings
	case EventLabel:
		return TripEvents
	default:
		return TripTickets
	}
}

func (h CommonHandler[T]) HandleListForTrip(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	tripID int64,
	pageN int,
	cfg *lang.Cfg,
) (tgbotapi.Chattable, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}
	if h.label == TripLabel {
		return nil, ErrInternalError
	}
	page, err := h.GetPageByTripID(ctx, tripID, callback.From.ID, pageN)
	if err != nil {
		return nil, fmt.Errorf("get page: %w", err)
	}
	objectCfg := TripObjectCfgByLabel(cfg, h.label)
	text := h.RenderObjects(page.Items, objectCfg)
	keyboard := makeTripListKeyboard(h.label, tripListAction(h.label), tripID, int64(pageN), page.HasNext, cfg)

	var edit tgbotapi.EditMessageTextConfig

	if keyboard != nil {
		edit = tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			*keyboard,
		)
	} else {
		edit = tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
		)
	}
	return edit, nil
}

func makeListKeyboard(
	label Label,
	action CallbackAction,
	page int,
	hasNext bool,
	needOther bool,
	otherName string,
	otherAction CallbackAction,
) *tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton

	if needOther {
		buttons = append(buttons, button(otherName, MakeCallback(label, otherAction, 0)))
	}
	if page > 0 {
		buttons = append(buttons, button("←", MakeCallback(label, action, int64(page-1))))
	}
	if hasNext {
		buttons = append(buttons, button("→", MakeCallback(label, action, int64(page+1))))
	}
	if len(buttons) == 0 {
		return nil
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
	return &keyboard
}

func makeTripListKeyboard(
	label Label,
	list CallbackAction,
	tripID int64,
	page int64,
	hasNext bool,
	cfg *lang.Cfg,
) *tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton

	if page > 0 {
		buttons = append(buttons, button("←", MakeExtendedCallback(label, list, tripID, page-1)))
	}
	if hasNext {
		buttons = append(buttons, button("→", MakeExtendedCallback(label, list, tripID, page+1)))
	}
	buttons = append(buttons, button(cfg.Buttons.Back, MakeCallback(TripLabel, GetSingle, tripID)))
	if len(buttons) == 0 {
		return nil
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
	return &keyboard
}

func (s CommonHandler[T]) GetPageForTime(ctx context.Context, userID int64, start time.Time, end time.Time, page int) (Page[T], error) {
	return s.service.GetPageForTime(ctx, userID, start, end, page)
}

func (h CommonHandler[T]) buildListForDate(
	ctx context.Context,
	userID int64,
	page int,
	date time.Time,
	cfg *lang.Cfg,
) (string, *tgbotapi.InlineKeyboardMarkup, error) {
	location, err := time.LoadLocation(LocationName)
	if err != nil {
		return "", nil, fmt.Errorf("load scheduler location: %w", err)
	}
	start := time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, location)
	end := start.AddDate(0, 0, 1)
	if err != nil {
		return "", nil, err
	}

	result, err := h.GetPageForTime(ctx, userID, start, end, page)
	if err != nil {
		return "", nil, fmt.Errorf("get tomorrow page: %w", err)
	}
	if len(result.Items) == 0 {
		return "", nil, ErrEmptyList
	}
	item := result.Items[0]

	singleMsg, err := h.ShowSingle(ctx, userID, item.GetID(), cfg)
	if err != nil {
		return "", nil, err
	}
	singleKeyboard := makeSingleKeyboard(h.label, item.GetID(), item.GetTripID(), cfg)
	listKeyboard := makeListKeyboard(h.label, ForDate(date), page, result.HasNext, false, "", "")
	var rows [][]tgbotapi.InlineKeyboardButton
	if singleKeyboard != nil {
		rows = append(rows, singleKeyboard.InlineKeyboard...)
	}
	if listKeyboard != nil {
		rows = append(rows, listKeyboard.InlineKeyboard...)
	}
	keyboard := &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	return singleMsg.Text, keyboard, nil
}

func (h CommonHandler[T]) ShowTomorrowList(
	ctx context.Context,
	userID int64,
	cfg *lang.Cfg,
) (*tgbotapi.MessageConfig, error) {
	location, err := time.LoadLocation(LocationName)
	now := time.Now().In(location)
	text, keyboard, err := h.buildListForDate(ctx, userID, 0, now, cfg)
	if err != nil {
		return nil, err
	}

	message := tgbotapi.NewMessage(userID, text)
	message.ReplyMarkup = keyboard
	message.ParseMode = tgbotapi.ModeMarkdownV2

	return &message, nil
}

func (h CommonHandler[T]) HandleListForDate(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	page int,
	date time.Time,
	cfg *lang.Cfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	text, keyboard, err := h.buildListForDate(ctx, callback.From.ID, page, date, cfg)
	if err != nil {
		return nil, err
	}

	replyMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		*keyboard,
	)
	replyMsg.ParseMode = tgbotapi.ModeMarkdownV2

	if len(keyboard.InlineKeyboard) == 0 {
		replyMsg = tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
		)
	}

	return &replyMsg, nil
}
