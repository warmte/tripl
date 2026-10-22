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

var ErrEventNotFound = errors.New("event not found")

type EventRepository struct {
	BasicRepository
}

var _ Repository[Event] = EventRepository{} // compile-time check

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{
		MakeBasic(db, EventsRepo, ErrEventNotFound),
	}
}

func (r EventRepository) Create(
	ctx context.Context,
	event *Event,
) (*Event, error) {
	var eventRepr Event

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO events (
			user_id,
			name,
			location,
			start_time,
			end_time,
			cost,
			currency,
			comment
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING
			id,
			user_id,
			trip_id,
			name,
			location,
			start_time,
			end_time,
			cost,
			currency,
			comment`,
		event.UserID,
		event.Name,
		event.Location,
		event.StartTime,
		event.EndTime,
		event.Cost,
		event.Currency,
		event.Comment,
	).Scan(
		&eventRepr.ID,
		&eventRepr.UserID,
		&eventRepr.TripID,
		&eventRepr.Name,
		&eventRepr.Location,
		&eventRepr.StartTime,
		&eventRepr.EndTime,
		&eventRepr.Cost,
		&eventRepr.Currency,
		&eventRepr.Comment,
	)

	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}

	return &eventRepr, nil
}

func (r EventRepository) GetByID(
	ctx context.Context,
	id int64,
	userID int64,
) (*Event, error) {
	var eventRepr Event

	err := r.db.QueryRow(
		ctx,
		`SELECT
			id,
			user_id,
			trip_id,
			name,
			location,
			start_time,
			end_time,
			cost,
			currency,
			comment
		FROM events
		WHERE id = $1
		  AND user_id = $2`,
		id,
		userID,
	).Scan(
		&eventRepr.ID,
		&eventRepr.UserID,
		&eventRepr.TripID,
		&eventRepr.Name,
		&eventRepr.Location,
		&eventRepr.StartTime,
		&eventRepr.EndTime,
		&eventRepr.Cost,
		&eventRepr.Currency,
		&eventRepr.Comment,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEventNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	return &eventRepr, nil
}

func (r *EventRepository) UpdateName(
	ctx context.Context,
	id int64,
	userID int64,
	name string,
) error {
	return r.UpdateField(ctx, id, userID, "name", name)
}

func (r *EventRepository) UpdateLocation(
	ctx context.Context,
	id int64,
	userID int64,
	location string,
) error {
	return r.UpdateField(ctx, id, userID, "location", location)
}

func (r *EventRepository) UpdateStartTime(
	ctx context.Context,
	id int64,
	userID int64,
	startTime time.Time,
) error {
	return r.UpdateField(ctx, id, userID, "start_time", startTime)
}

func (r *EventRepository) UpdateEndTime(
	ctx context.Context,
	id int64,
	userID int64,
	endTime *time.Time,
) error {
	return r.UpdateField(ctx, id, userID, "end_time", endTime)
}

func (r *EventRepository) UpdateCost(
	ctx context.Context,
	id int64,
	userID int64,
	cost *string,
) error {
	return r.UpdateField(ctx, id, userID, "cost", cost)
}

func (r *EventRepository) UpdateCurrency(
	ctx context.Context,
	id int64,
	userID int64,
	currency *string,
) error {
	return r.UpdateField(ctx, id, userID, "currency", currency)
}

func (r *EventRepository) UpdateComment(
	ctx context.Context,
	id int64,
	userID int64,
	comment *string,
) error {
	return r.UpdateField(ctx, id, userID, "comment", comment)
}

func (r EventRepository) GetActualPage(
	ctx context.Context,
	userID int64,
	page int,
) ([]Event, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			name,
			location,
			start_time,
			end_time,
			cost,
			currency,
			comment
		FROM events
		WHERE user_id = $1
		  AND (COALESCE(end_time, start_time) >= NOW() OR COALESCE(end_time, start_time) IS NULL)
		ORDER BY start_time, id
		LIMIT $2 OFFSET $3
	`, userID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get actual events page: %w", err)
	}
	defer rows.Close()

	var result []Event

	for rows.Next() {
		var event Event

		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.TripID,
			&event.Name,
			&event.Location,
			&event.StartTime,
			&event.EndTime,
			&event.Cost,
			&event.Currency,
			&event.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		result = append(result, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return result, nil
}

func (r EventRepository) GetArchivePage(
	ctx context.Context,
	userID int64,
	page int,
) ([]Event, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			name,
			location,
			start_time,
			end_time,
			cost,
			currency,
			comment
		FROM events
		WHERE user_id = $1
		  AND COALESCE(end_time, start_time) < NOW()
		ORDER BY start_time DESC, id DESC
		LIMIT $2 OFFSET $3
	`, userID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get archive events page: %w", err)
	}
	defer rows.Close()

	var result []Event

	for rows.Next() {
		var event Event

		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.TripID,
			&event.Name,
			&event.Location,
			&event.StartTime,
			&event.EndTime,
			&event.Cost,
			&event.Currency,
			&event.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		result = append(result, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return result, nil
}

func (r EventRepository) GetPageByTripID(ctx context.Context, tripID int64, userID int64, page int) ([]Event, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			name,
			location,
			start_time,
			end_time,
			cost,
			currency,
			comment
		FROM events
		WHERE user_id = $1
		  AND trip_id = $2
		ORDER BY start_time ASC, id ASC
		LIMIT $3 OFFSET $4
	`, userID, tripID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get events page by trip id: %w", err)
	}
	defer rows.Close()

	var result []Event

	for rows.Next() {
		var event Event

		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.TripID,
			&event.Name,
			&event.Location,
			&event.StartTime,
			&event.EndTime,
			&event.Cost,
			&event.Currency,
			&event.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		result = append(result, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return result, nil
}

func (r EventRepository) GetAllByTripID(ctx context.Context, tripID int64, userID int64) ([]Event, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			name,
			location,
			start_time,
			end_time,
			cost,
			currency,
			comment
		FROM events
		WHERE user_id = $1
		  AND trip_id = $2
		ORDER BY start_time ASC, id ASC
	`, userID, tripID)
	if err != nil {
		return nil, fmt.Errorf("get events page by trip id: %w", err)
	}
	defer rows.Close()

	var result []Event

	for rows.Next() {
		var event Event

		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.TripID,
			&event.Name,
			&event.Location,
			&event.StartTime,
			&event.EndTime,
			&event.Cost,
			&event.Currency,
			&event.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		result = append(result, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return result, nil
}

func (r EventRepository) GetForTime(
	ctx context.Context,
	userID int64,
	start time.Time,
	end time.Time,
	page int,
) ([]Event, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			name,
			location,
			start_time,
			end_time,
			cost,
			comment
		FROM events
		WHERE user_id = $1
		  AND start_time >= $2
		  AND start_time < $3
		ORDER BY start_time ASC, id ASC
		LIMIT $4 OFFSET $5
	`, userID, start, end, 2, page)
	if err != nil {
		return nil, fmt.Errorf("get events for time: %w", err)
	}
	defer rows.Close()

	result := make([]Event, 0, 2)

	for rows.Next() {
		var event Event

		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.TripID,
			&event.Name,
			&event.Location,
			&event.StartTime,
			&event.EndTime,
			&event.Cost,
			&event.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		result = append(result, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return result, nil
}
