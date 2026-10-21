package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/warmte/tripl/ai"
	"github.com/warmte/tripl/bot"
	"github.com/warmte/tripl/document"
	"github.com/warmte/tripl/handler"
	"github.com/warmte/tripl/internal/db"
	"github.com/warmte/tripl/internal/repo"
	"github.com/warmte/tripl/scheduler"
	"github.com/warmte/tripl/service"
)

type redactWriter struct {
	out    io.Writer
	secret []byte
}

func (w redactWriter) Write(p []byte) (int, error) {
	if _, err := w.out.Write(bytes.ReplaceAll(p, w.secret, []byte("<redacted>"))); err != nil {
		return 0, err
	}
	return len(p), nil
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is empty")
	}
	w := redactWriter{out: os.Stderr, secret: []byte(token)}
	log.SetOutput(w)
	tgbotapi.SetLogger(log.New(w, "", log.LstdFlags))

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	ctx, cancelWork := context.WithCancel(context.Background())
	defer cancelWork()

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

	callbackHandler := handler.NewCallbackHandler(ticketsHandler, bookingsHandler, eventsHandler, tripsHandler, userService)
	handlers := handler.NewHandlers(ticketsHandler, bookingsHandler, eventsHandler, tripsHandler)

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("create bot api: %v", err)
	}

	aiService := ai.NewAIService(
		"http://llama-server:8080/v1/chat/completions",
		"Qwen3-4B-Q4_K_M.gguf",
	)
	pdfService := document.NewPDFService(api)
	pdfParser := document.NewPDFParser()

	documentHandler := handler.NewDocumentHandler(api, pdfService, pdfParser, aiService, userService, ticketsHandler, bookingsHandler, eventsHandler)

	b := bot.NewBot(api, ctx, userService, callbackHandler, documentHandler, handlers)

	notificationSender := scheduler.NewNotificationSender(api, userService, ticketsHandler, bookingsHandler, eventsHandler)
	notificationScheduler, err := scheduler.NewScheduler(*notificationSender)
	if err != nil {
		log.Fatalf("create bot api: %v", err)
	}

	resultDone := make(chan struct{})

	go func() {
		defer close(resultDone)
		documentHandler.RunHandleResult(ctx)
	}()

	go notificationScheduler.Run(signalCtx)
	go userService.RunDailyUsageCleanup(signalCtx)

	b.Run(signalCtx)

	documentHandler.Shutdown()
	<-resultDone
	cancelWork()
}
