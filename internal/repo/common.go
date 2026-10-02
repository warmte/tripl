package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	TicketsRepo  = "tickets"
	BookingsRepo = "bookings"
	EventsRepo   = "events"
	TripsRepo    = "trips"
)

type Repository[T any] interface {
	Create(ctx context.Context, t *T) (*T, error)
	Delete(ctx context.Context, id int64, userID int64) error
	GetByID(ctx context.Context, id int64, userID int64) (*T, error)
	UpdateField(ctx context.Context, id int64, userID int64, field string, value any) error
	GetActual(ctx context.Context, userID int64) ([]T, error)
	GetArchive(ctx context.Context, userID int64) ([]T, error)
	GetActualPage(ctx context.Context, userID int64, page int) ([]T, error)
	GetArchivePage(ctx context.Context, userID int64, page int) ([]T, error)
	AttachToTrip(ctx context.Context, objID int64, tripID int64, userID int64) error
	DetachFromTrip(ctx context.Context, objID int64, userID int64) error
}

type BasicRepository struct {
	db          *pgxpool.Pool
	repoType    string
	errNotFound error
}

func MakeBasic(db *pgxpool.Pool, repoType string, errNotFound error) BasicRepository {
	return BasicRepository{db, repoType, errNotFound}
}

func (r BasicRepository) Delete(ctx context.Context, id int64, userID int64) error {
	tag, err := r.db.Exec(
		ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE id = $1 AND user_id = $2`, r.repoType),
		id,
		userID,
	)
	if err != nil {
		return fmt.Errorf("can't delete %s: %w", r.repoType, err)
	}

	if tag.RowsAffected() == 0 {
		return r.errNotFound
	}

	return nil
}

func (r BasicRepository) UpdateField(ctx context.Context, id int64, userID int64, field string, value any) error {
	query := fmt.Sprintf(
		`UPDATE %s SET %s = $1 WHERE id = $2 and user_id = $3`,
		r.repoType, field,
	)

	tag, err := r.db.Exec(ctx, query, value, id, userID)
	if err != nil {
		return fmt.Errorf("can't update field %s: %w", field, err)
	}

	if tag.RowsAffected() == 0 {
		return r.errNotFound
	}

	return nil
}

func (r BasicRepository) AttachToTrip(ctx context.Context, objID int64, tripID int64, userID int64) error {
	tag, err := r.db.Exec(
		ctx,
		fmt.Sprintf(`
			UPDATE %s AS obj
			SET trip_id = $1
			WHERE obj.id = $2
			  AND obj.user_id = $3
			  AND EXISTS (
				  SELECT 1
				  FROM trips
				  WHERE trips.id = $1
				    AND trips.user_id = $3
			  )
		`, r.repoType),
		tripID,
		objID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("can't attach %s to trip: %w", r.repoType, err)
	}

	if tag.RowsAffected() == 0 {
		return r.errNotFound
	}

	return nil
}

func (r BasicRepository) DetachFromTrip(ctx context.Context, objID int64, userID int64) error {
	tag, err := r.db.Exec(
		ctx,
		fmt.Sprintf(`
			UPDATE %s
			SET trip_id = NULL
			WHERE id = $1
			  AND user_id = $2
		`, r.repoType),
		objID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("can't detach %s from trip: %w", r.repoType, err)
	}

	if tag.RowsAffected() == 0 {
		return r.errNotFound
	}

	return nil
}
