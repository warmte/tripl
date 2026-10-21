package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/warmte/tripl/common"
)

var ErrTripNotFound = errors.New("trip not found")

type TripRepository struct {
	BasicRepository
}

var _ Repository[Trip] = TripRepository{} // compile-time check

func NewTripRepository(db *pgxpool.Pool) *TripRepository {
	return &TripRepository{
		MakeBasic(db, TripsRepo, ErrTripNotFound),
	}
}

func (r TripRepository) Create(
	ctx context.Context,
	trip *Trip,
) (*Trip, error) {
	var tripRepr Trip

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO trips (
			user_id,
			name,
			start_date,
			end_date,
			comment
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			user_id,
			name,
			start_date,
			end_date,
			comment`,
		trip.UserID,
		trip.Name,
		trip.StartDate,
		trip.EndDate,
		trip.Comment,
	).Scan(
		&tripRepr.ID,
		&tripRepr.UserID,
		&tripRepr.Name,
		&tripRepr.StartDate,
		&tripRepr.EndDate,
		&tripRepr.Comment,
	)

	if err != nil {
		return nil, fmt.Errorf("create trip: %w", err)
	}

	return &tripRepr, nil
}

func (r TripRepository) GetByID(
	ctx context.Context,
	id int64,
	userID int64,
) (*Trip, error) {
	var tripRepr Trip

	err := r.db.QueryRow(
		ctx,
		`SELECT
			id,
			user_id,
			name,
			start_date,
			end_date,
			comment
		FROM trips
		WHERE id = $1
		  AND user_id = $2`,
		id,
		userID,
	).Scan(
		&tripRepr.ID,
		&tripRepr.UserID,
		&tripRepr.Name,
		&tripRepr.StartDate,
		&tripRepr.EndDate,
		&tripRepr.Comment,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTripNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get trip: %w", err)
	}

	return &tripRepr, nil
}

func (r *TripRepository) UpdateName(
	ctx context.Context,
	id int64,
	userID int64,
	name string,
) error {
	return r.UpdateField(ctx, id, userID, "name", name)
}

func (r *TripRepository) UpdateStartDate(
	ctx context.Context,
	id int64,
	userID int64,
	startDate time.Time,
) error {
	return r.UpdateField(ctx, id, userID, "start_date", startDate)
}

func (r *TripRepository) UpdateEndDate(
	ctx context.Context,
	id int64,
	userID int64,
	endDate time.Time,
) error {
	return r.UpdateField(ctx, id, userID, "end_date", endDate)
}

func (r *TripRepository) UpdateComment(
	ctx context.Context,
	id int64,
	userID int64,
	comment *string,
) error {
	return r.UpdateField(ctx, id, userID, "comment", comment)
}

func (r TripRepository) GetActual(
	ctx context.Context,
	userID int64,
) ([]Trip, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
			id,
			user_id,
			name,
			start_date,
			end_date,
			comment
		FROM trips
		WHERE user_id = $1
		  AND end_date >= CURRENT_DATE
		ORDER BY start_date ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get actual trips: %w", err)
	}
	defer rows.Close()

	trips := make([]Trip, 0)

	for rows.Next() {
		var trip Trip

		if err := rows.Scan(
			&trip.ID,
			&trip.UserID,
			&trip.Name,
			&trip.StartDate,
			&trip.EndDate,
			&trip.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan actual trip: %w", err)
		}

		trips = append(trips, trip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate actual trips: %w", err)
	}

	return trips, nil
}

func (r TripRepository) GetArchive(
	ctx context.Context,
	userID int64,
) ([]Trip, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
			id,
			user_id,
			name,
			start_date,
			end_date,
			comment
		FROM trips
		WHERE user_id = $1
		  AND end_date < CURRENT_DATE
		ORDER BY start_date DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get archive trips: %w", err)
	}
	defer rows.Close()

	trips := make([]Trip, 0)

	for rows.Next() {
		var trip Trip

		if err := rows.Scan(
			&trip.ID,
			&trip.UserID,
			&trip.Name,
			&trip.StartDate,
			&trip.EndDate,
			&trip.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan archive trip: %w", err)
		}

		trips = append(trips, trip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive trips: %w", err)
	}

	return trips, nil
}

func (r TripRepository) GetActualPage(
	ctx context.Context,
	userID int64,
	page int,
) ([]Trip, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			name,
			start_date,
			end_date,
			comment
		FROM trips
		WHERE user_id = $1
		  AND end_date >= CURRENT_DATE
		ORDER BY start_date, id
		LIMIT $2 OFFSET $3
	`, userID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get actual trips page: %w", err)
	}
	defer rows.Close()

	var result []Trip

	for rows.Next() {
		var trip Trip

		if err := rows.Scan(
			&trip.ID,
			&trip.UserID,
			&trip.Name,
			&trip.StartDate,
			&trip.EndDate,
			&trip.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan trip: %w", err)
		}

		result = append(result, trip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trips: %w", err)
	}

	return result, nil
}

func (r TripRepository) GetArchivePage(
	ctx context.Context,
	userID int64,
	page int,
) ([]Trip, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			name,
			start_date,
			end_date,
			comment
		FROM trips
		WHERE user_id = $1
		  AND end_date < CURRENT_DATE
		ORDER BY end_date DESC, id DESC
		LIMIT $2 OFFSET $3
	`, userID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get archive trips page: %w", err)
	}
	defer rows.Close()

	var result []Trip

	for rows.Next() {
		var trip Trip

		if err := rows.Scan(
			&trip.ID,
			&trip.UserID,
			&trip.Name,
			&trip.StartDate,
			&trip.EndDate,
			&trip.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan trip: %w", err)
		}

		result = append(result, trip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trips: %w", err)
	}

	return result, nil
}

func (r TripRepository) GetPageByTripID(ctx context.Context, tripID int64, userID int64, page int) ([]Trip, error) {
	return nil, fmt.Errorf("unsupported: can't get trips by trip id")
}

func (r TripRepository) GetAllByTripID(ctx context.Context, tripID int64, userID int64) ([]Trip, error) {
	return nil, fmt.Errorf("unsupported: can't get trips by trip id")
}

func (r TripRepository) GetForTime(ctx context.Context, userID int64, start time.Time, end time.Time, page int) ([]Trip, error) {
	return nil, fmt.Errorf("unsupported: don't need trips for notifies")
}
