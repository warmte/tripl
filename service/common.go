package service

import (
	"context"
	"fmt"
	"time"

	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/internal/repo"
)

type Service[T TripObjectView] interface {
	Create(ctx context.Context, data TripObjectView) (*T, error)
	GetPage(ctx context.Context, userID int64, listType ListType, page int) (Page[T], error)
	GetActualPage(ctx context.Context, userID int64, page int) (Page[T], error)
	GetArchivePage(ctx context.Context, userID int64, page int) (Page[T], error)
	GetByID(ctx context.Context, userID int64, objectID int64) (*T, error)
	UpdateField(ctx context.Context, id int64, userID int64, field string, value any) error
	Delete(ctx context.Context, id int64, userID int64) error
	AttachToTrip(ctx context.Context, objID int64, tripID int64, userID int64) error
	DetachFromTrip(ctx context.Context, objID int64, userID int64) error
	GetPageByTripID(ctx context.Context, tripID int64, userID int64, page int) (Page[T], error)
	CountByTripID(ctx context.Context, tripID int64, userID int64) (int64, error)
	GetCostsByTripID(ctx context.Context, tripID int64, userID int64) (map[string]float64, error)
	GetPageForTime(ctx context.Context, userID int64, start time.Time, end time.Time, page int) (Page[T], error)
	CountByUserID(ctx context.Context, userID int64) (int64, error)
}

type CommonService[T TripObjectView] struct {
	repo repo.Repository[T]
}

func NewCommonService[T TripObjectView](repo repo.Repository[T]) *CommonService[T] {
	return &CommonService[T]{repo: repo}
}

func PreparePage[T TripObjectView](items []T, pageSize int) Page[T] {
	hasNext := len(items) > pageSize

	if hasNext {
		items = items[:pageSize]
	}

	return Page[T]{
		Items:   items,
		HasNext: hasNext,
	}
}

func (s CommonService[T]) GetActualPage(
	ctx context.Context,
	userID int64,
	page int,
) (Page[T], error) {
	objects, err := s.repo.GetActualPage(ctx, userID, page)
	if err != nil {
		return Page[T]{},
			fmt.Errorf("get actual page: %w", err)
	}

	return PreparePage(objects, PageSize), nil
}

func (s CommonService[T]) GetArchivePage(
	ctx context.Context,
	userID int64,
	page int,
) (Page[T], error) {
	objects, err := s.repo.GetArchivePage(ctx, userID, page)
	if err != nil {
		return Page[T]{},
			fmt.Errorf("get archive page: %w", err)
	}

	return PreparePage(objects, PageSize), nil
}

func (s CommonService[T]) Create(ctx context.Context, data TripObjectView) (*T, error) {
	value, ok := data.(T)
	if !ok {
		return nil, ErrInternalError
	}
	return s.repo.Create(ctx, &value)
}

func (s CommonService[T]) GetPage(
	ctx context.Context,
	userID int64,
	listType ListType,
	page int,
) (Page[T], error) {

	switch listType {
	case ActualListType:
		return s.GetActualPage(ctx, userID, page)
	case ArchiveListType:
		return s.GetArchivePage(ctx, userID, page)
	default:
		return Page[T]{}, fmt.Errorf("unknown list type: %s", listType)
	}
}

func (s CommonService[T]) GetPageByTripID(
	ctx context.Context,
	tripID int64,
	userID int64,
	page int,
) (Page[T], error) {
	tickets, err := s.repo.GetPageByTripID(ctx, tripID, userID, page)
	if err != nil {
		return Page[T]{},
			fmt.Errorf("get page by trip id: %w", err)
	}

	return PreparePage(tickets, PageSize), nil
}

func (s CommonService[T]) GetByID(
	ctx context.Context,
	userID int64,
	objectID int64,
) (*T, error) {
	return s.repo.GetByID(ctx, objectID, userID)
}

func (s CommonService[T]) UpdateField(ctx context.Context, id int64, userID int64, field string, value any) error {
	return s.repo.UpdateField(ctx, id, userID, field, value)
}

func (s CommonService[T]) Delete(
	ctx context.Context,
	id int64,
	userID int64,
) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s CommonService[T]) AttachToTrip(
	ctx context.Context,
	objID int64,
	tripID int64,
	userID int64,
) error {
	if objID <= 0 {
		return fmt.Errorf("invalid object id: %d", objID)
	}

	if tripID <= 0 {
		return fmt.Errorf("invalid trip id: %d", tripID)
	}

	if userID <= 0 {
		return fmt.Errorf("invalid user id: %d", userID)
	}

	if err := s.repo.AttachToTrip(
		ctx,
		objID,
		tripID,
		userID,
	); err != nil {
		return fmt.Errorf("attach to trip: %w", err)
	}

	return nil
}

func (s CommonService[T]) DetachFromTrip(
	ctx context.Context,
	objID int64,
	userID int64,
) error {
	if objID <= 0 {
		return fmt.Errorf("invalid object id: %d", objID)
	}

	if userID <= 0 {
		return fmt.Errorf("invalid user id: %d", userID)
	}

	if err := s.repo.DetachFromTrip(
		ctx,
		objID,
		userID,
	); err != nil {
		return fmt.Errorf("detach from trip: %w", err)
	}

	return nil
}

func (s CommonService[T]) GetCostsByTripID(ctx context.Context, tripID int64, userID int64) (map[string]float64, error) {
	return s.repo.GetCostsByTripID(ctx, tripID, userID)
}

func (s CommonService[T]) CountByTripID(ctx context.Context, tripID int64, userID int64) (int64, error) {
	return s.repo.CountByTripID(ctx, tripID, userID)
}

func (s CommonService[T]) GetPageForTime(
	ctx context.Context,
	userID int64,
	start time.Time,
	end time.Time,
	page int,
) (Page[T], error) {

	objects, err := s.repo.GetForTime(ctx, userID, start, end, page)
	if err != nil {
		return Page[T]{},
			fmt.Errorf("get actual page: %w", err)
	}

	return PreparePage(objects, 1), nil
}

func (s CommonService[T]) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	return s.repo.CountByUserID(ctx, userID)
}
