package handler

import (
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/service"
)

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
