package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/todo-api/internal/infrastructure/push"
)

type pushTokenRepository struct {
	db *pgxpool.Pool
}

func NewPushTokenRepository(db *pgxpool.Pool) *pushTokenRepository {
	return &pushTokenRepository{db: db}
}

func (r *pushTokenRepository) GetActiveTokensByUserID(userID uuid.UUID) ([]push.TokenInfo, error) {
	query := `
		SELECT id, user_id, token, platform
		FROM push_notification_tokens
		WHERE user_id = $1 AND is_active = true
		AND (expires_at IS NULL OR expires_at > NOW())
	`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTokens(rows)
}

func (r *pushTokenRepository) GetAllActiveTokens(limit int) ([]push.TokenInfo, error) {
	query := `
		SELECT id, user_id, token, platform
		FROM push_notification_tokens
		WHERE is_active = true
		AND (expires_at IS NULL OR expires_at > NOW())
		LIMIT $1
	`

	rows, err := r.db.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTokens(rows)
}

func (r *pushTokenRepository) DeactivateToken(token string) error {
	query := `
		UPDATE push_notification_tokens
		SET is_active = false, updated_at = NOW()
		WHERE token = $1
	`

	_, err := r.db.Exec(context.Background(), query, token)
	return err
}

func (r *pushTokenRepository) DeleteToken(token string) error {
	query := `DELETE FROM push_notification_tokens WHERE token = $1`

	_, err := r.db.Exec(context.Background(), query, token)
	return err
}

func (r *pushTokenRepository) UpdateLastUsed(token string) error {
	query := `
		UPDATE push_notification_tokens
		SET last_used_at = NOW(), updated_at = NOW()
		WHERE token = $1
	`

	_, err := r.db.Exec(context.Background(), query, token)
	return err
}

func (r *pushTokenRepository) scanTokens(rows pgx.Rows) ([]push.TokenInfo, error) {
	var tokens []push.TokenInfo

	for rows.Next() {
		var id, userID uuid.UUID
		var token, platform string

		err := rows.Scan(&id, &userID, &token, &platform)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, push.TokenInfo{
			ID:       id.String(),
			UserID:   userID.String(),
			Token:    token,
			Platform: platform,
		})
	}

	return tokens, rows.Err()
}

func (r *pushTokenRepository) CreateDeadLetterEntry(userID uuid.UUID, notificationType string, payload []byte, errorMessage string) error {
	query := `
		INSERT INTO notification_dead_letter (user_id, type, payload, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(context.Background(), query, userID, notificationType, payload, errorMessage, time.Now())
	return err
}
