package service

import (
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/internal/repo"
)

var _ Service[Event] = &EventService{}

type EventService struct {
	CommonService[Event]
}

func NewEventService(
	repo repo.EventRepository,
) *EventService {
	return &EventService{
		CommonService[Event]{
			repo: repo,
		},
	}
}
