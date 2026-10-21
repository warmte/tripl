package service

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	. "github.com/warmte/tripl/common"
)

type MockService[T TripObjectView] struct {
	mock.Mock
}

func (m *MockService[T]) Create(ctx context.Context, data TripObjectView) (*T, error) {
	args := m.Called(ctx, data)

	var result *T
	if v := args.Get(0); v != nil {
		result = v.(*T)
	}
	return result, args.Error(1)
}

func (m *MockService[T]) GetPage(ctx context.Context, userID int64, listType ListType, page int) (Page[T], error) {
	args := m.Called(ctx, userID, listType, page)
	return args.Get(0).(Page[T]), args.Error(1)
}

func (m *MockService[T]) GetActualPage(ctx context.Context, userID int64, page int) (Page[T], error) {
	args := m.Called(ctx, userID, page)
	return args.Get(0).(Page[T]), args.Error(1)
}

func (m *MockService[T]) GetArchivePage(ctx context.Context, userID int64, page int) (Page[T], error) {
	args := m.Called(ctx, userID, page)
	return args.Get(0).(Page[T]), args.Error(1)
}

func (m *MockService[T]) GetByID(ctx context.Context, userID, objectID int64) (*T, error) {
	args := m.Called(ctx, userID, objectID)

	var result *T
	if v := args.Get(0); v != nil {
		result = v.(*T)
	}
	return result, args.Error(1)
}

func (m *MockService[T]) UpdateField(ctx context.Context, id, userID int64, field string, value any) error {
	return m.Called(ctx, id, userID, field, value).Error(0)
}

func (m *MockService[T]) Delete(ctx context.Context, id, userID int64) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *MockService[T]) AttachToTrip(ctx context.Context, objID, tripID, userID int64) error {
	return m.Called(ctx, objID, tripID, userID).Error(0)
}

func (m *MockService[T]) DetachFromTrip(ctx context.Context, objID, userID int64) error {
	return m.Called(ctx, objID, userID).Error(0)
}

func (m *MockService[T]) GetPageByTripID(ctx context.Context, tripID, userID int64, page int) (Page[T], error) {
	args := m.Called(ctx, tripID, userID, page)
	return args.Get(0).(Page[T]), args.Error(1)
}

func (m *MockService[T]) GetAllByTripID(ctx context.Context, tripID, userID int64) ([]T, error) {
	args := m.Called(ctx, tripID, userID)
	return args.Get(0).([]T), args.Error(1)
}

func (m *MockService[T]) CountByTripID(ctx context.Context, tripID, userID int64) (int64, error) {
	args := m.Called(ctx, tripID, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockService[T]) GetCostsByTripID(ctx context.Context, tripID, userID int64) (map[string]float64, error) {
	args := m.Called(ctx, tripID, userID)

	var result map[string]float64
	if v := args.Get(0); v != nil {
		result = v.(map[string]float64)
	}
	return result, args.Error(1)
}

func (m *MockService[T]) GetPageForTime(ctx context.Context, userID int64, start, end time.Time, page int) (Page[T], error) {
	args := m.Called(ctx, userID, start, end, page)
	return args.Get(0).(Page[T]), args.Error(1)
}

func (m *MockService[T]) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
