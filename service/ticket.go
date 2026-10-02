package service

import (
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/internal/repo"
)

var _ Service[Ticket] = &TicketService{}

type TicketService struct {
	CommonService[Ticket]
}

func NewTicketService(
	repo repo.TicketRepository,
) *TicketService {
	return &TicketService{
		CommonService[Ticket]{
			repo: repo,
		},
	}
}
