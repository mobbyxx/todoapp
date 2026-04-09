package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/todo-api/internal/domain"
)

type notificationQueueRepository struct {
	db *pgxpool.Pool
}

func NewNotificationQueueRepository(db *pgxpool.Pool) domain.NotificationQueueRepository {
	return &notificationQueueRepository{db: db}
}

func (r *notificationQueueRepository) Enqueue(item *domain.NotificationQueueItem) error {
	query := `
		INSERT INTO notification_queue (user_id, token_id, type, priority, payload, 
		                                scheduled_at, status, retry_count, max_retries, 
		                                created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now
	if item.ScheduledAt.IsZero() {
		item.ScheduledAt = now
	}
	if item.Status == "" {
		item.Status = domain.NotificationStatusPending
	}
	if item.MaxRetries == 0 {
		item.MaxRetries = 3
	}

	return r.db.QueryRow(
		context.Background(),
		query,
		item.UserID,
		item.TokenID,
		item.Type,
		item.Priority,
		item.Payload,
		item.ScheduledAt,
		item.Status,
		item.RetryCount,
		item.MaxRetries,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&item.ID)
}

func (r *notificationQueueRepository) Dequeue(limit int) ([]*domain.NotificationQueueItem, error) {
	query := `
		UPDATE notification_queue 
		SET status = 'processing', updated_at = NOW()
		WHERE id IN (
			SELECT id FROM notification_queue
			WHERE status = 'pending' 
			AND scheduled_at <= NOW()
			ORDER BY priority ASC, created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, notification_id, user_id, token_id, type, priority, payload, 
		          scheduled_at, sent_at, delivered_at, failed_at, error_message, 
		          retry_count, max_retries, status, created_at, updated_at
	`

	rows, err := r.db.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanItems(rows)
}

func (r *notificationQueueRepository) MarkSent(id uuid.UUID) error {
	query := `
		UPDATE notification_queue 
		SET status = 'sent', sent_at = NOW(), updated_at = NOW(), error_message = ''
		WHERE id = $1
	`

	_, err := r.db.Exec(context.Background(), query, id)
	return err
}

func (r *notificationQueueRepository) MarkFailed(id uuid.UUID, errorMessage string) error {
	query := `
		UPDATE notification_queue 
		SET status = CASE 
				WHEN retry_count >= max_retries THEN 'failed' 
				ELSE 'pending' 
			END,
			failed_at = NOW(),
			error_message = $2,
			retry_count = retry_count + 1,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(context.Background(), query, id, errorMessage)
	return err
}

func (r *notificationQueueRepository) GetFailed() ([]*domain.NotificationQueueItem, error) {
	query := `
		SELECT id, notification_id, user_id, token_id, type, priority, payload, 
		       scheduled_at, sent_at, delivered_at, failed_at, error_message, 
		       retry_count, max_retries, status, created_at, updated_at
		FROM notification_queue
		WHERE status = 'failed' AND retry_count < max_retries
		ORDER BY priority ASC, created_at ASC
	`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanItems(rows)
}

func (r *notificationQueueRepository) GetPendingCount() (int, error) {
	query := `
		SELECT COUNT(*) FROM notification_queue WHERE status = 'pending'
	`

	var count int
	err := r.db.QueryRow(context.Background(), query).Scan(&count)
	return count, err
}

func (r *notificationQueueRepository) UpdateRetry(id uuid.UUID, errorMessage string, scheduledAt time.Time) error {
	query := `
		UPDATE notification_queue 
		SET status = 'pending',
		    error_message = $2,
		    retry_count = retry_count + 1,
		    scheduled_at = $3,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(context.Background(), query, id, errorMessage, scheduledAt)
	return err
}

func (r *notificationQueueRepository) scanItems(rows pgx.Rows) ([]*domain.NotificationQueueItem, error) {
	var items []*domain.NotificationQueueItem

	for rows.Next() {
		item := &domain.NotificationQueueItem{}
		err := rows.Scan(
			&item.ID,
			&item.NotificationID,
			&item.UserID,
			&item.TokenID,
			&item.Type,
			&item.Priority,
			&item.Payload,
			&item.ScheduledAt,
			&item.SentAt,
			&item.DeliveredAt,
			&item.FailedAt,
			&item.ErrorMessage,
			&item.RetryCount,
			&item.MaxRetries,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
