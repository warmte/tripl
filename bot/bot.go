package bot

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/handler"
	"github.com/warmte/tripl/lang"
	"github.com/warmte/tripl/service"
)

type Bot struct {
	api             *tgbotapi.BotAPI
	ctx             context.Context
	userService     *service.UserService
	callbackHandler *handler.CallbackHandler
	handlers        *handler.Handlers
}

func NewBot(
	api *tgbotapi.BotAPI,
	ctx context.Context,
	userService *service.UserService,
	callbackHandler *handler.CallbackHandler,
	handlers *handler.Handlers,
) *Bot {
	return &Bot{api, ctx, userService, callbackHandler, handlers}
}

func (bot Bot) getLangCfg(userID int64) *lang.Cfg {
	language, err := bot.userService.GetLanguage(bot.ctx, userID)
	if err == nil && lang.Language(language) == lang.LanguageEn {
		return lang.Get(lang.LanguageEn)
	} else {
		return lang.Get(lang.LanguageRu)
	}
}

func (bot Bot) replyWithKeyboard(userID int64, text string, buttons [][]string) {
	raws := BuildKeyboard(buttons, bot.getLangCfg(userID))
	replyMsg := tgbotapi.NewMessage(userID, lang.EscapeMarkdownV2(text))
	replyMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(raws...)
	replyMsg.ParseMode = tgbotapi.ModeMarkdownV2
	bot.api.Send(replyMsg)
}

func (bot Bot) replyForError(userID int64, err error) {
	langCfg := bot.getLangCfg(userID)
	bot.replyWithKeyboard(userID, GetErrorMessage(err, *langCfg), [][]string{{langCfg.Buttons.Cancel}})
	log.Printf("internal error: %v", err)
}

func (bot Bot) matchByText(msg *tgbotapi.Message, matchCmd string, matchValues []string) bool {
	return msg.IsCommand() && msg.Command() == matchCmd || slices.Contains(matchValues, msg.Text)
}

func (bot Bot) matchByStage(msg *tgbotapi.Message, currentStage string) bool {
	return bot.userService.GetState(msg.Chat.ID).Stage == currentStage
}

func (bot Bot) matchByCmd(msg *tgbotapi.Message, matchCmdPrefix string) bool {
	return msg.IsCommand() && strings.HasPrefix(msg.Command(), matchCmdPrefix)

}
func (bot Bot) parseByCmd(msg *tgbotapi.Message, matchCmdPrefix string) int {
	suffix := strings.TrimPrefix(msg.Command(), matchCmdPrefix)
	n, err := strconv.Atoi(suffix)
	if err != nil || n <= 0 {
		return -1
	}
	return n
}

func (bot Bot) matchListObjects(
	msg *tgbotapi.Message,
	tripObject *TripObject,
) bool {
	if bot.matchByText(
		msg,
		fmt.Sprintf("%s_%s", ActualListType, tripObject.Label),
		[]string{tripObject.Cfg.Many, tripObject.Cfg.Actual},
	) {
		langCfg := bot.getLangCfg(msg.Chat.ID)
		newMsg, err := bot.handlers.ShowListFor(tripObject.Label, bot.ctx, msg.Chat.ID, ActualListType, langCfg)
		if err == nil {
			bot.api.Send(*newMsg)
		}

		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s. %s", tripObject.Cfg.Actual, langCfg.Info),
			[][]string{{tripObject.Cfg.Archive, langCfg.Buttons.Back}})

		return true
	}

	if bot.matchByText(
		msg,
		fmt.Sprintf("%s_%s", ArchiveListType, tripObject.Label),
		[]string{tripObject.Cfg.Many, tripObject.Cfg.Archive},
	) {
		langCfg := bot.getLangCfg(msg.Chat.ID)
		newMsg, err := bot.handlers.ShowListFor(tripObject.Label, bot.ctx, msg.Chat.ID, ArchiveListType, langCfg)
		if err == nil {
			bot.api.Send(*newMsg)
		}

		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s. %s", tripObject.Cfg.Archive, langCfg.Info),
			[][]string{{tripObject.Cfg.Actual, langCfg.Buttons.Back}})

		return true
	}

	return false
}

func (bot Bot) matchShowSingle(msg *tgbotapi.Message) bool {
	if bot.matchByCmd(msg, "tt") {
		langCfg := bot.getLangCfg(msg.Chat.ID)
		id := bot.parseByCmd(msg, "tt")
		if id != -1 {
			newMsg, err := bot.handlers.TicketHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), langCfg)
			if err == nil {
				bot.api.Send(*newMsg)
			}
		}
		return true
	} else if bot.matchByCmd(msg, "bk") {
		langCfg := bot.getLangCfg(msg.Chat.ID)
		id := bot.parseByCmd(msg, "bk")
		if id != -1 {
			newMsg, err := bot.handlers.BookingHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), langCfg)
			if err == nil {
				bot.api.Send(*newMsg)
			}
		}
		return true
	} else if bot.matchByCmd(msg, "ev") {
		langCfg := bot.getLangCfg(msg.Chat.ID)
		id := bot.parseByCmd(msg, "ev")
		if id != -1 {
			newMsg, err := bot.handlers.EventHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), langCfg)
			if err == nil {
				bot.api.Send(*newMsg)
			}
		}
		return true
	} else if bot.matchByCmd(msg, "tr") {
		langCfg := bot.getLangCfg(msg.Chat.ID)
		id := bot.parseByCmd(msg, "tr")
		if id != -1 {
			newMsg, err := bot.handlers.TripHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), langCfg)
			if err == nil {
				bot.api.Send(*newMsg)
			}
		}
		return true
	}
	return false
}

func (bot Bot) matchAddMainFieldsObject(msg *tgbotapi.Message, tripObject *TripObject) bool {
	if bot.matchByText(msg, fmt.Sprintf("add_%s", tripObject.Label), []string{tripObject.Cfg.One}) {
		langCfg := bot.getLangCfg(msg.Chat.ID)
		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s\n\n- %s\n\n%s", tripObject.Cfg.AddMainFields, strings.Join(tripObject.Cfg.MainFields, "\n- "), tripObject.Cfg.AddExample.Main),
			[][]string{{langCfg.Buttons.Cancel}})
		bot.userService.SetState(msg.Chat.ID, service.AddInfo.For(string(tripObject.Label)))
		return true
	}
	return false
}

func (bot Bot) matchAddOtherFieldsObject(msg *tgbotapi.Message, tripObject *TripObject) bool {
	possibleCurrentStage := service.AddInfo.For(string(tripObject.Label)).Stage
	if bot.matchByStage(msg, possibleCurrentStage) {
		mainFieldsParsed, err := selectMainFieldsParser(msg.Text, tripObject.Label, msg.Chat.ID)
		if err != nil {
			bot.replyForError(msg.Chat.ID, err)
			return true
		}
		langCfg := bot.getLangCfg(msg.Chat.ID)
		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s\n\n- %s\n\n%s\n\n%s", tripObject.Cfg.AddOtherFields, strings.Join(tripObject.Cfg.OtherFields, "\n- "), tripObject.Cfg.AddExample.Other, tripObject.Cfg.AddOtherFieldsTips),
			[][]string{{langCfg.Buttons.Skip, langCfg.Buttons.Cancel}})
		bot.userService.SetState(msg.Chat.ID, service.Added.For(string(tripObject.Label)).With(mainFieldsParsed))
		return true
	}
	return false
}

func (bot Bot) matchAddedObject(msg *tgbotapi.Message, tripObject *TripObject) bool {
	possibleCurrentStage := service.Added.For(string(tripObject.Label)).Stage
	if bot.matchByStage(msg, possibleCurrentStage) {
		langCfg := bot.getLangCfg(msg.Chat.ID)
		currentContext := bot.userService.GetState(msg.Chat.ID).Context
		if msg.Text != langCfg.Buttons.Skip {
			otherFieldsParsed, err := selectOtherFieldsParser(msg.Text, tripObject.Label, currentContext)
			if err == ErrInternalError {
				bot.replyForError(msg.Chat.ID, err)
				return true
			}
			if err != nil {
				bot.replyForError(msg.Chat.ID, err)
				return true
			}
			currentContext = otherFieldsParsed
		}

		added, err := bot.handlers.AddFor(tripObject.Label, bot.ctx, currentContext)

		if err != nil {
			bot.replyForError(msg.Chat.ID, err)
			return true
		}
		showAdded, err := bot.handlers.ShowSingleFor(bot.ctx, tripObject.Label, added.GetID(), msg.Chat.ID, langCfg)
		if err != nil {
			bot.replyForError(msg.Chat.ID, err)
			return true
		}
		bot.api.Send(*showAdded)
		bot.replyWithKeyboard(msg.Chat.ID, fmt.Sprintf("%s", langCfg.Added),
			[][]string{
				{langCfg.Buttons.Add},
				{langCfg.TripObjects.Ticket.Many, langCfg.TripObjects.Booking.Many, langCfg.TripObjects.Event.Many, langCfg.TripObjects.Trip.Many},
				{langCfg.Buttons.Settings},
			})
		bot.userService.SetState(msg.Chat.ID, service.Default)
		return true
	}
	return false
}

func (bot *Bot) Run() {
	bot.api.Debug = true

	log.Printf("Authorized on account %s", bot.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.api.GetUpdatesChan(u)

	processTripObjects := func(langCfg *lang.Cfg, f func(to TripObject) bool) bool {
		result := false
		tripObjects := TripObjects(langCfg)
		for _, to := range tripObjects {
			result = result || f(to)
		}
		return result
	}

	for update := range updates {
		switch {
		case update.CallbackQuery != nil:
			callback := update.CallbackQuery

			_, err := bot.api.Request(
				tgbotapi.NewCallback(callback.ID, ""),
			)
			if err != nil {
				log.Printf("callback answer: %v", err)
			}

			langCfg := bot.getLangCfg(callback.From.ID)
			edits, err := bot.callbackHandler.Handle(
				bot.ctx,
				callback,
				langCfg,
			)
			if err != nil {
				log.Printf("callback: %v", err)
				break
			}

			for _, edit := range edits {
				if edit != nil {
					if _, err := bot.api.Send(edit); err != nil {
						log.Printf("callback edit: %v", err)
					}
				}
			}
		case update.Message != nil:
			{
				msg := update.Message
				log.Printf("[%s] %s", msg.From.UserName, msg.Text)
				langCfg := bot.getLangCfg(msg.Chat.ID)
				if msg.ReplyToMessage != nil {
					{
						message, err := bot.handlers.HandleEditReply(
							bot.ctx,
							update.Message,
							langCfg,
						)
						if err != nil {
							bot.replyForError(update.Message.Chat.ID, err)
						} else {
							if message != nil {
								if _, err := bot.api.Send(message); err != nil {
									bot.replyForError(update.Message.Chat.ID, err)
								}
							}
						}
					}
				} else if bot.matchByText(msg, "menu", []string{langCfg.Buttons.Back, langCfg.Buttons.Cancel}) {
					bot.userService.SetState(msg.Chat.ID, service.Default)
					bot.replyWithKeyboard(msg.Chat.ID, langCfg.Menu, BMain(langCfg))
				} else if bot.matchByText(msg, "settings", []string{langCfg.Buttons.Settings}) {
					bot.userService.SetState(msg.Chat.ID, service.Default)
					bot.replyWithKeyboard(msg.Chat.ID, langCfg.Menu, [][]string{{langCfg.Buttons.Lang.Lang, langCfg.Buttons.Back}})
				} else if bot.matchByText(msg, "add", []string{langCfg.Buttons.Add}) {
					bot.userService.SetState(msg.Chat.ID, service.Default)
					bot.replyWithKeyboard(
						msg.Chat.ID,
						langCfg.Add,
						[][]string{{langCfg.TripObjects.Ticket.One, langCfg.TripObjects.Booking.One, langCfg.TripObjects.Event.One, langCfg.TripObjects.Trip.One}, {langCfg.Buttons.Back}})
				} else if processTripObjects(langCfg, func(tripObject TripObject) bool { return bot.matchListObjects(msg, &tripObject) }) {
				} else if processTripObjects(langCfg, func(tripObject TripObject) bool { return bot.matchAddMainFieldsObject(msg, &tripObject) }) {
				} else if processTripObjects(langCfg, func(tripObject TripObject) bool { return bot.matchAddOtherFieldsObject(msg, &tripObject) }) {
				} else if processTripObjects(langCfg, func(tripObject TripObject) bool { return bot.matchAddedObject(msg, &tripObject) }) {
				} else if bot.matchByText(msg, "start", []string{}) {
					bot.userService.SetState(msg.Chat.ID, service.Default)
					bot.replyWithKeyboard(
						msg.Chat.ID,
						fmt.Sprintf("%s\n\n%s", langCfg.Welcome, langCfg.ChooseLang),
						[][]string{
							{langCfg.Buttons.Lang.Ru, langCfg.Buttons.Lang.En}})
				} else if bot.matchByText(msg, "lang", []string{langCfg.Buttons.Lang.Lang}) {
					bot.userService.SetState(msg.Chat.ID, service.Default)
					bot.replyWithKeyboard(
						msg.Chat.ID,
						langCfg.ChooseLang,
						[][]string{
							{langCfg.Buttons.Lang.Ru, langCfg.Buttons.Lang.En}})
				} else if bot.matchByText(msg, "ru", []string{langCfg.Buttons.Lang.Ru}) {
					bot.userService.SetOrUpdateLang(bot.ctx, msg.Chat.ID, lang.LanguageRu)
					bot.replyWithKeyboard(msg.Chat.ID, langCfg.Menu, BMain(langCfg))
				} else if bot.matchByText(msg, "en", []string{langCfg.Buttons.Lang.En}) {
					bot.userService.SetOrUpdateLang(bot.ctx, msg.Chat.ID, lang.LanguageEn)
					bot.replyWithKeyboard(msg.Chat.ID, langCfg.Menu, BMain(langCfg))
				} else if bot.matchShowSingle(msg) {
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "unknown command")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.api.Send(msg)
				}
			}
		}
	}
}
