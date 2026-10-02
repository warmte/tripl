package handler

import (
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/service"
)

type EventHandler struct {
	CommonHandler[Event]
}

func NewEventHandler(
	service service.EventService,
	tripService service.TripService,
) *EventHandler {
	return &EventHandler{
		CommonHandler[Event]{
			label:   EventLabel,
			service: service,
			tripService: tripService,
		},
	}
}
