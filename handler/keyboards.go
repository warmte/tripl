package handler

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
)

func makeListKeyboard(
	label Label,
	list CallbackAction,
	page int64,
	hasNext bool,
) *tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton

	if page > 0 {
		buttons = append(
			buttons,
			tgbotapi.NewInlineKeyboardButtonData(
				"←",
				MakeCallback(label, list, page-1),
			),
		)
	}

	if hasNext {
		buttons = append(
			buttons,
			tgbotapi.NewInlineKeyboardButtonData(
				"→",
				MakeCallback(label, list, page+1),
			),
		)
	}

	if len(buttons) == 0 {
		return nil
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
	return &keyboard
}

func makeSingleKeyboard(label Label, objectID int64, tripID *int64, cfg *lang.Cfg) *tgbotapi.InlineKeyboardMarkup {
	deleteButton := tgbotapi.NewInlineKeyboardButtonData(
		cfg.Buttons.Delete,
		fmt.Sprintf("%s:%s:%d", label, DeleteSingle, objectID),
	)

	editButton := tgbotapi.NewInlineKeyboardButtonData(
		cfg.Buttons.Edit,
		fmt.Sprintf("%s:%s:%d", label, EditSingle, objectID),
	)

	keyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{editButton, deleteButton},
		},
	}

	if label != TripLabel {
		tripButton := tgbotapi.NewInlineKeyboardButtonData(
			cfg.Buttons.Attach,
			fmt.Sprintf("%s:%s:%d:0", label, AttachSingleActual, objectID),
		)

		if tripID != nil {
			tripButton = tgbotapi.NewInlineKeyboardButtonData(
				cfg.Buttons.Detach,
				fmt.Sprintf("%s:%s:%d", label, DetachSingle, objectID),
			)
		}

		keyboard = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{tripButton, editButton, deleteButton},
			},
		}
	}

	return &keyboard
}
