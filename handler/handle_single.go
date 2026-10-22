package handler

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
)

func (h CommonHandler[T]) Count(ctx context.Context, userID int64) (int64, error) {
	return h.service.CountByUserID(ctx, userID)
}

func (h CommonHandler[T]) Add(ctx context.Context, data TripObjectView) (*T, error) {
	return h.service.Create(ctx, data)
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

	return []tgbotapi.Chattable{&edit}, nil
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

func makeSingleKeyboard(label Label, objectID int64, tripID *int64, cfg *lang.Cfg) *tgbotapi.InlineKeyboardMarkup {
	deleteButton := button(cfg.Buttons.Delete, MakeCallback(label, DeleteSingle, objectID))
	editButton := button(cfg.Buttons.Edit, MakeCallback(label, EditSingle, objectID))

	keyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{editButton, deleteButton},
		},
	}

	if label != TripLabel {
		tripButton := button(cfg.Buttons.Attach, MakeExtendedCallback(label, AttachSingleActual, objectID, 0))
		if tripID != nil {
			tripButton = button(cfg.Buttons.Detach, MakeCallback(label, DetachSingle, objectID))
		}
		icsButton := button(cfg.Buttons.MakeIcs, MakeCallback(label, MakeIcs, objectID))

		keyboard = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{tripButton, editButton, deleteButton},
				{icsButton},
			},
		}
	} else {
		statsButton := button(cfg.Buttons.Stats, MakeCallback(label, TripStats, objectID))
		csvButton := button("→ .csv", MakeCallback(label, TripCSV, objectID))
		xlsxButton := button("→ .xlsx", MakeCallback(label, TripXLSX, objectID))
		ticketsButton := button(cfg.TripObjects.Ticket.Many, MakeExtendedCallback(label, TripTickets, objectID, 0))
		bookingsButton := button(cfg.TripObjects.Booking.Many, MakeExtendedCallback(label, TripBookings, objectID, 0))
		eventsButton := button(cfg.TripObjects.Event.Many, MakeExtendedCallback(label, TripEvents, objectID, 0))
		keyboard = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{statsButton, editButton, deleteButton},
				{ticketsButton, bookingsButton, eventsButton},
				{csvButton, xlsxButton},
			},
		}
	}

	return &keyboard
}
