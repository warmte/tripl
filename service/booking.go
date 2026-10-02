package service

import (
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/internal/repo"
)

var _ Service[Booking] = &BookingService{}

type BookingService struct {
	CommonService[Booking]
}

func NewBookingService(
	repo repo.BookingRepository,
) *BookingService {
	return &BookingService{
		CommonService[Booking]{
			repo: repo,
		},
	}
}
