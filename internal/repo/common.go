package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/warmte/tripl/common"
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
	GetPageByTripID(ctx context.Context, tripID int64, userID int64, page int) ([]T, error)
	GetAllByTripID(ctx context.Context, tripID int64, userID int64) ([]T, error)
	CountByTripID(ctx context.Context, tripID int64, userID int64) (int64, error)
	GetCostsByTripID(ctx context.Context, tripID int64, userID int64) (map[string]float64, error)
	GetForTime(ctx context.Context, userID int64, start time.Time, end time.Time, page int) ([]T, error)
	CountByUserID(ctx context.Context, userID int64) (int64, error)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23514" {
			return ErrDatetimeRestraint
		}

		return fmt.Errorf("update field %s: %w", field, err)
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

func (r BasicRepository) CountByTripID(
	ctx context.Context,
	tripID int64,
	userID int64,
) (int64, error) {
	var count int64

	err := r.db.QueryRow(
		ctx,
		fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE trip_id = $1
		  AND user_id = $2
		`, r.repoType),
		tripID,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count by trip id: %w", err)
	}

	return count, nil
}

func (r BasicRepository) GetCostsByTripID(
	ctx context.Context,
	tripID int64,
	userID int64,
) (map[string]float64, error) {
	rows, err := r.db.Query(
		ctx,
		fmt.Sprintf(`
 		SELECT COALESCE(currency, '') AS currency, SUM(cost::numeric)
    FROM %s
    WHERE trip_id = $1 
		  AND user_id = $2
      AND cost ~ '^[[:space:]]*[+-]?([0-9]+([.][0-9]+)?|[.][0-9]+)[[:space:]]*$'
    GROUP BY COALESCE(currency, '') 
		`, r.repoType),
		tripID,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get costs by trip id: %w", err)
	}
	defer rows.Close()

	result := make(map[string]float64)

	for rows.Next() {
		var (
			currency string
			sum      float64
		)

		if err := rows.Scan(&currency, &sum); err != nil {
			return nil, fmt.Errorf("scan costs: %w", err)
		}

		result[currency] = sum
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate costs: %w", err)
	}

	return result, nil
}

func (r BasicRepository) CountByUserID(
	ctx context.Context,
	userID int64,
) (int64, error) {
	var count int64

	err := r.db.QueryRow(
		ctx,
		fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE user_id = $1
		`, r.repoType),
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count by user id: %w", err)
	}

	return count, nil
}
