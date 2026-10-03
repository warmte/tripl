package common

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/warmte/tripl/lang"
)

func BMain(cfg *lang.Cfg) [][]string {
	return [][]string{
		{cfg.Buttons.Add},
		{cfg.TripObjects.Ticket.Many, cfg.TripObjects.Booking.Many, cfg.TripObjects.Event.Many, cfg.TripObjects.Trip.Many},
		{cfg.Buttons.Settings}}
}

func BList(listType ListType, cfg *lang.Cfg, objectCfg *lang.TripObjectCfg) [][]string {
	first := objectCfg.Actual
	if listType == ArchiveListType {
		first = objectCfg.Archive
	}
	return [][]string{{first, cfg.Buttons.Back}}
}

func BListAll(cfg *lang.Cfg, objectCfg *lang.TripObjectCfg) [][]string {
	return [][]string{{objectCfg.Actual, objectCfg.Archive, cfg.Buttons.Back}}
}

func BuildKeyboard(buttons [][]string, cfg *lang.Cfg) [][]tgbotapi.KeyboardButton {
	raws := make([][]tgbotapi.KeyboardButton, len(buttons))
	for i, r := range buttons {
		cols := make([]tgbotapi.KeyboardButton, len(r))
		for j, b := range buttons[i] {
			cols[j] = tgbotapi.NewKeyboardButton(b)
		}
		raws[i] = tgbotapi.NewKeyboardButtonRow(cols...)
	}
	return raws
}

func BuildMsgWithKeyboard(chatId int64, text string, buttons [][]string, cfg *lang.Cfg) tgbotapi.Chattable {
	raws := BuildKeyboard(buttons, cfg)
	msg := tgbotapi.NewMessage(chatId, lang.EscapeMarkdownV2(text))
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(raws...)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	return msg
}