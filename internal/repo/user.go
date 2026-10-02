package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/warmte/tripl/lang"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID        int64
	Language  string
	CreatedAt time.Time
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) Get(
	ctx context.Context,
	userID int64,
) (*User, error) {
	var user User

	err := r.db.QueryRow(
		ctx,
		`SELECT id, language, created_at
		 FROM users
		 WHERE id = $1`,
		userID,
	).Scan(
		&user.ID,
		&user.Language,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by telegram id: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetOrCreate(
	ctx context.Context,
	telegramID int64,
	language string,
) (*User, error) {
	var user User

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO users (id, language)
		 VALUES ($1, $2)
		 ON CONFLICT (id)
		 DO UPDATE SET id = EXCLUDED.telegram_id
		 RETURNING id, language, created_at`,
		telegramID,
		language,
	).Scan(
		&user.ID,
		&user.Language,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get or create user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateLangOrCreate(
	ctx context.Context,
	telegramID int64,
	language lang.Language,
) (*User, error) {
	var user User

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO users (id, language)
		 VALUES ($1, $2)
		 ON CONFLICT (id)
		 DO UPDATE SET language = EXCLUDED.language
		 RETURNING id, language, created_at`,
		telegramID,
		language,
	).Scan(
		&user.ID,
		&user.Language,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("update language or create user: %w", err)
	}

	return &user, nil
}
