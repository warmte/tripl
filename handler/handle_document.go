package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	Result []TripObjectView
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

	stopCh       chan struct{}
	stopped      bool
	shutdownOnce sync.Once
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
		stopCh:         make(chan struct{}),
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
	if h.stopped {
		h.mu.Unlock()
		return errors.New("document handler is shutting down")
	}
	h.queue = append(h.queue, task)
	h.mu.Unlock()

	select {
	case h.notify <- struct{}{}:
	default:
	}

	return nil
}

func (h *DocumentHandler) Shutdown() {
	h.shutdownOnce.Do(func() {
		h.mu.Lock()
		h.stopped = true
		close(h.stopCh)
		h.mu.Unlock()
		h.wg.Wait()
		close(h.results)
	})
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

		if h.stopped {
			h.mu.Unlock()
			return
		}

		h.mu.Unlock()

		select {
		case <-h.notify:
		case <-h.stopCh:
		}
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

	lang := h.userService.GetLanguage(task.ctx, task.chatID)

	err = nil
	extracted := []TripObjectView{}
	switch task.label {
	case TicketLabel:
		tickets, err1 := h.aiService.ExtractTicket(task.ctx, task.chatID, text, lang)
		if err1 == nil {
			for _, ticket := range tickets {
				extracted = append(extracted, ticket)
			}
		} else {
			err = err1
		}
	case BookingLabel:
		booking, err1 := h.aiService.ExtractBooking(task.ctx, task.chatID, text, lang)
		if err1 == nil {
			extracted = []TripObjectView{booking}
		} else {
			err = err1
		}
	case EventLabel:
		event, err1 := h.aiService.ExtractEvent(task.ctx, task.chatID, text, lang)
		if err1 == nil {
			extracted = []TripObjectView{event}
		} else {
			err = err1
		}
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

func (h *DocumentHandler) RunHandleResult(ctx context.Context) {
	for result := range h.results {
		h.handleResult(ctx, result)
	}
}

func (h *DocumentHandler) handleResult(ctx context.Context, result DocumentResult) {
	cfg := h.userService.GetLanguageCfg(ctx, result.ChatID)
	if result.Err != nil {
		h.api.Send(tgbotapi.NewMessage(result.ChatID, cfg.Errors.DocumentFailed))
		log.Printf("document error: %v", result.Err)
		return
	}

	msgs := []tgbotapi.MessageConfig{}

	switch result.Label {
	case TicketLabel:
		for _, t := range result.Result {
			added, err := h.ticketHandler.Add(ctx, t)
			if err != nil {
				h.api.Send(tgbotapi.NewMessage(result.ChatID, cfg.Errors.DocumentFailed))
				log.Printf("document error: %v", err)
				return
			}
			showAdded, err := h.ticketHandler.ShowSingle(ctx, result.ChatID, added.GetID(), cfg)
			if err != nil {
				h.api.Send(tgbotapi.NewMessage(result.ChatID, cfg.Errors.DocumentFailed))
				log.Printf("document error: %v", err)
				return
			}
			msgs = append(msgs, *showAdded)
		}

	case BookingLabel:
		for _, b := range result.Result {
			added, err := h.bookingHandler.Add(ctx, b)
			if err != nil {
				h.api.Send(tgbotapi.NewMessage(result.ChatID, cfg.Errors.DocumentFailed))
				log.Printf("document error: %v", err)
				return
			}
			showAdded, err := h.bookingHandler.ShowSingle(ctx, result.ChatID, added.GetID(), cfg)
			if err != nil {
				h.api.Send(tgbotapi.NewMessage(result.ChatID, cfg.Errors.DocumentFailed))
				log.Printf("document error: %v", err)
				return
			}
			msgs = append(msgs, *showAdded)
		}

	case EventLabel:
		for _, e := range result.Result {
			added, err := h.eventHandler.Add(ctx, e)
			if err != nil {
				h.api.Send(tgbotapi.NewMessage(result.ChatID, cfg.Errors.DocumentFailed))
				log.Printf("document error: %v", err)
				return
			}
			showAdded, err := h.eventHandler.ShowSingle(ctx, result.ChatID, added.GetID(), cfg)
			if err != nil {
				h.api.Send(tgbotapi.NewMessage(result.ChatID, cfg.Errors.DocumentFailed))
				log.Printf("document error: %v", err)
				return
			}
			msgs = append(msgs, *showAdded)
		}
	}
	if len(msgs) > 0 {
		for _, msg := range msgs {
			h.api.Send(msg)
		}
		h.api.Send(tgbotapi.NewMessage(result.ChatID, cfg.DocumentFinished))
	}
}
