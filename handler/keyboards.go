package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
)

func button(text string, callback string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, callback)
}

func makeListKeyboard(
	label Label,
	list CallbackAction,
	page int64,
	hasNext bool,
) *tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton

	if page > 0 {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("←", MakeCallback(label, list, page-1)))
	}
	if hasNext {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("→", MakeCallback(label, list, page+1)))
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
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("←", MakeExtendedCallback(label, list, tripID, page-1)))
	}
	if hasNext {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("→", MakeExtendedCallback(label, list, tripID, page+1)))
	}
	buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(cfg.Buttons.Back, MakeCallback(TripLabel, GetSingle, tripID)))
	if len(buttons) == 0 {
		return nil
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
	return &keyboard
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
			tripButton = tgbotapi.NewInlineKeyboardButtonData(cfg.Buttons.Detach, MakeCallback(label, DetachSingle, objectID))
		}

		keyboard = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{tripButton, editButton, deleteButton},
			},
		}
	} else {
		statsButton := button(cfg.Buttons.Stats, MakeCallback(label, TripStats, objectID))
		ticketsButton := button(cfg.TripObjects.Ticket.Many, MakeExtendedCallback(label, TripTickets, objectID, 0))
		bookingsButton := button(cfg.TripObjects.Booking.Many, MakeExtendedCallback(label, TripBookings, objectID, 0))
		eventsButton := button(cfg.TripObjects.Event.Many, MakeExtendedCallback(label, TripEvents, objectID, 0))
		keyboard = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{statsButton, editButton, deleteButton},
				{ticketsButton, bookingsButton, eventsButton},
			},
		}
	}

	return &keyboard
}
