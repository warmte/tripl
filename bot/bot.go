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
	langCfg         *lang.Cfg
	ctx             context.Context
	userService     *service.UserService
	ticketsService  *service.TicketService
	bookingsService *service.BookingService
	eventsService   *service.EventService
	tripsService    *service.TripService
	callbackHandler *handler.CallbackHandler
	handlers        *handler.Handlers
}

func NewBot(
	botId string,
	ctx context.Context,
	userService *service.UserService,
	ticketsService *service.TicketService,
	bookingsService *service.BookingService,
	eventsService *service.EventService,
	tripsService *service.TripService,
	callbackHandler *handler.CallbackHandler,
	handlers *handler.Handlers,
) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(botId)
	if err != nil {
		return nil, err
	} else {
		langCfg := lang.Get(lang.LanguageRu)
		return &Bot{api, &langCfg, ctx, userService, ticketsService, bookingsService, eventsService, tripsService, callbackHandler, handlers}, nil
	}
}

func (bot Bot) replyWithKeyboard(chatId int64, text string, buttons [][]string) {
	raws := BuildKeyboard(buttons, bot.langCfg)
	replyMsg := tgbotapi.NewMessage(chatId, lang.EscapeMarkdownV2(text))
	replyMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(raws...)
	replyMsg.ParseMode = tgbotapi.ModeMarkdownV2
	bot.api.Send(replyMsg)
}

func (bot Bot) replyForIncorrectInput(chatId int64, err error) {
	bot.replyWithKeyboard(chatId, GetErrorMessage(err, *bot.langCfg), [][]string{{bot.langCfg.Buttons.Cancel}})
}

func (bot Bot) replyForInternalError(chatId int64) {
	bot.replyWithKeyboard(chatId, bot.langCfg.Errors.InternalError, [][]string{{bot.langCfg.Buttons.Cancel}})
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
		newMsg, err := bot.handlers.ShowListFor(tripObject.Label, bot.ctx, msg.Chat.ID, ActualListType, bot.langCfg)
		if err == nil {
			bot.api.Send(*newMsg)
		}

		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s. %s", tripObject.Cfg.Actual, bot.langCfg.Info),
			[][]string{{tripObject.Cfg.Archive, bot.langCfg.Buttons.Back}})

		return true
	}

	if bot.matchByText(
		msg,
		fmt.Sprintf("%s_%s", ArchiveListType, tripObject.Label),
		[]string{tripObject.Cfg.Many, tripObject.Cfg.Archive},
	) {
		newMsg, err := bot.handlers.ShowListFor(tripObject.Label, bot.ctx, msg.Chat.ID, ArchiveListType, bot.langCfg)
		if err == nil {
			bot.api.Send(*newMsg)
		}

		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s. %s", tripObject.Cfg.Archive, bot.langCfg.Info),
			[][]string{{tripObject.Cfg.Actual, bot.langCfg.Buttons.Back}})

		return true
	}

	return false
}

func (bot Bot) matchShowSingle(msg *tgbotapi.Message) bool {
	if bot.matchByCmd(msg, "tt") {
		id := bot.parseByCmd(msg, "tt")
		if id != -1 {
			newMsg, err := bot.handlers.TicketHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), bot.langCfg)
			if err == nil {
				bot.api.Send(*newMsg)
			}
		}
		return true
	} else if bot.matchByCmd(msg, "bk") {
		id := bot.parseByCmd(msg, "bk")
		if id != -1 {
			newMsg, err := bot.handlers.BookingHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), bot.langCfg)
			if err == nil {
				bot.api.Send(*newMsg)
			}
		}
		return true
	} else if bot.matchByCmd(msg, "ev") {
		id := bot.parseByCmd(msg, "ev")
		if id != -1 {
			newMsg, err := bot.handlers.EventHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), bot.langCfg)
			if err == nil {
				bot.api.Send(*newMsg)
			}
		}
		return true
	} else if bot.matchByCmd(msg, "tr") {
		id := bot.parseByCmd(msg, "tr")
		if id != -1 {
			newMsg, err := bot.handlers.TripHandler.ShowSingle(bot.ctx, msg.Chat.ID, int64(id), bot.langCfg)
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
		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s\n\n- %s\n\n%s", tripObject.Cfg.AddMainFields, strings.Join(tripObject.Cfg.MainFields, "\n- "), tripObject.Cfg.AddExample.Main),
			[][]string{{bot.langCfg.Buttons.Cancel}})
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
			bot.replyForIncorrectInput(msg.Chat.ID, err)
			return true
		}
		bot.replyWithKeyboard(
			msg.Chat.ID,
			fmt.Sprintf("%s\n\n- %s\n\n%s\n\n%s", tripObject.Cfg.AddOtherFields, strings.Join(tripObject.Cfg.OtherFields, "\n- "), tripObject.Cfg.AddExample.Other, tripObject.Cfg.AddOtherFieldsTips),
			[][]string{{bot.langCfg.Buttons.Skip, bot.langCfg.Buttons.Cancel}})
		bot.userService.SetState(msg.Chat.ID, service.Added.For(string(tripObject.Label)).With(mainFieldsParsed))
		return true
	}
	return false
}

func (bot Bot) matchAddedObject(msg *tgbotapi.Message, tripObject *TripObject) bool {
	possibleCurrentStage := service.Added.For(string(tripObject.Label)).Stage
	if bot.matchByStage(msg, possibleCurrentStage) {
		currentContext := bot.userService.GetState(msg.Chat.ID).Context
		if msg.Text != bot.langCfg.Buttons.Skip {
			otherFieldsParsed, err := selectOtherFieldsParser(msg.Text, tripObject.Label, currentContext)
			if err == ErrInternalError {
				bot.replyForInternalError(msg.Chat.ID)
				return true
			}
			if err != nil {
				bot.replyForIncorrectInput(msg.Chat.ID, err)
				return true
			}
			currentContext = otherFieldsParsed
		}

		added, err := bot.handlers.AddFor(tripObject.Label, bot.ctx, currentContext)

		if err != nil {
			bot.replyForInternalError(msg.Chat.ID)
			return true
		}
		showAdded, err := bot.handlers.ShowSingleFor(bot.ctx, tripObject.Label, added.GetID(), msg.Chat.ID, bot.langCfg)
		if err != nil {
			bot.replyForInternalError(msg.Chat.ID)
			return true
		}
		bot.api.Send(*showAdded)
		bot.replyWithKeyboard(msg.Chat.ID, fmt.Sprintf("%s", bot.langCfg.Added),
			[][]string{
				{bot.langCfg.Buttons.Add},
				{bot.langCfg.TripObjects.Ticket.Many, bot.langCfg.TripObjects.Booking.Many, bot.langCfg.TripObjects.Event.Many, bot.langCfg.TripObjects.Trip.Many},
				{bot.langCfg.Buttons.Settings},
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

	tripObjects := TripObjects(bot.langCfg)
	processTripObjects := func(f func(to TripObject) bool) bool {
		result := false
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

			edits, err := bot.callbackHandler.Handle(
				bot.ctx,
				callback,
				bot.langCfg,
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
				if msg.ReplyToMessage != nil {
					{
						message, err := bot.handlers.HandleEditReply(
							bot.ctx,
							update.Message,
							bot.langCfg,
						)
						if err != nil {
							bot.replyForInternalError(update.Message.Chat.ID)
						} else {
							if message != nil {
								if _, err := bot.api.Send(message); err != nil {
									bot.replyForInternalError(update.Message.Chat.ID)
								}
							}
						}
					}
				} else if bot.matchByText(msg, "menu", []string{bot.langCfg.Buttons.Back, bot.langCfg.Buttons.Cancel}) {
					bot.userService.SetState(msg.Chat.ID, service.Default)
					bot.replyWithKeyboard(msg.Chat.ID, bot.langCfg.Menu, BMain(bot.langCfg))
				} else if bot.matchByText(msg, "settings", []string{bot.langCfg.Buttons.Settings}) {
					bot.userService.SetState(msg.Chat.ID, service.Default)
					bot.replyWithKeyboard(msg.Chat.ID, bot.langCfg.Menu, [][]string{{bot.langCfg.Buttons.Lang.Lang, bot.langCfg.Buttons.Back}})
				} else if bot.matchByText(msg, "add", []string{bot.langCfg.Buttons.Add}) {
					bot.userService.SetState(msg.Chat.ID, service.Default)
					bot.replyWithKeyboard(
						msg.Chat.ID,
						bot.langCfg.Add,
						[][]string{{bot.langCfg.TripObjects.Ticket.One, bot.langCfg.TripObjects.Booking.One, bot.langCfg.TripObjects.Event.One, bot.langCfg.TripObjects.Trip.One}, {bot.langCfg.Buttons.Back}})
				} else if processTripObjects(func(tripObject TripObject) bool { return bot.matchListObjects(msg, &tripObject) }) {
				} else if processTripObjects(func(tripObject TripObject) bool { return bot.matchAddMainFieldsObject(msg, &tripObject) }) {
				} else if processTripObjects(func(tripObject TripObject) bool { return bot.matchAddOtherFieldsObject(msg, &tripObject) }) {
				} else if processTripObjects(func(tripObject TripObject) bool { return bot.matchAddedObject(msg, &tripObject) }) {
				} else if bot.matchByText(msg, "start", []string{}) {
					bot.userService.SetState(msg.Chat.ID, service.Default)
					bot.replyWithKeyboard(
						msg.Chat.ID,
						fmt.Sprintf("%s\n\n%s", bot.langCfg.Welcome, bot.langCfg.ChooseLang),
						[][]string{
							{bot.langCfg.Buttons.Lang.Ru, bot.langCfg.Buttons.Lang.En}})
				} else if bot.matchByText(msg, "lang", []string{bot.langCfg.Buttons.Lang.Lang}) {
					bot.userService.SetState(msg.Chat.ID, service.Default)
					bot.replyWithKeyboard(
						msg.Chat.ID,
						bot.langCfg.ChooseLang,
						[][]string{
							{bot.langCfg.Buttons.Lang.Ru, bot.langCfg.Buttons.Lang.En}})
				} else if bot.matchByText(msg, "ru", []string{bot.langCfg.Buttons.Lang.Ru}) {
					bot.userService.SetOrUpdateLang(bot.ctx, msg.Chat.ID, lang.LanguageRu)
					newCfg := lang.Get(lang.LanguageRu)
					bot.langCfg = &newCfg
					bot.replyWithKeyboard(msg.Chat.ID, bot.langCfg.Menu, BMain(bot.langCfg))
				} else if bot.matchByText(msg, "en", []string{bot.langCfg.Buttons.Lang.En}) {
					bot.userService.SetOrUpdateLang(bot.ctx, msg.Chat.ID, lang.LanguageEn)
					newCfg := lang.Get(lang.LanguageEn)
					bot.langCfg = &newCfg
					bot.replyWithKeyboard(msg.Chat.ID, bot.langCfg.Menu, BMain(bot.langCfg))
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
