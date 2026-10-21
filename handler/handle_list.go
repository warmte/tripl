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
	cfg *lang.Cfg,
) (*tgbotapi.MessageConfig, error) {
	const page = 0

	result, err := h.service.GetPage(ctx, userID, listType, page)
	if err != nil {
		return nil, err
	}

	var listAction CallbackAction
	switch listType {
	case ActualListType:
		listAction = ActualList
	case ArchiveListType:
		listAction = ArchiveList
	default:
		return nil, fmt.Errorf("unsopported list type: %s", listType)
	}

	keyboard := makeListKeyboard(h.label, listAction, page, result.HasNext)

	message := tgbotapi.NewMessage(
		userID,
		h.RenderObjects(
			result.Items,
			TripObjectCfgByLabel(cfg, h.label),
		),
	)

	message.ReplyMarkup = keyboard

	return &message, nil
}

func (h CommonHandler[T]) HandleList(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	listCallback Callback,
	cfg lang.TripObjectCfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	var listType ListType
	switch listCallback.Action {
	case ActualList:
		listType = ActualListType
	case ArchiveList:
		listType = ArchiveListType
	default:
		return nil, fmt.Errorf("unsupported list type")
	}

	page, err := h.service.GetPage(
		ctx,
		callback.From.ID,
		listType,
		int(listCallback.V1),
	)
	if err != nil {
		return nil, fmt.Errorf("get page: %w", err)
	}

	text := h.RenderObjects(page.Items, cfg)
	keyboard := makeListKeyboard(h.label, listCallback.Action, listCallback.V1, page.HasNext)

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
	list CallbackAction,
	page int64,
	hasNext bool,
) *tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton

	if page > 0 {
		buttons = append(buttons, button("←", MakeCallback(label, list, page-1)))
	}
	if hasNext {
		buttons = append(buttons, button("→", MakeCallback(label, list, page+1)))
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

func getStartEndTime(location *time.Location) (time.Time, time.Time, error) {
	now := time.Now().In(location)
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, location)
	end := start.AddDate(0, 0, 1)

	return start, end, nil
}

func (h CommonHandler[T]) buildTomorrowList(
	ctx context.Context,
	userID int64,
	page int,
	cfg *lang.Cfg,
) (string, *tgbotapi.InlineKeyboardMarkup, error) {
	location, err := time.LoadLocation(LocationName)
	if err != nil {
		return "", nil, fmt.Errorf("load scheduler location: %w", err)
	}
	start, end, err := getStartEndTime(location)
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
	listKeyboard := makeListKeyboard(h.label, TomorrowList, int64(page), result.HasNext)
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
	text, keyboard, err := h.buildTomorrowList(ctx, userID, 0, cfg)
	if err != nil {
		return nil, err
	}

	message := tgbotapi.NewMessage(userID, text)
	message.ReplyMarkup = keyboard
	message.ParseMode = tgbotapi.ModeMarkdownV2

	return &message, nil
}

func (h CommonHandler[T]) HandleTomorrowList(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	page int,
	cfg *lang.Cfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	text, keyboard, err := h.buildTomorrowList(ctx, callback.From.ID, page, cfg)
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
