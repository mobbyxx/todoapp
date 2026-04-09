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

type rewardRepository struct {
	db *pgxpool.Pool
}

func NewRewardRepository(db *pgxpool.Pool) domain.RewardRepository {
	return &rewardRepository{db: db}
}

func (r *rewardRepository) Create(reward *domain.Reward) error {
	query := `
		INSERT INTO rewards (id, user_id, name, description, cost_points, is_active, type, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'points', 0, $7, $8)
		RETURNING id
	`

	now := time.Now()
	reward.ID = uuid.New()
	reward.CreatedAt = now
	reward.UpdatedAt = now

	return r.db.QueryRow(
		context.Background(),
		query,
		reward.ID,
		reward.UserID,
		reward.Name,
		reward.Description,
		reward.Cost,
		reward.IsActive,
		reward.CreatedAt,
		reward.UpdatedAt,
	).Scan(&reward.ID)
}

func (r *rewardRepository) GetByID(id uuid.UUID) (*domain.Reward, error) {
	query := `
		SELECT id, user_id, name, description, cost_points, is_active, created_at, updated_at
		FROM rewards
		WHERE id = $1 AND deleted_at IS NULL
	`

	return r.scanReward(r.db.QueryRow(context.Background(), query, id))
}

func (r *rewardRepository) GetByUserID(userID uuid.UUID) ([]*domain.Reward, error) {
	query := `
		SELECT id, user_id, name, description, cost_points, is_active, created_at, updated_at
		FROM rewards
		WHERE (user_id = $1 OR user_id IS NULL) 
		  AND is_active = true 
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRewards(rows)
}

func (r *rewardRepository) GetAllActive() ([]*domain.Reward, error) {
	query := `
		SELECT id, user_id, name, description, cost_points, is_active, created_at, updated_at
		FROM rewards
		WHERE is_active = true AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRewards(rows)
}

func (r *rewardRepository) Update(reward *domain.Reward) error {
	query := `
		UPDATE rewards
		SET name = $1, description = $2, cost_points = $3, is_active = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
		RETURNING updated_at
	`

	reward.UpdatedAt = time.Now()

	err := r.db.QueryRow(
		context.Background(),
		query,
		reward.Name,
		reward.Description,
		reward.Cost,
		reward.IsActive,
		reward.UpdatedAt,
		reward.ID,
	).Scan(&reward.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrRewardNotFound
		}
		return err
	}

	return nil
}

func (r *rewardRepository) Delete(id uuid.UUID) error {
	query := `
		UPDATE rewards
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(context.Background(), query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrRewardNotFound
	}

	return nil
}

func (r *rewardRepository) CreateRedemption(redemption *domain.RewardRedemption) error {
	query := `
		INSERT INTO reward_redemptions (id, user_id, reward_id, points_spent, status, redeemed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	redemption.ID = uuid.New()
	redemption.RedeemedAt = now
	redemption.CreatedAt = now
	redemption.UpdatedAt = now

	_, err := r.db.Exec(
		context.Background(),
		query,
		redemption.ID,
		redemption.UserID,
		redemption.RewardID,
		0,
		redemption.Status,
		redemption.RedeemedAt,
		redemption.CreatedAt,
		redemption.UpdatedAt,
	)

	return err
}

func (r *rewardRepository) GetRedemptionByID(id uuid.UUID) (*domain.RewardRedemption, error) {
	query := `
		SELECT id, user_id, reward_id, status, redeemed_at, created_at, updated_at
		FROM reward_redemptions
		WHERE id = $1
	`

	return r.scanRedemption(r.db.QueryRow(context.Background(), query, id))
}

func (r *rewardRepository) GetRedemptionsByUser(userID uuid.UUID) ([]*domain.RewardRedemption, error) {
	query := `
		SELECT id, user_id, reward_id, status, redeemed_at, created_at, updated_at
		FROM reward_redemptions
		WHERE user_id = $1
		ORDER BY redeemed_at DESC
	`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRedemptions(rows)
}

func (r *rewardRepository) GetRedemptionsByReward(rewardID uuid.UUID) ([]*domain.RewardRedemption, error) {
	query := `
		SELECT id, user_id, reward_id, status, redeemed_at, created_at, updated_at
		FROM reward_redemptions
		WHERE reward_id = $1
		ORDER BY redeemed_at DESC
	`

	rows, err := r.db.Query(context.Background(), query, rewardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRedemptions(rows)
}

func (r *rewardRepository) UpdateRedemptionStatus(id uuid.UUID, status domain.RedemptionStatus) error {
	query := `
		UPDATE reward_redemptions
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(context.Background(), query, status, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrRedemptionNotFound
	}

	return nil
}

func (r *rewardRepository) scanReward(row pgx.Row) (*domain.Reward, error) {
	var reward domain.Reward
	var userID *uuid.UUID

	err := row.Scan(
		&reward.ID,
		&userID,
		&reward.Name,
		&reward.Description,
		&reward.Cost,
		&reward.IsActive,
		&reward.CreatedAt,
		&reward.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRewardNotFound
		}
		return nil, err
	}

	reward.UserID = userID
	return &reward, nil
}

func (r *rewardRepository) scanRewards(rows pgx.Rows) ([]*domain.Reward, error) {
	var rewards []*domain.Reward

	for rows.Next() {
		var reward domain.Reward
		var userID *uuid.UUID

		err := rows.Scan(
			&reward.ID,
			&userID,
			&reward.Name,
			&reward.Description,
			&reward.Cost,
			&reward.IsActive,
			&reward.CreatedAt,
			&reward.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		reward.UserID = userID
		rewards = append(rewards, &reward)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rewards, nil
}

func (r *rewardRepository) scanRedemption(row pgx.Row) (*domain.RewardRedemption, error) {
	var redemption domain.RewardRedemption

	err := row.Scan(
		&redemption.ID,
		&redemption.UserID,
		&redemption.RewardID,
		&redemption.Status,
		&redemption.RedeemedAt,
		&redemption.CreatedAt,
		&redemption.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRedemptionNotFound
		}
		return nil, err
	}

	return &redemption, nil
}

func (r *rewardRepository) scanRedemptions(rows pgx.Rows) ([]*domain.RewardRedemption, error) {
	var redemptions []*domain.RewardRedemption

	for rows.Next() {
		var redemption domain.RewardRedemption

		err := rows.Scan(
			&redemption.ID,
			&redemption.UserID,
			&redemption.RewardID,
			&redemption.Status,
			&redemption.RedeemedAt,
			&redemption.CreatedAt,
			&redemption.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		redemptions = append(redemptions, &redemption)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return redemptions, nil
}
