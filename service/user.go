package service

import (
	"context"
	"fmt"
	"log"
	"time"

	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/internal/repo"
	"github.com/warmte/tripl/lang"
)

const (
	locationName = "Europe/Moscow"
	hour         = 9
	minute       = 0
)

type State struct {
	Stage   string
	Context TripObjectView
}

func (s State) For(label string) State {
	return State{fmt.Sprintf("%s_%s", s.Stage, label), s.Context}
}

func (s State) With(context TripObjectView) State {
	return State{s.Stage, context}
}

type EmptyContext struct{}

func (ec EmptyContext) Show(cfg *lang.TripObjectCfg) string {
	return ""
}
func (ec EmptyContext) Preview(cfg *lang.TripObjectCfg) string {
	return ""
}
func (ec EmptyContext) GetTripID() *int64 {
	return nil
}
func (ec EmptyContext) GetID() int64 {
	return 0
}

var (
	Default State = State{"", EmptyContext{}}
	AddInfo State = State{"add_info", EmptyContext{}}
	Added   State = State{"added", EmptyContext{}}
)

type UserService struct {
	state    map[int64]State
	repo     repo.UserRepository
	location *time.Location
}

func NewUserService(repo repo.UserRepository) *UserService {
	location, err := time.LoadLocation(locationName)
	if err != nil {
		return nil
	}
	return &UserService{repo: repo, state: map[int64]State{}, location: location}
}

func (s UserService) GetState(userID int64) State {
	v, ok := s.state[userID]
	if ok {
		return v
	} else {
		return Default
	}
}

func (s UserService) SetState(telegramID int64, newState State) {
	if newState == Default {
		delete(s.state, telegramID)
	} else {
		s.state[telegramID] = newState
	}
}

func (s UserService) SetOrUpdateLang(ctx context.Context, telegramID int64, lang lang.Language) (*User, error) {
	s.SetState(telegramID, Default)
	user, err := s.repo.UpdateLangOrCreate(ctx, telegramID, lang)
	return user, err
}

func (s UserService) GetAll(ctx context.Context) ([]User, error) {
	return s.repo.GetAll(ctx)
}

func (s UserService) GetLanguageCfg(ctx context.Context, userID int64) *lang.Cfg {
	language, err := s.repo.GetLanguage(ctx, userID)
	if err == nil && lang.Language(language) == lang.LanguageEn {
		return lang.Get(lang.LanguageEn)
	} else {
		return lang.Get(lang.LanguageRu)
	}
}

func (s UserService) GetLanguage(ctx context.Context, userID int64) lang.Language {
	language, err := s.repo.GetLanguage(ctx, userID)
	if err == nil && lang.Language(language) == lang.LanguageEn {
		return lang.LanguageEn
	} else {
		return lang.LanguageRu
	}
}

func (s UserService) TryConsume(ctx context.Context, userID int64) error {

	now := time.Now().In(s.location)

	ok, err := s.repo.TryConsume(ctx, userID, MaxUploadsPerDay, now)
	if err != nil {
		return fmt.Errorf("check daily upload limit: %w", err)
	}
	if !ok {
		return ErrDailyUploadLimitReached
	}

	return nil
}

func (s UserService) GetRemaining(ctx context.Context, userID int64) (int, error) {
	now := time.Now().In(s.location)

	count, err := s.repo.GetCount(ctx, userID, now)
	if err != nil {
		return 0, fmt.Errorf("get daily upload usage: %w", err)
	}

	remaining := MaxUploadsPerDay - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

func (s UserService) RunDailyUsageCleanup(ctx context.Context) {
	for {
		now := time.Now().In(s.location)
		nextRun := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, s.location)

		timer := time.NewTimer(time.Until(nextRun))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}

		cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := s.repo.DeleteOlderThan(cleanupCtx, time.Now().In(s.location))
		cancel()

		if err != nil {
			log.Printf("daily usage cleanup failed: %v", err)
		}
	}
}
