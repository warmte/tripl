package main

import (
	"context"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/warmte/tripl/bot"
	"github.com/warmte/tripl/handler"
	"github.com/warmte/tripl/internal/db"
	"github.com/warmte/tripl/internal/repo"
	"github.com/warmte/tripl/scheduler"
	"github.com/warmte/tripl/service"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	ctx := context.Background()

	pool, err := db.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	userRepo := repo.NewUserRepository(pool)
	ticketsRepo := repo.NewTicketRepository(pool)
	bookingsRepo := repo.NewBookingRepository(pool)
	eventsRepo := repo.NewEventRepository(pool)
	tripsRepo := repo.NewTripRepository(pool)

	userService := service.NewUserService(*userRepo)
	ticketsService := service.NewTicketService(*ticketsRepo)
	bookingsService := service.NewBookingService(*bookingsRepo)
	eventsService := service.NewEventService(*eventsRepo)
	tripsService := service.NewTripService(*tripsRepo)

	ticketsHandler := handler.NewTicketHandler(*ticketsService, *tripsService)
	bookingsHandler := handler.NewBookingHandler(*bookingsService, *tripsService)
	eventsHandler := handler.NewEventHandler(*eventsService, *tripsService)
	tripsHandler := handler.NewTripHandler(*tripsService)

	callbackHandler := handler.NewCallbackHandler(ticketsHandler, bookingsHandler, eventsHandler, tripsHandler)
	handlers := handler.NewHandlers(ticketsHandler, bookingsHandler, eventsHandler, tripsHandler)

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic("err create api: %w", err)
	}

	b := bot.NewBot(api, ctx, userService, callbackHandler, handlers)

	notificationSender := scheduler.NewNotificationSender(api, userService, ticketsHandler, bookingsHandler, eventsHandler)
	sched, err := scheduler.New(*notificationSender)
	if err != nil {
		log.Panic("err create scheduler: %w", err)
	}

	go sched.Run(ctx)
	b.Run()
}
