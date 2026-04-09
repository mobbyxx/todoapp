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

type sharedGoalRepository struct {
	db *pgxpool.Pool
}

func NewSharedGoalRepository(db *pgxpool.Pool) domain.SharedGoalRepository {
	return &sharedGoalRepository{db: db}
}

func (r *sharedGoalRepository) Create(goal *domain.SharedGoal) error {
	query := `
		INSERT INTO shared_goals (id, connection_id, target_type, target_value, current_value, 
		                          reward_description, status, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	goal.ID = uuid.New()
	goal.CreatedAt = time.Now()
	goal.Status = domain.SharedGoalStatusActive
	goal.CurrentValue = 0

	_, err := r.db.Exec(
		context.Background(),
		query,
		goal.ID,
		goal.ConnectionID,
		goal.TargetType,
		goal.TargetValue,
		goal.CurrentValue,
		goal.RewardDescription,
		goal.Status,
		goal.CreatedAt,
		goal.CompletedAt,
	)

	return err
}

func (r *sharedGoalRepository) GetByID(id uuid.UUID) (*domain.SharedGoal, error) {
	query := `
		SELECT id, connection_id, target_type, target_value, current_value,
		       reward_description, status, created_at, completed_at
		FROM shared_goals
		WHERE id = $1
	`

	return r.scanGoal(r.db.QueryRow(context.Background(), query, id))
}

func (r *sharedGoalRepository) GetByConnection(connectionID uuid.UUID) ([]*domain.SharedGoal, error) {
	query := `
		SELECT id, connection_id, target_type, target_value, current_value,
		       reward_description, status, created_at, completed_at
		FROM shared_goals
		WHERE connection_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(context.Background(), query, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanGoals(rows)
}

func (r *sharedGoalRepository) GetByUserID(userID uuid.UUID) ([]*domain.SharedGoal, error) {
	query := `
		SELECT sg.id, sg.connection_id, sg.target_type, sg.target_value, sg.current_value,
		       sg.reward_description, sg.status, sg.created_at, sg.completed_at
		FROM shared_goals sg
		INNER JOIN connections c ON sg.connection_id = c.id
		WHERE (c.user_a_id = $1 OR c.user_b_id = $1)
		  AND c.status = 'accepted'
		ORDER BY sg.created_at DESC
	`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanGoals(rows)
}

func (r *sharedGoalRepository) UpdateProgress(id uuid.UUID, amount int) error {
	query := `
		UPDATE shared_goals
		SET current_value = current_value + $2,
		    status = CASE 
		        WHEN current_value + $2 >= target_value THEN 'completed'::shared_goal_status
		        ELSE status
		    END,
		    completed_at = CASE 
		        WHEN current_value + $2 >= target_value THEN $3
		        ELSE completed_at
		    END
		WHERE id = $1
		RETURNING current_value, target_value, status
	`

	now := time.Now()
	var currentValue, targetValue int
	var status domain.SharedGoalStatus

	err := r.db.QueryRow(context.Background(), query, id, amount, now).Scan(
		&currentValue, &targetValue, &status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrSharedGoalNotFound
		}
		return err
	}

	return nil
}

func (r *sharedGoalRepository) Update(goal *domain.SharedGoal) error {
	query := `
		UPDATE shared_goals
		SET connection_id = $1, target_type = $2, target_value = $3, current_value = $4,
		    reward_description = $5, status = $6, completed_at = $7
		WHERE id = $8
	`

	result, err := r.db.Exec(
		context.Background(),
		query,
		goal.ConnectionID,
		goal.TargetType,
		goal.TargetValue,
		goal.CurrentValue,
		goal.RewardDescription,
		goal.Status,
		goal.CompletedAt,
		goal.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrSharedGoalNotFound
	}

	return nil
}

func (r *sharedGoalRepository) scanGoal(row pgx.Row) (*domain.SharedGoal, error) {
	var goal domain.SharedGoal
	var completedAt *time.Time

	err := row.Scan(
		&goal.ID,
		&goal.ConnectionID,
		&goal.TargetType,
		&goal.TargetValue,
		&goal.CurrentValue,
		&goal.RewardDescription,
		&goal.Status,
		&goal.CreatedAt,
		&completedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSharedGoalNotFound
		}
		return nil, err
	}

	goal.CompletedAt = completedAt
	return &goal, nil
}

func (r *sharedGoalRepository) scanGoals(rows pgx.Rows) ([]*domain.SharedGoal, error) {
	var goals []*domain.SharedGoal

	for rows.Next() {
		var goal domain.SharedGoal
		var completedAt *time.Time

		err := rows.Scan(
			&goal.ID,
			&goal.ConnectionID,
			&goal.TargetType,
			&goal.TargetValue,
			&goal.CurrentValue,
			&goal.RewardDescription,
			&goal.Status,
			&goal.CreatedAt,
			&completedAt,
		)
		if err != nil {
			return nil, err
		}

		goal.CompletedAt = completedAt
		goals = append(goals, &goal)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return goals, nil
}
