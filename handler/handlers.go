package handler

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
	"github.com/warmte/tripl/service"
)

type TicketHandler struct {
	CommonHandler[Ticket]
}

func NewTicketHandler(
	service service.TicketService,
	tripService service.TripService,
) *TicketHandler {
	return &TicketHandler{
		CommonHandler[Ticket]{
			label:       TicketLabel,
			service:     service,
			tripService: tripService,
		},
	}
}

type BookingHandler struct {
	CommonHandler[Booking]
}

func NewBookingHandler(
	service service.BookingService,
	tripService service.TripService,
) *BookingHandler {
	return &BookingHandler{
		CommonHandler[Booking]{
			label:       BookingLabel,
			service:     service,
			tripService: tripService,
		},
	}
}

type EventHandler struct {
	CommonHandler[Event]
}

func NewEventHandler(
	service service.EventService,
	tripService service.TripService,
) *EventHandler {
	return &EventHandler{
		CommonHandler[Event]{
			label:       EventLabel,
			service:     service,
			tripService: tripService,
		},
	}
}

type TripHandler struct {
	CommonHandler[Trip]
}

func NewTripHandler(
	service service.TripService,
) *TripHandler {
	return &TripHandler{
		CommonHandler[Trip]{
			label:   TripLabel,
			service: service,
		},
	}
}

type Handlers struct {
	TicketHandler  *TicketHandler
	BookingHandler *BookingHandler
	EventHandler   *EventHandler
	TripHandler    *TripHandler
}

func NewHandlers(
	ticketHandler *TicketHandler,
	bookingHandler *BookingHandler,
	eventHandler *EventHandler,
	tripHandler *TripHandler,
) *Handlers {
	return &Handlers{
		TicketHandler:  ticketHandler,
		BookingHandler: bookingHandler,
		EventHandler:   eventHandler,
		TripHandler:    tripHandler,
	}
}

func (h *Handlers) ShowListFor(
	label Label,
	ctx context.Context,
	userID int64,
	listType ListType,
	cfg *lang.Cfg,
) (*tgbotapi.MessageConfig, error) {
	switch label {
	case TicketLabel:
		return h.TicketHandler.ShowList(ctx, userID, listType, cfg)
	case BookingLabel:
		return h.BookingHandler.ShowList(ctx, userID, listType, cfg)
	case EventLabel:
		return h.EventHandler.ShowList(ctx, userID, listType, cfg)
	case TripLabel:
		return h.TripHandler.ShowList(ctx, userID, listType, cfg)
	default:
		return nil, ErrIncorrectInput
	}
}

func (h *Handlers) ShowSingleFor(
	ctx context.Context,
	label Label,
	objectID int64,
	userID int64,
	cfg *lang.Cfg,
) (*tgbotapi.MessageConfig, error) {
	switch label {
	case TicketLabel:
		return h.TicketHandler.ShowSingle(ctx, userID, objectID, cfg)
	case BookingLabel:
		return h.BookingHandler.ShowSingle(ctx, userID, objectID, cfg)
	case EventLabel:
		return h.EventHandler.ShowSingle(ctx, userID, objectID, cfg)
	case TripLabel:
		return h.TripHandler.ShowSingle(ctx, userID, objectID, cfg)
	default:
		return nil, ErrIncorrectInput
	}
}

func (h *Handlers) AddFor(label Label, ctx context.Context, data TripObjectView) (TripObjectView, error) {
	switch label {
	case TicketLabel:
		return h.TicketHandler.Add(ctx, data)
	case BookingLabel:
		return h.BookingHandler.Add(ctx, data)
	case EventLabel:
		return h.EventHandler.Add(ctx, data)
	case TripLabel:
		return h.TripHandler.Add(ctx, data)
	default:
		return nil, ErrIncorrectInput
	}
}

func (h *Handlers) CountFor(label Label, ctx context.Context, userID int64) (int64, error) {
	switch label {
	case TicketLabel:
		return h.TicketHandler.Count(ctx, userID)
	case BookingLabel:
		return h.BookingHandler.Count(ctx, userID)
	case EventLabel:
		return h.EventHandler.Count(ctx, userID)
	case TripLabel:
		return h.TripHandler.Count(ctx, userID)
	default:
		return 0, ErrIncorrectInput
	}
}

func (h *Handlers) HandleEditReply(
	ctx context.Context,
	msg *tgbotapi.Message,
	cfg *lang.Cfg,
) (tgbotapi.Chattable, error) {
	if msg == nil || msg.ReplyToMessage == nil {
		return nil, nil
	}

	editCtx, err := parseEditContext(msg.ReplyToMessage.Text, cfg)
	if err != nil {
		return nil, nil
	}

	switch editCtx.Label {
	case TicketLabel:
		return h.TicketHandler.HandleEditReply(ctx, msg, cfg)
	case BookingLabel:
		return h.BookingHandler.HandleEditReply(ctx, msg, cfg)
	case EventLabel:
		return h.EventHandler.HandleEditReply(ctx, msg, cfg)
	case TripLabel:
		return h.TripHandler.HandleEditReply(ctx, msg, cfg)
	default:
		return nil, nil
	}
}
