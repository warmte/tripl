package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) GetOrCreate(
	ctx context.Context,
	userID int64,
	language string,
) (*User, error) {
	var user User

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO users (id, language)
		 VALUES ($1, $2)
		 ON CONFLICT (id)
		 DO UPDATE SET id = users.telegram_id
		 RETURNING id, language, created_at`,
		userID,
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
		 DO UPDATE SET language = users.language
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

func (r UserRepository) GetAll(ctx context.Context) ([]User, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
			id,
			language,
			created_at
		FROM users
		ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)

	for rows.Next() {
		var user User

		if err := rows.Scan(
			&user.ID,
			&user.Language,
			&user.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (r UserRepository) GetLanguage(ctx context.Context, userID int64) (string, error) {
	user, err := r.GetOrCreate(ctx, userID, string(lang.LanguageRu))

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	}

	if err != nil {
		return "", fmt.Errorf("get user language: %w", err)
	}

	return user.Language, nil
}

func (r UserRepository) TryConsume(
	ctx context.Context,
	userID int64,
	limit int,
	date time.Time,
) (bool, error) {
	if userID <= 0 {
		return false, errors.New("invalid user ID")
	}

	tag, err := r.db.Exec(ctx, `
		INSERT INTO user_daily_usage (user_id, usage_date, upload_count)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, usage_date)
		DO UPDATE SET upload_count = user_daily_usage.upload_count + 1
		WHERE user_daily_usage.upload_count < $3
	`, userID, date.UTC(), limit)
	if err != nil {
		return false, fmt.Errorf("reserve daily upload: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r UserRepository) GetCount(
	ctx context.Context,
	userID int64,
	date time.Time,
) (int, error) {
	var count int

	err := r.db.QueryRow(ctx, `
		SELECT upload_count
		FROM user_daily_usage
		WHERE user_id = $1 AND usage_date = $2
	`, userID, date.UTC()).Scan(&count)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get daily upload count: %w", err)
	}

	return count, nil
}

func (r UserRepository) DeleteOlderThan(ctx context.Context, date time.Time) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM user_daily_usage
		WHERE usage_date < $1::date - 7
	`, date.UTC())
	if err != nil {
		return fmt.Errorf("delete old daily usage: %w", err)
	}

	return nil
}
