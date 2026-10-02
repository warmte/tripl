package handler

import (
	. "github.com/warmte/tripl/common"
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
