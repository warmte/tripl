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

var ErrBookingNotFound = errors.New("booking not found")

type BookingRepository struct {
	BasicRepository
}

var _ Repository[Booking] = BookingRepository{} // compile-time check

func NewBookingRepository(db *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{
		MakeBasic(db, BookingsRepo, ErrBookingNotFound),
	}
}

func (r BookingRepository) Create(
	ctx context.Context,
	booking *Booking,
) (*Booking, error) {
	var bookingRepr Booking

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO bookings (
			user_id,
			city,
			place,
			check_in_date,
			check_out_date,
			cost,
			currency,
			address,
			check_in_time,
			check_out_time,
			number,
			comment
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING
			id,
			user_id,
			trip_id,
			city,
			place,
			check_in_date,
			check_out_date,
			cost,
			currency,
			address,
			check_in_time,
			check_out_time,
			number,
			comment`,
		booking.UserID,
		booking.City,
		booking.Place,
		booking.CheckInDate,
		booking.CheckOutDate,
		booking.Cost,
		booking.Currency,
		booking.Address,
		booking.CheckInTime,
		booking.CheckOutTime,
		booking.Number,
		booking.Comment,
	).Scan(
		&bookingRepr.ID,
		&bookingRepr.UserID,
		&bookingRepr.TripID,
		&bookingRepr.City,
		&bookingRepr.Place,
		&bookingRepr.CheckInDate,
		&bookingRepr.CheckOutDate,
		&bookingRepr.Cost,
		&bookingRepr.Currency,
		&bookingRepr.Address,
		&bookingRepr.CheckInTime,
		&bookingRepr.CheckOutTime,
		&bookingRepr.Number,
		&bookingRepr.Comment,
	)

	if err != nil {
		return nil, fmt.Errorf("create booking: %w", err)
	}

	return &bookingRepr, nil
}

func (r BookingRepository) GetByID(
	ctx context.Context,
	id int64,
	userID int64,
) (*Booking, error) {
	var bookingRepr Booking

	err := r.db.QueryRow(
		ctx,
		`SELECT
			id,
			user_id,
			trip_id,
			city,
			place,
			check_in_date,
			check_out_date,
			cost,
			currency,
			address,
			check_in_time,
			check_out_time,
			number,
			comment
		FROM bookings
		WHERE id = $1
		  AND user_id = $2`,
		id,
		userID,
	).Scan(
		&bookingRepr.ID,
		&bookingRepr.UserID,
		&bookingRepr.TripID,
		&bookingRepr.City,
		&bookingRepr.Place,
		&bookingRepr.CheckInDate,
		&bookingRepr.CheckOutDate,
		&bookingRepr.Cost,
		&bookingRepr.Currency,
		&bookingRepr.Address,
		&bookingRepr.CheckInTime,
		&bookingRepr.CheckOutTime,
		&bookingRepr.Number,
		&bookingRepr.Comment,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBookingNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get booking: %w", err)
	}

	return &bookingRepr, nil
}

func (r *BookingRepository) UpdateCity(
	ctx context.Context,
	id int64,
	userID int64,
	city string,
) error {
	return r.UpdateField(ctx, id, userID, "city", city)
}

func (r *BookingRepository) UpdatePlace(
	ctx context.Context,
	id int64,
	userID int64,
	place string,
) error {
	return r.UpdateField(ctx, id, userID, "place", place)
}

func (r *BookingRepository) UpdateCheckInDate(
	ctx context.Context,
	id int64,
	userID int64,
	checkInDate time.Time,
) error {
	return r.UpdateField(ctx, id, userID, "check_in_date", checkInDate)
}

func (r *BookingRepository) UpdateCheckOutDate(
	ctx context.Context,
	id int64,
	userID int64,
	checkOutDate time.Time,
) error {
	return r.UpdateField(ctx, id, userID, "check_out_date", checkOutDate)
}

func (r *BookingRepository) UpdateCost(
	ctx context.Context,
	id int64,
	userID int64,
	cost *string,
) error {
	return r.UpdateField(ctx, id, userID, "cost", cost)
}

func (r *BookingRepository) UpdateCurrency(
	ctx context.Context,
	id int64,
	userID int64,
	currency *string,
) error {
	return r.UpdateField(ctx, id, userID, "currency", currency)
}

func (r *BookingRepository) UpdateAddress(
	ctx context.Context,
	id int64,
	userID int64,
	address *string,
) error {
	return r.UpdateField(ctx, id, userID, "address", address)
}

func (r *BookingRepository) UpdateCheckInTime(
	ctx context.Context,
	id int64,
	userID int64,
	checkInTime *time.Time,
) error {
	return r.UpdateField(ctx, id, userID, "check_in_time", checkInTime)
}

func (r *BookingRepository) UpdateCheckOutTime(
	ctx context.Context,
	id int64,
	userID int64,
	checkOutTime *time.Time,
) error {
	return r.UpdateField(ctx, id, userID, "check_out_time", checkOutTime)
}

func (r *BookingRepository) UpdateNumber(
	ctx context.Context,
	id int64,
	userID int64,
	number *string,
) error {
	return r.UpdateField(ctx, id, userID, "number", number)
}

func (r *BookingRepository) UpdateComment(
	ctx context.Context,
	id int64,
	userID int64,
	comment *string,
) error {
	return r.UpdateField(ctx, id, userID, "comment", comment)
}

func (r BookingRepository) GetActual(
	ctx context.Context,
	userID int64,
) ([]Booking, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
			id,
			user_id,
			trip_id,
			city,
			place,
			check_in_date,
			check_out_date,
			cost,
			currency,
			address,
			check_in_time,
			check_out_time,
			number,
			comment
		FROM bookings
		WHERE user_id = $1
		  AND check_out_date >= CURRENT_DATE
		ORDER BY check_in_date ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get actual bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]Booking, 0)

	for rows.Next() {
		var booking Booking

		if err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.TripID,
			&booking.City,
			&booking.Place,
			&booking.CheckInDate,
			&booking.CheckOutDate,
			&booking.Cost,
			&booking.Currency,
			&booking.Address,
			&booking.CheckInTime,
			&booking.CheckOutTime,
			&booking.Number,
			&booking.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan actual booking: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate actual bookings: %w", err)
	}

	return bookings, nil
}

func (r BookingRepository) GetArchive(
	ctx context.Context,
	userID int64,
) ([]Booking, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
			id,
			user_id,
			trip_id,
			city,
			place,
			check_in_date,
			check_out_date,
			cost,
			currency,
			address,
			check_in_time,
			check_out_time,
			number,
			comment
		FROM bookings
		WHERE user_id = $1
		  AND check_out_date < CURRENT_DATE
		ORDER BY check_out_date DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get archive bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]Booking, 0)

	for rows.Next() {
		var booking Booking

		if err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.TripID,
			&booking.City,
			&booking.Place,
			&booking.CheckInDate,
			&booking.CheckOutDate,
			&booking.Cost,
			&booking.Currency,
			&booking.Address,
			&booking.CheckInTime,
			&booking.CheckOutTime,
			&booking.Number,
			&booking.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan archive booking: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive bookings: %w", err)
	}

	return bookings, nil
}

func (r BookingRepository) GetActualPage(
	ctx context.Context,
	userID int64,
	page int,
) ([]Booking, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			city,
			place,
			check_in_date,
			check_out_date,
			cost,
			currency,
			address,
			check_in_time,
			check_out_time,
			number,
			comment
		FROM bookings
		WHERE user_id = $1
		  AND check_out_date >= CURRENT_DATE
		ORDER BY check_in_date, id
		LIMIT $2 OFFSET $3
	`, userID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get actual bookings page: %w", err)
	}
	defer rows.Close()

	var result []Booking

	for rows.Next() {
		var booking Booking

		if err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.TripID,
			&booking.City,
			&booking.Place,
			&booking.CheckInDate,
			&booking.CheckOutDate,
			&booking.Cost,
			&booking.Currency,
			&booking.Address,
			&booking.CheckInTime,
			&booking.CheckOutTime,
			&booking.Number,
			&booking.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}

		result = append(result, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}

	return result, nil
}

func (r BookingRepository) GetArchivePage(
	ctx context.Context,
	userID int64,
	page int,
) ([]Booking, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			city,
			place,
			check_in_date,
			check_out_date,
			cost,
			currency,
			address,
			check_in_time,
			check_out_time,
			number,
			comment
		FROM bookings
		WHERE user_id = $1
		  AND check_out_date < CURRENT_DATE
		ORDER BY check_out_date DESC, id DESC
		LIMIT $2 OFFSET $3
	`, userID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get archive bookings page: %w", err)
	}
	defer rows.Close()

	var result []Booking

	for rows.Next() {
		var booking Booking

		if err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.TripID,
			&booking.City,
			&booking.Place,
			&booking.CheckInDate,
			&booking.CheckOutDate,
			&booking.Cost,
			&booking.Currency,
			&booking.Address,
			&booking.CheckInTime,
			&booking.CheckOutTime,
			&booking.Number,
			&booking.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}

		result = append(result, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}

	return result, nil
}

func (r BookingRepository) GetPageByTripID(ctx context.Context, tripID int64, userID int64, page int) ([]Booking, error) {
	if page < 0 {
		return nil, fmt.Errorf("invalid page: %d", page)
	}

	offset := page * PageSize

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			trip_id,
			city,
			place,
			check_in_date,
			check_out_date,
			cost,
			currency,
			address,
			check_in_time,
			check_out_time,
			number,
			comment
		FROM bookings
		WHERE user_id = $1
		  AND trip_id = $2
		ORDER BY check_in_date ASC, id ASC
		LIMIT $3 OFFSET $4
	`, userID, tripID, PageSize+1, offset)
	if err != nil {
		return nil, fmt.Errorf("get bookings page by trip id: %w", err)
	}
	defer rows.Close()

	var result []Booking

	for rows.Next() {
		var booking Booking

		if err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.TripID,
			&booking.City,
			&booking.Place,
			&booking.CheckInDate,
			&booking.CheckOutDate,
			&booking.Cost,
			&booking.Currency,
			&booking.Address,
			&booking.CheckInTime,
			&booking.CheckOutTime,
			&booking.Number,
			&booking.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}

		result = append(result, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}

	return result, nil
}
