package handler

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
	"github.com/warmte/tripl/service"
)

type CommonHandler[T TripObjectView] struct {
	label       Label
	service     service.Service[T]
	tripService service.Service[Trip]
}

func (h CommonHandler[T]) Add(ctx context.Context, data TripObjectView) (*T, error) {
	return h.service.Create(ctx, data)
}

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

func (h CommonHandler[T]) ShowSingle(
	ctx context.Context,
	userID int64,
	objectID int64,
	cfg *lang.Cfg,
) (*tgbotapi.MessageConfig, error) {
	objectCfg := TripObjectCfgByLabel(cfg, h.label)

	object, err := h.service.GetByID(ctx, userID, objectID)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", h.label, err)
	}
	text := fmt.Sprintf("```%s\n%s\n\n```", objectCfg.One, (*object).Show(&objectCfg))
	tripId := (*object).GetTripID()
	if h.label != TripLabel {
		tripInfo := ""
		if tripId != nil {
			trip, err := h.tripService.GetByID(ctx, userID, *tripId)
			if err == nil {
				tripInfo = trip.Preview(&cfg.TripObjects.Trip)
			}
		}
		if tripInfo != "" {
			text += fmt.Sprintf("🔗 %s", tripInfo)
		}
	} else {

	}

	msg := tgbotapi.NewMessage(userID, lang.EscapeMarkdownV2(text))
	msg.ReplyMarkup = makeSingleKeyboard(h.label, objectID, tripId, cfg)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	return &msg, nil
}

func (h CommonHandler[T]) HandleGetSingle(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	cfg *lang.Cfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	message, err := h.ShowSingle(ctx, callback.From.ID, objectID, cfg)
	if err != nil {
		return nil, err
	}

	keyboard, ok := message.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup)
	if !ok {
		return nil, fmt.Errorf("invalid single message keyboard")
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message.Text,
		*keyboard,
	)

	edit.ParseMode = message.ParseMode

	return &edit, nil
}

func (h CommonHandler[T]) HandleDeleteSingle(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	cfg *lang.Cfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	deleteButton := button(cfg.Buttons.Delete, MakeCallback(h.label, DeleteSingleConfirmed, objectID))
	cancelButton := button(cfg.Buttons.Cancel, MakeCallback(h.label, GetSingle, objectID))

	keyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{cancelButton, deleteButton},
		},
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		cfg.OnDelete,
		keyboard,
	)

	return &edit, nil
}

func (h CommonHandler[T]) HandleDeleteSingleConfirmed(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	if err := h.service.Delete(ctx, objectID, callback.From.ID); err != nil {
		return nil, fmt.Errorf("delete %s: %w", h.label, err)
	}

	edit := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		cfg.Deleted,
	)

	return []tgbotapi.Chattable{&edit, BuildMsgWithKeyboard(edit.ChatID, cfg.Menu, BMain(cfg), cfg)}, nil
}

func (h CommonHandler[T]) HandleAttachSingle(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	action CallbackAction,
	objectID int64,
	page int,
	cfg *lang.Cfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}

	listType := ActualListType
	if action == AttachSingleArchived {
		listType = ArchiveListType
	}

	tripPage, err := h.tripService.GetPage(ctx, callback.From.ID, listType, page)
	if err != nil {
		return nil, fmt.Errorf("get trips: %w", err)
	}

	text := h.renderTrips(tripPage.Items, cfg.TripObjects.Trip)

	if len(tripPage.Items) > 0 {
		text = fmt.Sprintf("%s:\n\n%s", cfg.TripObjects.Trip.Many, text)
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	var navigation []tgbotapi.InlineKeyboardButton

	if page > 0 {
		navigation = append(
			navigation,
			button("←", fmt.Sprintf("%s:%s:%d:%d", h.label, action, objectID, page-1)),
		)
	}

	if tripPage.HasNext {
		navigation = append(
			navigation,
			button("→", fmt.Sprintf("%s:%s:%d:%d", h.label, action, objectID, page+1)),
		)
	}

	if len(navigation) > 0 {
		rows = append(rows, navigation)
	}

	changeListView :=
		button(
			cfg.TripObjects.Trip.Archive,
			fmt.Sprintf("%s:%s:%d:0", h.label, AttachSingleArchived, objectID),
		)
	if action != AttachSingleActual {
		changeListView =
			button(
				cfg.TripObjects.Trip.Actual,
				MakeExtendedCallback(h.label, AttachSingleActual, objectID, 0),
			)
	}

	for _, trip := range tripPage.Items {
		rows = append(
			rows,
			[]tgbotapi.InlineKeyboardButton{
				button(
					fmt.Sprintf("[%d] %s", trip.ID, trip.Name),
					fmt.Sprintf("%s:%s:%d:%d", h.label, AttachTrip, objectID, trip.ID),
				),
			},
		)
	}

	rows = append(
		rows,
		[]tgbotapi.InlineKeyboardButton{
			changeListView,
			button(cfg.Buttons.Cancel, MakeCallback(h.label, GetSingle, objectID)),
		},
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

func (h CommonHandler[T]) HandleAttachTrip(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	tripID int64,
	cfg *lang.Cfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}
	if err := h.service.AttachToTrip(ctx, objectID, tripID, callback.From.ID); err != nil {
		return nil, fmt.Errorf("attach %s to trip: %w", h.label, err)
	}
	return h.HandleGetSingle(ctx, callback, objectID, cfg)
}

func (h CommonHandler[T]) HandleDetachSingle(
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	objectID int64,
	cfg *lang.Cfg,
) (*tgbotapi.EditMessageTextConfig, error) {
	if callback.Message == nil {
		return nil, fmt.Errorf("callback message is nil")
	}
	if err := h.service.DetachFromTrip(ctx, objectID, callback.From.ID); err != nil {
		return nil, fmt.Errorf("detach %s from trip: %w", h.label, err)
	}
	return h.HandleGetSingle(ctx, callback, objectID, cfg)
}

func (h CommonHandler[T]) GetPageByTripID(ctx context.Context, tripID int64, userID int64, page int) (Page[T], error) {
	if h.label == TripLabel {
		return Page[T]{}, ErrInternalError
	}
	return h.service.GetPageByTripID(ctx, tripID, userID, page)
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
	keyboard := makeTripListKeyboard(TicketLabel, TripTickets, tripID, int64(pageN), page.HasNext, cfg)

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

func (h CommonHandler[T]) SelectAction(
	parsedCallback Callback,
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	unsupportedCallbackError := fmt.Errorf(
		"unsupported callback: %s:%s:%d",
		parsedCallback.Label,
		parsedCallback.Action,
		parsedCallback.V1,
	)

	switch parsedCallback.Action {
	case ActualList, ArchiveList:
		r, err := h.HandleList(ctx, callback, parsedCallback, cfg.TripObjects.Ticket)
		return []tgbotapi.Chattable{r}, err
	case GetSingle:
		r, err := h.HandleGetSingle(ctx, callback, parsedCallback.V1, cfg)
		return []tgbotapi.Chattable{r}, err
	case EditSingle:
		r, err := h.HandleEditSingle(ctx, callback, parsedCallback.V1, cfg)
		return []tgbotapi.Chattable{r}, err
	case DeleteSingle:
		r, err := h.HandleDeleteSingle(ctx, callback, parsedCallback.V1, cfg)
		return []tgbotapi.Chattable{r}, err
	case DeleteSingleConfirmed:
		return h.HandleDeleteSingleConfirmed(ctx, callback, parsedCallback.V1, cfg)
	case DetachSingle:
		r, err := h.HandleDetachSingle(ctx, callback, parsedCallback.V1, cfg)
		return []tgbotapi.Chattable{r}, err
	default:
		return nil, unsupportedCallbackError
	}
}

func (h CommonHandler[T]) SelectExtendedAction(
	parsedCallback ExtendedCallback,
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	unsupportedCallbackError := fmt.Errorf(
		"unsupported callback: %s:%s:%d%d",
		parsedCallback.Label,
		parsedCallback.Action,
		parsedCallback.V1,
		parsedCallback.V2,
	)

	switch parsedCallback.Action {
	case AttachSingleActual, AttachSingleArchived:
		r, err := h.HandleAttachSingle(ctx, callback, parsedCallback.Action, parsedCallback.V1, int(parsedCallback.V2), cfg)
		return []tgbotapi.Chattable{r}, err
	case AttachTrip:
		r, err := h.HandleAttachTrip(ctx, callback, parsedCallback.V1, parsedCallback.V2, cfg)
		return []tgbotapi.Chattable{r}, err
	case EditField:
		r, err := h.HandleEditField(ctx, callback, parsedCallback.V1, int(parsedCallback.V2), cfg)
		return []tgbotapi.Chattable{r}, err
	default:
		return nil, unsupportedCallbackError
	}
}
