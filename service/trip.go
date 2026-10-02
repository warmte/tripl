package service

import (
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/internal/repo"
)

var _ Service[Trip] = &TripService{}

type TripService struct {
	CommonService[Trip]
}

func NewTripService(
	repo repo.TripRepository,
) *TripService {
	return &TripService{
		CommonService[Trip]{
			repo: repo,
		},
	}
}
