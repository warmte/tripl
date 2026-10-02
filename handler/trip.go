package handler

import (
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/service"
)

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
