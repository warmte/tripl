package handler

import (
	"context"
	"errors"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/warmte/tripl/ai"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/document"
	"github.com/warmte/tripl/service"
)

const workersCount = 2

type DocumentResult struct {
	ChatID int64
	Label  Label
	Result TripObjectView
	Err    error
}

type documentTask struct {
	ctx    context.Context
	chatID int64
	doc    *tgbotapi.Document
	label  Label
}

type DocumentHandler struct {
	api    *tgbotapi.BotAPI
	mu     sync.Mutex
	queue  []documentTask
	notify chan struct{}

	pdfService  *document.PDFService
	pdfParser   *document.PDFParser
	aiService   *ai.AIService
	userService *service.UserService

	ticketHandler  *TicketHandler
	bookingHandler *BookingHandler
	eventHandler   *EventHandler

	results chan DocumentResult

	wg sync.WaitGroup
}

func NewDocumentHandler(
	api *tgbotapi.BotAPI,
	pdfService *document.PDFService,
	pdfParser *document.PDFParser,
	aiService *ai.AIService,
	userService *service.UserService,
	ticketHandler *TicketHandler,
	bookingHandler *BookingHandler,
	eventHandler *EventHandler,
) *DocumentHandler {
	h := &DocumentHandler{
		api:            api,
		pdfService:     pdfService,
		pdfParser:      pdfParser,
		aiService:      aiService,
		userService:    userService,
		ticketHandler:  ticketHandler,
		bookingHandler: bookingHandler,
		eventHandler:   eventHandler,
		notify:         make(chan struct{}, 1),
		results:        make(chan DocumentResult),
	}

	h.wg.Add(workersCount)

	for i := 0; i < workersCount; i++ {
		go h.worker()
	}

	return h
}

func (h *DocumentHandler) Handle(
	ctx context.Context,
	message *tgbotapi.Message,
	label Label,
) error {
	if message == nil {
		return errors.New("message is nil")
	}

	if message.Document == nil {
		return errors.New("document is nil")
	}

	task := documentTask{
		ctx:    ctx,
		chatID: message.Chat.ID,
		doc:    message.Document,
		label:  label,
	}

	h.mu.Lock()
	h.queue = append(h.queue, task)
	h.mu.Unlock()

	select {
	case h.notify <- struct{}{}:
	default:
	}

	return nil
}

func (h *DocumentHandler) worker() {
	defer h.wg.Done()

	for {
		h.mu.Lock()

		if len(h.queue) > 0 {
			task := h.queue[0]
			h.queue = h.queue[1:]
			h.mu.Unlock()

			h.process(task)
			continue
		}

		h.mu.Unlock()
		<-h.notify
	}
}

func (h *DocumentHandler) process(task documentTask) {
	result := DocumentResult{
		ChatID: task.chatID,
		Label:  task.label,
	}

	if err := task.ctx.Err(); err != nil {
		result.Err = err
		h.results <- result
		return
	}

	pdf, err := h.pdfService.Download(task.ctx, task.doc)
	if err != nil {
		result.Err = fmt.Errorf("download PDF: %w", err)
		h.results <- result
		return
	}
	defer h.pdfService.Remove(pdf)

	text, err := h.pdfParser.Parse(task.ctx, pdf)
	if err != nil {
		result.Err = fmt.Errorf("parse PDF: %w", err)
		h.results <- result
		return
	}

	var extracted TripObjectView
	switch task.label {
	case TicketLabel:
		ticket, err1 := h.aiService.ExtractTicket(task.ctx, task.chatID, text)
		extracted, err = *ticket, err1
	case BookingLabel:
		booking, err1 := h.aiService.ExtractBooking(task.ctx, task.chatID, text)
		extracted, err = *booking, err1
	case EventLabel:
		event, err1 := h.aiService.ExtractEvent(task.ctx, task.chatID, text)
		extracted, err = *event, err1
	default:
		extracted, err = nil, fmt.Errorf("unknown label")
	}

	if err != nil {
		result.Err = fmt.Errorf("extract document: %w", err)
		h.results <- result
		return
	}

	result.Result = extracted
	h.results <- result
}

func (h *DocumentHandler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case result, ok := <-h.results:
			if !ok {
				return
			}
			h.handleResult(ctx, result)
		}
	}
}

func (h *DocumentHandler) handleResult(ctx context.Context, result DocumentResult) {
	cfg := h.userService.GetLanguage(ctx, result.ChatID)
	if result.Err != nil {
		h.api.Send(tgbotapi.NewMessage(result.ChatID, fmt.Sprintf(cfg.Errors.DocumentFailed)))
		return
	}

	switch result.Label {
	case TicketLabel:
		added, err := h.ticketHandler.Add(ctx, result.Result)
		if err != nil {
			h.api.Send(tgbotapi.NewMessage(result.ChatID, fmt.Sprintf(cfg.Errors.DocumentFailed)))
			return
		}
		showAdded, err := h.ticketHandler.ShowSingle(ctx, result.ChatID, added.GetID(), cfg)
		if err != nil {
			h.api.Send(tgbotapi.NewMessage(result.ChatID, fmt.Sprintf(cfg.Errors.DocumentFailed)))
			return
		}
		h.api.Send(*showAdded)

	case BookingLabel:
		added, err := h.bookingHandler.Add(ctx, result.Result)
		if err != nil {
			h.api.Send(tgbotapi.NewMessage(result.ChatID, fmt.Sprintf(cfg.Errors.DocumentFailed)))
			return
		}
		showAdded, err := h.bookingHandler.ShowSingle(ctx, result.ChatID, added.GetID(), cfg)
		if err != nil {
			h.api.Send(tgbotapi.NewMessage(result.ChatID, fmt.Sprintf(cfg.Errors.DocumentFailed)))
			return
		}
		h.api.Send(*showAdded)

	case EventLabel:
		added, err := h.eventHandler.Add(ctx, result.Result)
		if err != nil {
			h.api.Send(tgbotapi.NewMessage(result.ChatID, fmt.Sprintf(cfg.Errors.DocumentFailed)))
			return
		}
		showAdded, err := h.eventHandler.ShowSingle(ctx, result.ChatID, added.GetID(), cfg)
		if err != nil {
			h.api.Send(tgbotapi.NewMessage(result.ChatID, fmt.Sprintf(cfg.Errors.DocumentFailed)))
			return
		}
		h.api.Send(*showAdded)
		h.api.Send(tgbotapi.NewMessage(result.ChatID, cfg.DocumentFinished))
	}
}
