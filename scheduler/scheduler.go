package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	. "github.com/warmte/tripl/common"
)

const (
	hour   = 15
	minute = 41
)

type Scheduler struct {
	location *time.Location
	sender   NotificationSender
}

func NewScheduler(sender NotificationSender) (*Scheduler, error) {
	location, err := time.LoadLocation(LocationName)
	if err != nil {
		return nil, fmt.Errorf("load scheduler location: %w", err)
	}

	return &Scheduler{
		location: location,
		sender:   sender,
	}, nil
}

func (s *Scheduler) Run(ctx context.Context) {
	for {
		log.Printf("scheduler: ready")
		now := time.Now().In(s.location)
		next := nextRun(now)

		timer := time.NewTimer(time.Until(next))

		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return

		case <-timer.C:
			date := next.In(s.location).AddDate(0, 0, 1)

			if err := s.sender.SendTomorrow(ctx, date, s.location); err != nil {
				log.Printf("scheduler: send tomorrow notifications: %v", err)
			}
		}
	}
}

func nextRun(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}

	return next
}
