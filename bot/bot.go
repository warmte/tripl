package bot

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
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
	documentHandler *handler.DocumentHandler
	handlers        *handler.Handlers
}

func NewBot(
	api *tgbotapi.BotAPI,
	ctx context.Context,
	userService *service.UserService,
	callbackHandler *handler.CallbackHandler,
	documentHandler *handler.DocumentHandler,
	handlers *handler.Handlers,
) *Bot {
	return &Bot{api, ctx, userService, callbackHandler, documentHandler, handlers}
}

func (bot Bot) replyWithKeyboard(userID int64, text string, buttons [][]string, cfg *lang.Cfg) {
	raws := BuildKeyboard(buttons, cfg)
	replyMsg := tgbotapi.NewMessage(userID, lang.EscapeMarkdownV2(text))
	replyMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(raws...)
	replyMsg.ParseMode = tgbotapi.ModeMarkdownV2
	bot.api.Send(replyMsg)
}

func (bot Bot) replyForError(userID int64, err error, cfg *lang.Cfg) {
	bot.replyWithKeyboard(userID, GetErrorMessage(err, *cfg), [][]string{{cfg.Buttons.Cancel}}, cfg)
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
	cfg *lang.Cfg,
) bool {
	if bot.matchByText(
		msg,
		fmt.Sprintf("%s_%s", ActualListType, tripObject.Label),
		[]string{tripObject.Cfg.Many, tripObject.Cfg.Actual},
	) {
		newMsg, err := bot.handlers.ShowListFor(tripObject.Label, bot.ctx, msg.Chat.ID, ActualListType, cfg)
		if err == nil {
			bot.api.Send(*newMsg)
		}

		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s. %s", tripObject.Cfg.Actual, cfg.Info),
			[][]string{{tripObject.Cfg.Archive, cfg.Buttons.Back}}, cfg)

		return true
	}

	if bot.matchByText(
		msg,
		fmt.Sprintf("%s_%s", ArchiveListType, tripObject.Label),
		[]string{tripObject.Cfg.Many, tripObject.Cfg.Archive},
	) {
		newMsg, err := bot.handlers.ShowListFor(tripObject.Label, bot.ctx, msg.Chat.ID, ArchiveListType, cfg)
		if err == nil {
			bot.api.Send(*newMsg)
		}

		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s. %s", tripObject.Cfg.Archive, cfg.Info),
			[][]string{{tripObject.Cfg.Actual, cfg.Buttons.Back}}, cfg)

		return true
	}

	return false
}

func (bot Bot) matchShowSingle(msg *tgbotapi.Message, cfg *lang.Cfg) bool {
	var newMsg tgbotapi.Chattable
	var err error
	if bot.matchByCmd(msg, "tt") {
		id := bot.parseByCmd(msg, "tt")
		if id != -1 {
			newMsg, err = bot.handlers.TicketHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), cfg)
		} else {
			bot.replyForError(msg.Chat.ID, ErrInternalError, cfg)
		}
	} else if bot.matchByCmd(msg, "bk") {
		id := bot.parseByCmd(msg, "bk")
		if id != -1 {
			newMsg, err = bot.handlers.BookingHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), cfg)
		} else {
			bot.replyForError(msg.Chat.ID, ErrInternalError, cfg)
		}
	} else if bot.matchByCmd(msg, "ev") {
		id := bot.parseByCmd(msg, "ev")
		if id != -1 {
			newMsg, err = bot.handlers.EventHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), cfg)
		} else {
			bot.replyForError(msg.Chat.ID, ErrInternalError, cfg)
		}
	} else if bot.matchByCmd(msg, "tr") {
		id := bot.parseByCmd(msg, "tr")
		if id != -1 {
			newMsg, err = bot.handlers.TripHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), cfg)
		} else {
			bot.replyForError(msg.Chat.ID, ErrInternalError, cfg)
		}
	}
	if err == nil {
		_, err := bot.api.Send(newMsg)
		if err == nil {
			return true
		}
		bot.replyForError(msg.Chat.ID, err, cfg)
	} else {
		bot.replyForError(msg.Chat.ID, err, cfg)
	}
	return false
}

func (bot Bot) matchAddMainFieldsObject(msg *tgbotapi.Message, tripObject *TripObject, cfg *lang.Cfg) bool {
	if bot.matchByText(msg, fmt.Sprintf("add_%s", tripObject.Label), []string{tripObject.Cfg.One}) {
		countObjects, err := bot.handlers.CountFor(tripObject.Label, bot.ctx, msg.Chat.ID)
		if err != nil {
			bot.replyForError(msg.Chat.ID, err, cfg)
			return true
		}
		if countObjects >= MaxObjects {
			objectName := strings.ToLower(tripObject.Cfg.Many)
			bot.replyWithKeyboard(msg.Chat.ID, fmt.Sprintf(cfg.Errors.ObjectsLimit, objectName, MaxObjects, objectName), [][]string{{cfg.Buttons.Cancel}}, cfg)
			return true
		}
		remainingUploads, err := bot.userService.GetRemaining(bot.ctx, msg.Chat.ID)
		fileText := ""
		if err == nil && tripObject.Label != TripLabel {
			fileText = fmt.Sprintf(cfg.AddAsFile, remainingUploads)
		}
		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s\n\n- %s\n\n%s\n\n%s", tripObject.Cfg.AddMainFields, strings.Join(tripObject.Cfg.MainFields, "\n- "), tripObject.Cfg.AddExample.Main, fileText),
			[][]string{{cfg.Buttons.Cancel}}, cfg)
		bot.userService.SetState(msg.Chat.ID, service.AddInfo.For(string(tripObject.Label)))
		return true
	}
	return false
}

func (bot Bot) matchAddOtherFieldsObject(msg *tgbotapi.Message, tripObject *TripObject, cfg *lang.Cfg) bool {
	possibleCurrentStage := service.AddInfo.For(string(tripObject.Label)).Stage
	if bot.matchByStage(msg, possibleCurrentStage) {
		if msg.IsCommand() {
			return false
		}
		if msg.Text == "" && msg.Document == nil {
			bot.replyWithKeyboard(msg.Chat.ID, cfg.Errors.IncorrectInput, [][]string{{cfg.Buttons.Cancel}}, cfg)
			return true
		}
		if msg.Document != nil && tripObject.Label != TripLabel {
			if err := bot.userService.TryConsume(bot.ctx, msg.Chat.ID); err != nil {
				bot.replyForError(msg.Chat.ID, err, cfg)
				return true
			}
			bot.documentHandler.Handle(bot.ctx, msg, tripObject.Label)
			bot.userService.SetState(msg.Chat.ID, service.Default)
			bot.replyWithKeyboard(msg.Chat.ID, cfg.DocumentInProcess, BMain(cfg), cfg)
			return true
		}
		mainFieldsParsed, err := selectMainFieldsParser(msg.Text, tripObject.Label, msg.Chat.ID)
		if err != nil {
			bot.replyForError(msg.Chat.ID, err, cfg)
			return true
		}
		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s\n\n- %s\n\n%s\n\n%s", tripObject.Cfg.AddOtherFields, strings.Join(tripObject.Cfg.OtherFields, "\n- "), tripObject.Cfg.AddExample.Other, tripObject.Cfg.AddOtherFieldsTips),
			[][]string{{cfg.Buttons.Skip, cfg.Buttons.Cancel}}, cfg)
		bot.userService.SetState(msg.Chat.ID, service.Added.For(string(tripObject.Label)).With(mainFieldsParsed))
		return true
	}
	return false
}

func (bot Bot) matchAddedObject(msg *tgbotapi.Message, tripObject *TripObject, cfg *lang.Cfg) bool {
	possibleCurrentStage := service.Added.For(string(tripObject.Label)).Stage
	if bot.matchByStage(msg, possibleCurrentStage) {
		if msg.IsCommand() {
			return false
		}
		if msg.Text == "" && msg.Document == nil {
			bot.replyWithKeyboard(msg.Chat.ID, cfg.Errors.IncorrectInput, [][]string{{cfg.Buttons.Cancel}}, cfg)
			return true
		}
		currentContext := bot.userService.GetState(msg.Chat.ID).Context
		if msg.Text != cfg.Buttons.Skip {
			otherFieldsParsed, err := selectOtherFieldsParser(msg.Text, tripObject.Label, currentContext)
			if err == ErrInternalError {
				bot.replyForError(msg.Chat.ID, err, cfg)
				return true
			}
			if err != nil {
				bot.replyForError(msg.Chat.ID, err, cfg)
				return true
			}
			currentContext = otherFieldsParsed
		}

		added, err := bot.handlers.AddFor(tripObject.Label, bot.ctx, currentContext)

		if err != nil {
			bot.replyForError(msg.Chat.ID, err, cfg)
			return true
		}
		showAdded, err := bot.handlers.ShowSingleFor(bot.ctx, tripObject.Label, added.GetID(), msg.Chat.ID, cfg)
		if err != nil {
			bot.replyForError(msg.Chat.ID, err, cfg)
			return true
		}
		bot.api.Send(*showAdded)
		bot.replyWithKeyboard(msg.Chat.ID, fmt.Sprintf("%s", cfg.Added),
			[][]string{
				{cfg.Buttons.Add},
				{cfg.TripObjects.Ticket.Many, cfg.TripObjects.Booking.Many, cfg.TripObjects.Event.Many, cfg.TripObjects.Trip.Many},
				{cfg.Buttons.Settings},
			}, cfg)
		bot.userService.SetState(msg.Chat.ID, service.Default)
		return true
	}
	return false
}

func processTripObjects(cfg *lang.Cfg, f func(to TripObject) bool) bool {
	result := false
	tripObjects := TripObjects(cfg)
	for _, to := range tripObjects {
		result = result || f(to)
	}
	return result
}

func isEditReply(text string, cfg *lang.Cfg) bool {
	_, err := handler.ParseEditContext(text, cfg)
	return err == nil
}

func (bot Bot) handleUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic on update %d: %v\n%s", update.UpdateID, r, debug.Stack())
		}
	}()

	switch {
	case update.CallbackQuery != nil:
		callback := update.CallbackQuery

		_, err := bot.api.Request(
			tgbotapi.NewCallback(callback.ID, ""),
		)
		if err != nil {
			log.Printf("callback answer: %v", err)
		}

		langCfg := bot.userService.GetLanguageCfg(bot.ctx, callback.From.ID)
		msgs, err := bot.callbackHandler.Handle(bot.ctx, callback, langCfg)
		if err != nil {
			bot.replyForError(callback.From.ID, err, langCfg)
			log.Printf("callback: %v", err)
			break
		}

		for _, msg := range msgs {
			if msg != nil {
				if _, err := bot.api.Send(msg); err != nil {
					bot.replyForError(callback.From.ID, err, langCfg)
					log.Printf("callback edit: %v", err)
				}
			}
		}
	case update.Message != nil:
		{
			msg := update.Message
			log.Printf("[%s] %s", msg.From.UserName, msg.Text)
			langCfg := bot.userService.GetLanguageCfg(bot.ctx, msg.Chat.ID)

			if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil && msg.ReplyToMessage.From.ID == bot.api.Self.ID && isEditReply(msg.ReplyToMessage.Text, langCfg) {
				{
					message, err := bot.handlers.HandleEditReply(bot.ctx, update.Message, langCfg)
					if err != nil {
						bot.replyForError(update.Message.Chat.ID, err, langCfg)
					} else {
						if message != nil {
							if _, err := bot.api.Send(message); err != nil {
								bot.replyForError(update.Message.Chat.ID, err, langCfg)
							}
						}
					}
				}
			} else if bot.matchByText(msg, "menu", []string{langCfg.Buttons.Back, langCfg.Buttons.Cancel}) {
				bot.userService.SetState(msg.Chat.ID, service.Default)
				bot.replyWithKeyboard(msg.Chat.ID, langCfg.Menu, BMain(langCfg), langCfg)
			} else if bot.matchByText(msg, "settings", []string{langCfg.Buttons.Settings}) {
				bot.userService.SetState(msg.Chat.ID, service.Default)
				bot.replyWithKeyboard(msg.Chat.ID, langCfg.Menu, [][]string{{langCfg.Buttons.Lang.Lang, langCfg.Buttons.Back}}, langCfg)
			} else if bot.matchByText(msg, "add", []string{langCfg.Buttons.Add}) {
				bot.userService.SetState(msg.Chat.ID, service.Default)
				bot.replyWithKeyboard(
					msg.Chat.ID,
					langCfg.Add,
					[][]string{{langCfg.TripObjects.Ticket.One, langCfg.TripObjects.Booking.One, langCfg.TripObjects.Event.One, langCfg.TripObjects.Trip.One}, {langCfg.Buttons.Back}}, langCfg)
			} else if processTripObjects(langCfg, func(tripObject TripObject) bool { return bot.matchListObjects(msg, &tripObject, langCfg) }) {
			} else if processTripObjects(langCfg, func(tripObject TripObject) bool { return bot.matchAddMainFieldsObject(msg, &tripObject, langCfg) }) {
			} else if processTripObjects(langCfg, func(tripObject TripObject) bool { return bot.matchAddOtherFieldsObject(msg, &tripObject, langCfg) }) {
			} else if processTripObjects(langCfg, func(tripObject TripObject) bool { return bot.matchAddedObject(msg, &tripObject, langCfg) }) {
			} else if bot.matchByText(msg, "start", []string{}) {
				bot.userService.GetOrCreate(bot.ctx, msg.Chat.ID)
				bot.userService.SetState(msg.Chat.ID, service.Default)
				bot.replyWithKeyboard(
					msg.Chat.ID,
					fmt.Sprintf("%s\n\n%s", langCfg.Welcome, langCfg.ChooseLang),
					[][]string{
						{langCfg.Buttons.Lang.Ru, langCfg.Buttons.Lang.En}}, langCfg)
			} else if bot.matchByText(msg, "lang", []string{langCfg.Buttons.Lang.Lang}) {
				bot.userService.SetState(msg.Chat.ID, service.Default)
				bot.replyWithKeyboard(
					msg.Chat.ID,
					langCfg.ChooseLang,
					[][]string{
						{langCfg.Buttons.Lang.Ru, langCfg.Buttons.Lang.En}}, langCfg)
			} else if bot.matchByText(msg, "ru", []string{langCfg.Buttons.Lang.Ru}) {
				bot.userService.SetOrUpdateLang(bot.ctx, msg.Chat.ID, lang.LanguageRu)
				newLangCfg := bot.userService.GetLanguageCfg(bot.ctx, msg.Chat.ID)
				bot.replyWithKeyboard(msg.Chat.ID, newLangCfg.Menu, BMain(newLangCfg), newLangCfg)
			} else if bot.matchByText(msg, "en", []string{langCfg.Buttons.Lang.En}) {
				bot.userService.SetOrUpdateLang(bot.ctx, msg.Chat.ID, lang.LanguageEn)
				newLangCfg := bot.userService.GetLanguageCfg(bot.ctx, msg.Chat.ID)
				bot.replyWithKeyboard(msg.Chat.ID, newLangCfg.Menu, BMain(newLangCfg), newLangCfg)
			} else if bot.matchShowSingle(msg, langCfg) {
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, langCfg.UnknownCmd)
				msg.ReplyToMessageID = update.Message.MessageID
				bot.api.Send(msg)
			}
		}
	}
}

func (bot *Bot) Run(shutdownCtx context.Context) {
	log.Printf("Authorized on account %s", bot.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.api.GetUpdatesChan(u)

	go func() {
		<-shutdownCtx.Done()
		bot.api.StopReceivingUpdates()
	}()

	for update := range updates {
		if bot.ctx.Err() != nil {
			break
		}
		bot.handleUpdate(update)
	}
}
