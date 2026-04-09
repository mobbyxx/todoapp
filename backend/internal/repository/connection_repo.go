package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/todo-api/internal/domain"
)

type connectionRepository struct {
	db *pgxpool.Pool
}

func NewConnectionRepository(db *pgxpool.Pool) domain.ConnectionRepository {
	return &connectionRepository{db: db}
}

func (r *connectionRepository) Create(connection *domain.Connection) error {
	query := `
		INSERT INTO connections (user_a_id, user_b_id, status, requested_by, invitation_token, 
		                         expires_at, accepted_at, rejected_at, blocked_at, block_reason, 
		                         created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	now := time.Now()
	connection.CreatedAt = now
	connection.UpdatedAt = now

	err := r.db.QueryRow(
		context.Background(),
		query,
		connection.UserAID,
		connection.UserBID,
		connection.Status,
		connection.RequestedBy,
		connection.InvitationToken,
		connection.ExpiresAt,
		connection.AcceptedAt,
		connection.RejectedAt,
		connection.BlockedAt,
		connection.BlockReason,
		connection.CreatedAt,
		connection.UpdatedAt,
	).Scan(&connection.ID)

	return err
}

func (r *connectionRepository) GetByID(id uuid.UUID) (*domain.Connection, error) {
	query := `
		SELECT id, user_a_id, user_b_id, status, requested_by, invitation_token,
		       expires_at, accepted_at, rejected_at, blocked_at, block_reason,
		       created_at, updated_at
		FROM connections
		WHERE id = $1
	`

	return r.scanConnection(r.db.QueryRow(context.Background(), query, id))
}

func (r *connectionRepository) GetByToken(token string) (*domain.Connection, error) {
	query := `
		SELECT id, user_a_id, user_b_id, status, requested_by, invitation_token,
		       expires_at, accepted_at, rejected_at, blocked_at, block_reason,
		       created_at, updated_at
		FROM connections
		WHERE invitation_token = $1
	`

	return r.scanConnection(r.db.QueryRow(context.Background(), query, token))
}

func (r *connectionRepository) GetByUserID(userID uuid.UUID) ([]*domain.Connection, error) {
	query := `
		SELECT id, user_a_id, user_b_id, status, requested_by, invitation_token,
		       expires_at, accepted_at, rejected_at, blocked_at, block_reason,
		       created_at, updated_at
		FROM connections
		WHERE user_a_id = $1 OR user_b_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanConnections(rows)
}

func (r *connectionRepository) GetByUserPair(userAID, userBID uuid.UUID) (*domain.Connection, error) {
	query := `
		SELECT id, user_a_id, user_b_id, status, requested_by, invitation_token,
		       expires_at, accepted_at, rejected_at, blocked_at, block_reason,
		       created_at, updated_at
		FROM connections
		WHERE (user_a_id = $1 AND user_b_id = $2) OR (user_a_id = $2 AND user_b_id = $1)
	`

	conn, err := r.scanConnection(r.db.QueryRow(context.Background(), query, userAID, userBID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConnectionNotFound
		}
		return nil, err
	}
	return conn, nil
}

func (r *connectionRepository) Update(connection *domain.Connection) error {
	query := `
		UPDATE connections
		SET user_a_id = $1, user_b_id = $2, status = $3, requested_by = $4, 
		    invitation_token = $5, expires_at = $6, accepted_at = $7, rejected_at = $8, 
		    blocked_at = $9, block_reason = $10, updated_at = $11
		WHERE id = $12
	`

	connection.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		context.Background(),
		query,
		connection.UserAID,
		connection.UserBID,
		connection.Status,
		connection.RequestedBy,
		connection.InvitationToken,
		connection.ExpiresAt,
		connection.AcceptedAt,
		connection.RejectedAt,
		connection.BlockedAt,
		connection.BlockReason,
		connection.UpdatedAt,
		connection.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrConnectionNotFound
	}

	return nil
}

func (r *connectionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM connections WHERE id = $1`

	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrConnectionNotFound
	}

	return nil
}

func (r *connectionRepository) scanConnection(row pgx.Row) (*domain.Connection, error) {
	var conn domain.Connection
	var expiresAt, acceptedAt, rejectedAt, blockedAt *time.Time
	var invitationToken, blockReason *string

	err := row.Scan(
		&conn.ID,
		&conn.UserAID,
		&conn.UserBID,
		&conn.Status,
		&conn.RequestedBy,
		&invitationToken,
		&expiresAt,
		&acceptedAt,
		&rejectedAt,
		&blockedAt,
		&blockReason,
		&conn.CreatedAt,
		&conn.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConnectionNotFound
		}
		return nil, err
	}

	if invitationToken != nil {
		conn.InvitationToken = *invitationToken
	}
	conn.ExpiresAt = expiresAt
	conn.AcceptedAt = acceptedAt
	conn.RejectedAt = rejectedAt
	conn.BlockedAt = blockedAt
	if blockReason != nil {
		conn.BlockReason = *blockReason
	}

	return &conn, nil
}

func (r *connectionRepository) scanConnections(rows pgx.Rows) ([]*domain.Connection, error) {
	var connections []*domain.Connection

	for rows.Next() {
		var conn domain.Connection
		var expiresAt, acceptedAt, rejectedAt, blockedAt *time.Time
		var invitationToken, blockReason *string

		err := rows.Scan(
			&conn.ID,
			&conn.UserAID,
			&conn.UserBID,
			&conn.Status,
			&conn.RequestedBy,
			&invitationToken,
			&expiresAt,
			&acceptedAt,
			&rejectedAt,
			&blockedAt,
			&blockReason,
			&conn.CreatedAt,
			&conn.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if invitationToken != nil {
			conn.InvitationToken = *invitationToken
		}
		conn.ExpiresAt = expiresAt
		conn.AcceptedAt = acceptedAt
		conn.RejectedAt = rejectedAt
		conn.BlockedAt = blockedAt
		if blockReason != nil {
			conn.BlockReason = *blockReason
		}

		connections = append(connections, &conn)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return connections, nil
}
