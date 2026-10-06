package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/warmte/tripl/handler"
	"github.com/warmte/tripl/service"
)

type NotificationSender struct {
	api            *tgbotapi.BotAPI
	userService    *service.UserService
	ticketHandler  *handler.TicketHandler
	bookingHandler *handler.BookingHandler
	eventHandler   *handler.EventHandler
}

func NewNotificationSender(
	api *tgbotapi.BotAPI,
	userService *service.UserService,
	ticketHandler *handler.TicketHandler,
	bookingHandler *handler.BookingHandler,
	eventHandler *handler.EventHandler,
) *NotificationSender {
	return &NotificationSender{
		api:            api,
		userService:    userService,
		ticketHandler:  ticketHandler,
		bookingHandler: bookingHandler,
		eventHandler:   eventHandler,
	}
}

func (s *NotificationSender) SendTomorrow(
	ctx context.Context,
	date time.Time,
	location *time.Location,
) error {
	users, err := s.userService.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("get users: %w", err)
	}
	for _, user := range users {
		if err := s.sendUserTomorrow(ctx, user.ID, location); err != nil {
			log.Printf(
				"scheduler: user %d: send tomorrow notification: %v",
				user.ID,
				err,
			)
		}
	}

	return nil
}

func (s *NotificationSender) sendUserTomorrow(
	ctx context.Context,
	userID int64,
	location *time.Location,
) error {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return fmt.Errorf("load Moscow timezone: %w", err)
	}

	cfg := s.userService.GetLanguage(ctx, userID)

	now := time.Now().In(location)
	tomorrow := now.AddDate(0, 0, 1)
	header := fmt.Sprintf("📅 %s — %s", cfg.TomorrowPlans, tomorrow.Format("02.01.2006"))

	message := tgbotapi.NewMessage(userID, header)
	s.api.Send(message)

	tomorrowTickets, err := s.ticketHandler.ShowTomorrowList(ctx, userID, location, cfg)
	if err == nil {
		if _, err := s.api.Send(tomorrowTickets); err != nil {
			log.Printf("send tomorrow tickets: %v", err)
		}
	}

	tomorrowBookings, err := s.bookingHandler.ShowTomorrowList(ctx, userID, location, cfg)
	if err == nil {
		if _, err := s.api.Send(tomorrowBookings); err != nil {
			log.Printf("send tomorrow bookings: %v", err)
		}
	}

	tomorrowEvents, err := s.eventHandler.ShowTomorrowList(ctx, userID, location, cfg)
	if err == nil {
		if _, err := s.api.Send(tomorrowEvents); err != nil {
			log.Printf("send tomorrow events: %v", err)
		}
	}

	return nil
}
