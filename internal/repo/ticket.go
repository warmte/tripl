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

var ErrTicketNotFound = errors.New("ticket not found")

type TicketRepository struct {
	BasicRepository
}

var _ Repository[Ticket] = TicketRepository{} // compile-time check

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{
		MakeBasic(db, TicketsRepo, ErrTicketNotFound),
	}
}

func (r TicketRepository) Create(ctx context.Context, ticket *Ticket) (*Ticket, error) {
	var ticketResult Ticket

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO tickets (
			user_id,
			vehicle,
			departure_point,
			arrival_point,
			departure_time,
			arrival_time,
			cost,
			currency,
			departure_address,
			arrival_address,
			seat,
			number,
			comment
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING
			id,
			user_id,
			trip_id,
			vehicle,
			departure_point,
			arrival_point,
			departure_time,
			arrival_time,
			cost,
			currency,
			departure_address,
			arrival_address,
			seat,
			number,
			comment`,
		ticket.UserID,
		ticket.Vehicle,
		ticket.DeparturePoint,
		ticket.ArrivalPoint,
		ticket.DepartureTime,
		ticket.ArrivalTime,
		ticket.Cost,
		ticket.Currency,
		ticket.DepartureAddress,
		ticket.ArrivalAddress,
		ticket.Seat,
		ticket.Number,
		ticket.Comment,
	).Scan(
		&ticketResult.ID,
		&ticketResult.UserID,
		&ticketResult.TripID,
		&ticketResult.Vehicle,
		&ticketResult.DeparturePoint,
		&ticketResult.ArrivalPoint,
		&ticketResult.DepartureTime,
		&ticketResult.ArrivalTime,
		&ticketResult.Cost,
		&ticketResult.Currency,
		&ticketResult.DepartureAddress,
		&ticketResult.ArrivalAddress,
		&ticketResult.Seat,
		&ticketResult.Number,
		&ticketResult.Comment,
	)

	if err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}

	return &ticketResult, nil
}

func (r TicketRepository) GetByID(
	ctx context.Context,
	id int64,
	userID int64,
) (*Ticket, error) {
	var ticket Ticket

	err := r.db.QueryRow(
		ctx,
		`SELECT
			id,
			user_id,
			trip_id,
			vehicle,
			departure_point,
			arrival_point,
			departure_time,
			arrival_time,
			cost,
			currency,
			departure_address,
			arrival_address,
			seat,
			number,
			comment
		FROM tickets
		WHERE id = $1 AND user_id = $2`,
		id,
		userID,
	).Scan(
		&ticket.ID,
		&ticket.UserID,
		&ticket.TripID,
		&ticket.Vehicle,
		&ticket.DeparturePoint,
		&ticket.ArrivalPoint,
		&ticket.DepartureTime,
		&ticket.ArrivalTime,
		&ticket.Cost,
		&ticket.Currency,
		&ticket.DepartureAddress,
		&ticket.ArrivalAddress,
		&ticket.Seat,
		&ticket.Number,
		&ticket.Comment,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTicketNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get ticket: %w", err)
	}

	return &ticket, nil
}

func (r *TicketRepository) UpdateVehicle(
	ctx context.Context,
	id int64,
	userID int64,
	vehicle string,
) error {
	return r.UpdateField(ctx, id, userID, "vehicle", vehicle)
}

func (r *TicketRepository) UpdateDeparturePoint(
	ctx context.Context,
	id int64,
	userID int64,
	departurePoint string,
) error {
	return r.UpdateField(ctx, id, userID, "departure_point", departurePoint)
}

func (r *TicketRepository) UpdateArrivalPoint(
	ctx context.Context,
	id int64,
	userID int64,
	arrivalPoint string,
) error {
	return r.UpdateField(ctx, id, userID, "arrival_point", arrivalPoint)
}

func (r *TicketRepository) UpdateDepartureTime(
	ctx context.Context,
	id int64,
	userID int64,
	departureTime time.Time,
) error {
	return r.UpdateField(ctx, id, userID, "departure_time", departureTime)
}

func (r *TicketRepository) UpdateArrivalTime(
	ctx context.Context,
	id int64,
	userID int64,
	arrivalTime time.Time,
) error {
	return r.UpdateField(ctx, id, userID, "arrival_time", arrivalTime)
}

func (r *TicketRepository) UpdateCost(
	ctx context.Context,
	id int64,
	userID int64,
	cost *string,
) error {
	return r.UpdateField(ctx, id, userID, "cost", cost)
}

func (r *TicketRepository) UpdateCurrency(
	ctx context.Context,
	id int64,
	userID int64,
	currency *string,
) error {
	return r.UpdateField(ctx, id, userID, "currency", currency)
}

func (r *TicketRepository) UpdateDepartureAddress(
	ctx context.Context,
	id int64,
	userID int64,
	address *string,
) error {
	return r.UpdateField(ctx, id, userID, "departure_address", address)
}

func (r *TicketRepository) UpdateArrivalAddress(
	ctx context.Context,
	id int64,
	userID int64,
	address *string,
) error {
	return r.UpdateField(ctx, id, userID, "arrival_address", address)
}

func (r *TicketRepository) UpdateSeat(
	ctx context.Context,
	id int64,
	userID int64,
	seat *string,
) error {
	return r.UpdateField(ctx, id, userID, "seat", seat)
}

func (r *TicketRepository) UpdateNumber(
	ctx context.Context,
	id int64,
	userID int64,
	number *string,
) error {
	return r.UpdateField(ctx, id, userID, "number", number)
}

func (r *TicketRepository) UpdateComment(
	ctx context.Context,
	id int64,
	userID int64,
	comment *string,
) error {
	return r.UpdateField(ctx, id, userID, "comment", comment)
}

func (r TicketRepository) GetActual(
	ctx context.Context,
	userID int64,
) ([]Ticket, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
			id,
			user_id,
			trip_id,
			vehicle,
			departure_point,
			arrival_point,
			departure_time,
			arrival_time,
			cost,
			currency,
			departure_address,
			arrival_address,
			seat,
			number,
			comment
		FROM tickets
		WHERE user_id = $1
		  AND arrival_time >= NOW()
		ORDER BY arrival_time ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get actual tickets: %w", err)
	}
	defer rows.Close()

	tickets := make([]Ticket, 0)

	for rows.Next() {
		var ticket Ticket

		if err := rows.Scan(
			&ticket.ID,
			&ticket.UserID,
			&ticket.TripID,
			&ticket.Vehicle,
			&ticket.DeparturePoint,
			&ticket.ArrivalPoint,
			&ticket.DepartureTime,
			&ticket.ArrivalTime,
			&ticket.Cost,
			&ticket.Currency,
			&ticket.DepartureAddress,
			&ticket.ArrivalAddress,
			&ticket.Seat,
			&ticket.Number,
			&ticket.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan actual ticket: %w", err)
		}

		tickets = append(tickets, ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate actual tickets: %w", err)
	}

	return tickets, nil
}

func (r TicketRepository) GetArchive(
	ctx context.Context,
	userID int64,
) ([]Ticket, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
			id,
			user_id,
			trip_id,
			vehicle,
			departure_point,
			arrival_point,
			departure_time,
			arrival_time,
			cost,
			currency,
			departure_address,
			arrival_address,
			seat,
			number,
			comment
		FROM tickets
		WHERE user_id = $1
		  AND arrival_time < NOW()
		ORDER BY arrival_time DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get archive tickets: %w", err)
	}
	defer rows.Close()

	tickets := make([]Ticket, 0)

	for rows.Next() {
		var ticket Ticket

		if err := rows.Scan(
			&ticket.ID,
			&ticket.UserID,
			&ticket.TripID,
			&ticket.Vehicle,
			&ticket.DeparturePoint,
			&ticket.ArrivalPoint,
			&ticket.DepartureTime,
			&ticket.ArrivalTime,
			&ticket.Cost,
			&ticket.Currency,
			&ticket.DepartureAddress,
			&ticket.ArrivalAddress,
			&ticket.Seat,
			&ticket.Number,
			&ticket.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan archive ticket: %w", err)
		}

		tickets = append(tickets, ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive tickets: %w", err)
	}

	return tickets, nil
}

func (r TicketRepository) GetActualPage(
	ctx context.Context,
	userID int64,
	page int,
) ([]Ticket, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			vehicle,
			departure_point,
			arrival_point,
			departure_time,
			arrival_time,
			cost,
			currency,
			departure_address,
			arrival_address,
			seat,
			number,
			comment
		FROM tickets
		WHERE user_id = $1
		  AND arrival_time >= NOW()
		ORDER BY departure_time, id
		LIMIT $2 OFFSET $3
	`, userID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get actual tickets page: %w", err)
	}
	defer rows.Close()

	var result []Ticket

	for rows.Next() {
		var ticket Ticket

		if err := rows.Scan(
			&ticket.ID,
			&ticket.UserID,
			&ticket.TripID,
			&ticket.Vehicle,
			&ticket.DeparturePoint,
			&ticket.ArrivalPoint,
			&ticket.DepartureTime,
			&ticket.ArrivalTime,
			&ticket.Cost,
			&ticket.Currency,
			&ticket.DepartureAddress,
			&ticket.ArrivalAddress,
			&ticket.Seat,
			&ticket.Number,
			&ticket.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}

		result = append(result, ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tickets: %w", err)
	}

	return result, nil
}

func (r TicketRepository) GetArchivePage(
	ctx context.Context,
	userID int64,
	page int,
) ([]Ticket, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			vehicle,
			departure_point,
			arrival_point,
			departure_time,
			arrival_time,
			cost,
			currency,
			departure_address,
			arrival_address,
			seat,
			number,
			comment
		FROM tickets
		WHERE user_id = $1
		  AND arrival_time < NOW()
		ORDER BY departure_time DESC, id DESC
		LIMIT $2 OFFSET $3
	`, userID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get archive tickets page: %w", err)
	}
	defer rows.Close()

	var result []Ticket

	for rows.Next() {
		var ticket Ticket

		if err := rows.Scan(
			&ticket.ID,
			&ticket.UserID,
			&ticket.TripID,
			&ticket.Vehicle,
			&ticket.DeparturePoint,
			&ticket.ArrivalPoint,
			&ticket.DepartureTime,
			&ticket.ArrivalTime,
			&ticket.Cost,
			&ticket.Currency,
			&ticket.DepartureAddress,
			&ticket.ArrivalAddress,
			&ticket.Seat,
			&ticket.Number,
			&ticket.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}

		result = append(result, ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tickets: %w", err)
	}

	return result, nil
}

func (r TicketRepository) GetPageByTripID(
	ctx context.Context,
	tripID int64,
	userID int64,
	page int,
) ([]Ticket, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			vehicle,
			departure_point,
			arrival_point,
			departure_time,
			arrival_time,
			cost,
			currency,
			departure_address,
			arrival_address,
			seat,
			number,
			comment
		FROM tickets
		WHERE user_id = $1
		  AND trip_id = $2
		ORDER BY departure_time ASC, id ASC
		LIMIT $3 OFFSET $4
	`, userID, tripID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get tickets page by trip id: %w", err)
	}
	defer rows.Close()

	var result []Ticket

	for rows.Next() {
		var ticket Ticket

		if err := rows.Scan(
			&ticket.ID,
			&ticket.UserID,
			&ticket.TripID,
			&ticket.Vehicle,
			&ticket.DeparturePoint,
			&ticket.ArrivalPoint,
			&ticket.DepartureTime,
			&ticket.ArrivalTime,
			&ticket.Cost,
			&ticket.Currency,
			&ticket.DepartureAddress,
			&ticket.ArrivalAddress,
			&ticket.Seat,
			&ticket.Number,
			&ticket.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}

		result = append(result, ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tickets: %w", err)
	}

	return result, nil
}

func (r TicketRepository) GetForTime(
	ctx context.Context,
	userID int64,
	start time.Time,
	end time.Time,
	page int,
) ([]Ticket, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			vehicle,
			departure_point,
			arrival_point,
			departure_time,
			arrival_time,
			cost,
			departure_address,
			arrival_address,
			seat,
			comment
		FROM tickets
		WHERE user_id = $1
		  AND departure_time >= $2
		  AND departure_time < $3
		ORDER BY departure_time ASC, id ASC
		LIMIT $4 OFFSET $5
	`, userID, start, end, 2, page)
	if err != nil {
		return nil, fmt.Errorf("get tickets for time: %w", err)
	}
	defer rows.Close()

	result := make([]Ticket, 0, 2)

	for rows.Next() {
		var ticket Ticket

		if err := rows.Scan(
			&ticket.ID,
			&ticket.UserID,
			&ticket.TripID,
			&ticket.Vehicle,
			&ticket.DeparturePoint,
			&ticket.ArrivalPoint,
			&ticket.DepartureTime,
			&ticket.ArrivalTime,
			&ticket.Cost,
			&ticket.DepartureAddress,
			&ticket.ArrivalAddress,
			&ticket.Seat,
			&ticket.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}

		result = append(result, ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tickets: %w", err)
	}

	return result, nil
}
