package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/todo-api/internal/domain"
)

type gamificationRepository struct {
	db *pgxpool.Pool
}

func NewGamificationRepository(db *pgxpool.Pool) domain.GamificationRepository {
	return &gamificationRepository{db: db}
}

func (r *gamificationRepository) GetAllBadges() ([]*domain.Badge, error) {
	query := `
		SELECT id, name, description, icon, points_value, created_at
		FROM badges
		ORDER BY points_value ASC
	`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBadges(rows)
}

func (r *gamificationRepository) GetBadgeByID(id uuid.UUID) (*domain.Badge, error) {
	query := `
		SELECT id, name, description, icon, points_value, created_at
		FROM badges
		WHERE id = $1
	`

	badge := &domain.Badge{}
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&badge.ID,
		&badge.Name,
		&badge.Description,
		&badge.Icon,
		&badge.PointsValue,
		&badge.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrBadgeNotFound
		}
		return nil, err
	}

	return badge, nil
}

func (r *gamificationRepository) GetUserBadges(userID uuid.UUID) ([]*domain.UserBadge, error) {
	query := `
		SELECT id, user_id, badge_id, earned_at
		FROM user_badges
		WHERE user_id = $1
		ORDER BY earned_at DESC
	`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanUserBadges(rows)
}

func (r *gamificationRepository) HasBadge(userID, badgeID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_badges 
			WHERE user_id = $1 AND badge_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(context.Background(), query, userID, badgeID).Scan(&exists)
	return exists, err
}

func (r *gamificationRepository) AwardBadge(userBadge *domain.UserBadge) error {
	query := `
		INSERT INTO user_badges (id, user_id, badge_id, earned_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(
		context.Background(),
		query,
		userBadge.ID,
		userBadge.UserID,
		userBadge.BadgeID,
		userBadge.EarnedAt,
	)

	return err
}

func (r *gamificationRepository) GetUserStats(userID uuid.UUID) (*domain.UserStats, error) {
	query := `
		SELECT total_points, streak_count, last_streak_date, 
		       (SELECT COUNT(*) FROM todos WHERE (created_by = $1 OR assigned_to = $1) AND status = 'completed' AND deleted_at IS NULL) as total_todos_completed
		FROM users
		WHERE id = $1 AND is_active = true
	`

	stats := &domain.UserStats{}
	var lastStreakDate *time.Time
	var totalPoints int
	var streakCount int

	err := r.db.QueryRow(context.Background(), query, userID).Scan(
		&totalPoints,
		&streakCount,
		&lastStreakDate,
		&stats.TotalTodosCompleted,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	stats.Points = totalPoints
	stats.Level = domain.CalculateLevel(totalPoints)
	stats.Streak = streakCount
	if lastStreakDate != nil {
		stats.LastActiveAt = *lastStreakDate
	}

	return stats, nil
}

func (r *gamificationRepository) CountCompletedTodos(userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM todos 
		WHERE (created_by = $1 OR assigned_to = $1) 
		  AND status = 'completed' 
		  AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(context.Background(), query, userID).Scan(&count)
	return count, err
}

func (r *gamificationRepository) CountEarlyCompletions(userID uuid.UUID, beforeHour int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM todos 
		WHERE (created_by = $1 OR assigned_to = $1) 
		  AND status = 'completed' 
		  AND deleted_at IS NULL
		  AND EXTRACT(HOUR FROM updated_at) < $2
	`

	var count int
	err := r.db.QueryRow(context.Background(), query, userID, beforeHour).Scan(&count)
	return count, err
}

func (r *gamificationRepository) CountConnections(userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM connections 
		WHERE (requester_id = $1 OR addressee_id = $1) 
		  AND status = 'accepted'
	`

	var count int
	err := r.db.QueryRow(context.Background(), query, userID).Scan(&count)
	return count, err
}

func (r *gamificationRepository) AddXP(userID uuid.UUID, amount int, reason string, referenceType string, referenceID uuid.UUID) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	updateQuery := `
		UPDATE users
		SET total_points = total_points + $2,
		    updated_at = $3
		WHERE id = $1 AND is_active = true
		RETURNING total_points
	`

	var newTotal int
	err = tx.QueryRow(context.Background(), updateQuery, userID, amount, time.Now()).Scan(&newTotal)
	if err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO points_history (id, user_id, amount, balance_after, reason, reference_type, reference_id, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.Exec(
		context.Background(),
		insertQuery,
		userID,
		amount,
		newTotal,
		reason,
		referenceType,
		referenceID,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (r *gamificationRepository) GetPointsHistory(userID uuid.UUID, limit int) ([]*domain.PointsTransaction, error) {
	query := `
		SELECT id, user_id, amount, reason, reference_type, reference_id, created_at
		FROM points_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(context.Background(), query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.PointsTransaction
	for rows.Next() {
		var t domain.PointsTransaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Reason, &t.ReferenceType, &t.ReferenceID, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &t)
	}

	return transactions, rows.Err()
}

func (r *gamificationRepository) GetUserXP(userID uuid.UUID) (int, error) {
	query := `
		SELECT total_points
		FROM users
		WHERE id = $1 AND is_active = true
	`

	var xp int
	err := r.db.QueryRow(context.Background(), query, userID).Scan(&xp)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, domain.ErrUserNotFound
		}
		return 0, err
	}

	return xp, nil
}

func (r *gamificationRepository) UpdateStreak(userID uuid.UUID, newStreak int, lastStreakDate time.Time) error {
	query := `
		UPDATE users
		SET streak_count = $2,
		    last_streak_date = $3,
		    updated_at = $4
		WHERE id = $1 AND is_active = true
	`

	result, err := r.db.Exec(
		context.Background(),
		query,
		userID,
		newStreak,
		lastStreakDate,
		time.Now(),
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *gamificationRepository) GetStreakInfo(userID uuid.UUID) (streakCount int, lastStreakDate *time.Time, freezeTokens int, err error) {
	query := `
		SELECT streak_count, last_streak_date, streak_freeze_tokens
		FROM users
		WHERE id = $1 AND is_active = true
	`

	err = r.db.QueryRow(context.Background(), query, userID).Scan(
		&streakCount,
		&lastStreakDate,
		&freezeTokens,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil, 0, domain.ErrUserNotFound
		}
		return 0, nil, 0, err
	}

	return streakCount, lastStreakDate, freezeTokens, nil
}

func (r *gamificationRepository) UseFreezeToken(userID uuid.UUID) error {
	query := `
		UPDATE users
		SET streak_freeze_tokens = streak_freeze_tokens - 1,
		    updated_at = $2
		WHERE id = $1 
		  AND is_active = true 
		  AND streak_freeze_tokens > 0
	`

	result, err := r.db.Exec(context.Background(), query, userID, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrInsufficientFreezeTokens
	}

	return nil
}

func (r *gamificationRepository) AwardFreezeToken(userID uuid.UUID) error {
	query := `
		UPDATE users
		SET streak_freeze_tokens = streak_freeze_tokens + 1,
		    updated_at = $2
		WHERE id = $1 AND is_active = true
	`

	result, err := r.db.Exec(context.Background(), query, userID, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *gamificationRepository) scanBadges(rows pgx.Rows) ([]*domain.Badge, error) {
	var badges []*domain.Badge

	for rows.Next() {
		var badge domain.Badge
		err := rows.Scan(
			&badge.ID,
			&badge.Name,
			&badge.Description,
			&badge.Icon,
			&badge.PointsValue,
			&badge.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		badges = append(badges, &badge)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return badges, nil
}

func (r *gamificationRepository) scanUserBadges(rows pgx.Rows) ([]*domain.UserBadge, error) {
	var userBadges []*domain.UserBadge

	for rows.Next() {
		var ub domain.UserBadge
		err := rows.Scan(
			&ub.ID,
			&ub.UserID,
			&ub.BadgeID,
			&ub.EarnedAt,
		)
		if err != nil {
			return nil, err
		}
		userBadges = append(userBadges, &ub)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userBadges, nil
}

func (r *gamificationRepository) GetBadgeCriteria(badgeID uuid.UUID) (*domain.BadgeCriteria, error) {
	query := `
		SELECT criteria
		FROM badges
		WHERE id = $1
	`

	var criteriaJSON []byte
	err := r.db.QueryRow(context.Background(), query, badgeID).Scan(&criteriaJSON)
	if err != nil {
		return nil, err
	}

	var criteria domain.BadgeCriteria
	if err := json.Unmarshal(criteriaJSON, &criteria); err != nil {
		return nil, err
	}

	return &criteria, nil
}
