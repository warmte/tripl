package service

import (
	"context"
	"fmt"

	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/internal/repo"
	"github.com/warmte/tripl/lang"
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
	state map[int64]State
	repo  repo.UserRepository
}

func NewUserService(repo repo.UserRepository) *UserService {
	return &UserService{repo: repo, state: map[int64]State{}}
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

func (s UserService) GetLanguage(ctx context.Context, userID int64) (string, error) {
	return s.repo.GetLanguage(ctx, userID)
}
