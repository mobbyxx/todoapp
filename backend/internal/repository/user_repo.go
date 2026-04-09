package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/todo-api/internal/domain"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, display_name, avatar_url, is_active, preferences, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	preferencesJSON, err := json.Marshal(user.Preferences)
	if err != nil {
		return err
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = true

	return r.db.QueryRow(
		context.Background(),
		query,
		user.Email,
		user.PasswordHash,
		user.DisplayName,
		user.AvatarURL,
		user.IsActive,
		preferencesJSON,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
}

func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, avatar_url, is_active, email_verified_at,
		       last_seen_at, current_level_id, total_points, streak_count, streak_freeze_tokens,
		       last_streak_date, preferences, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = true
	`

	return r.scanUser(r.db.QueryRow(context.Background(), query, id))
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, avatar_url, is_active, email_verified_at,
		       last_seen_at, current_level_id, total_points, streak_count, streak_freeze_tokens,
		       last_streak_date, preferences, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
	`

	return r.scanUser(r.db.QueryRow(context.Background(), query, email))
}

func (r *userRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET display_name = $1, avatar_url = $2, preferences = $3, updated_at = $4
		WHERE id = $5 AND is_active = true
	`

	preferencesJSON, err := json.Marshal(user.Preferences)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(
		context.Background(),
		query,
		user.DisplayName,
		user.AvatarURL,
		preferencesJSON,
		time.Now(),
		user.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	query := `
		UPDATE users
		SET is_active = false, updated_at = $1
		WHERE id = $2 AND is_active = true
	`

	result, err := r.db.Exec(context.Background(), query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) UpdateLastSeen(id uuid.UUID) error {
	query := `
		UPDATE users
		SET last_seen_at = $1, updated_at = $1
		WHERE id = $2 AND is_active = true
	`

	result, err := r.db.Exec(context.Background(), query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) scanUser(row pgx.Row) (*domain.User, error) {
	var user domain.User
	var avatarURL, emailVerifiedAt, lastSeenAt, lastStreakDate *time.Time
	var currentLevelID *uuid.UUID
	var preferencesJSON []byte

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&avatarURL,
		&user.IsActive,
		&emailVerifiedAt,
		&lastSeenAt,
		&currentLevelID,
		&user.TotalPoints,
		&user.StreakCount,
		&user.StreakFreezeTokens,
		&lastStreakDate,
		&preferencesJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	user.AvatarURL = avatarURL
	user.EmailVerifiedAt = emailVerifiedAt
	user.LastSeenAt = lastSeenAt
	user.CurrentLevelID = currentLevelID
	user.LastStreakDate = lastStreakDate

	if preferencesJSON != nil {
		json.Unmarshal(preferencesJSON, &user.Preferences)
	}

	return &user, nil
}
