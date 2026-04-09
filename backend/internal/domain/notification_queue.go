package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypePush  NotificationType = "push"
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
)

type NotificationStatus string

const (
	NotificationStatusPending    NotificationStatus = "pending"
	NotificationStatusProcessing NotificationStatus = "processing"
	NotificationStatusSent       NotificationStatus = "sent"
	NotificationStatusDelivered  NotificationStatus = "delivered"
	NotificationStatusFailed     NotificationStatus = "failed"
	NotificationStatusCancelled  NotificationStatus = "cancelled"
)

type NotificationQueueItem struct {
	ID             uuid.UUID          `json:"id"`
	NotificationID *uuid.UUID         `json:"notification_id,omitempty"`
	UserID         uuid.UUID          `json:"user_id"`
	TokenID        *uuid.UUID         `json:"token_id,omitempty"`
	Type           NotificationType   `json:"type"`
	Priority       int                `json:"priority"`
	Payload        json.RawMessage    `json:"payload"`
	ScheduledAt    time.Time          `json:"scheduled_at"`
	SentAt         *time.Time         `json:"sent_at,omitempty"`
	DeliveredAt    *time.Time         `json:"delivered_at,omitempty"`
	FailedAt       *time.Time         `json:"failed_at,omitempty"`
	ErrorMessage   string             `json:"error_message,omitempty"`
	RetryCount     int                `json:"retry_count"`
	MaxRetries     int                `json:"max_retries"`
	Status         NotificationStatus `json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

type ConnectionRequestPayload struct {
	ConnectionID   string `json:"connection_id"`
	FromUserID     string `json:"from_user_id"`
	FromUserName   string `json:"from_user_name"`
	InvitationLink string `json:"invitation_link"`
}

type ConnectionAcceptedPayload struct {
	ConnectionID string `json:"connection_id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
}

type TodoAssignedPayload struct {
	TodoID    string `json:"todo_id"`
	Title     string `json:"title"`
	AssignerID string `json:"assigner_id"`
	AssignerName string `json:"assigner_name"`
}

type TodoCompletedPayload struct {
	TodoID     string `json:"todo_id"`
	Title      string `json:"title"`
	CompletedBy string `json:"completed_by"`
	UserName   string `json:"user_name"`
}

type BadgeEarnedPayload struct {
	BadgeID     string `json:"badge_id"`
	BadgeName   string `json:"badge_name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type GoalCompletedPayload struct {
	GoalID   string `json:"goal_id"`
	GoalName string `json:"goal_name"`
	Reward   string `json:"reward,omitempty"`
}

type NotificationQueueRepository interface {
	Enqueue(item *NotificationQueueItem) error
	Dequeue(limit int) ([]*NotificationQueueItem, error)
	MarkSent(id uuid.UUID) error
	MarkFailed(id uuid.UUID, errorMessage string) error
	GetFailed() ([]*NotificationQueueItem, error)
	GetPendingCount() (int, error)
	UpdateRetry(id uuid.UUID, errorMessage string, scheduledAt time.Time) error
}

type NotificationService interface {
	QueueNotification(userID uuid.UUID, notificationType NotificationType, title string, body string, data map[string]interface{}, priority int) error
	QueueConnectionRequest(userID uuid.UUID, fromUserID uuid.UUID, fromUserName string, connectionID uuid.UUID) error
	QueueConnectionAccepted(userID uuid.UUID, acceptedByID uuid.UUID, acceptedByName string, connectionID uuid.UUID) error
	QueueTodoAssigned(userID uuid.UUID, todoID uuid.UUID, title string, assignerID uuid.UUID, assignerName string) error
	QueueTodoCompleted(userID uuid.UUID, todoID uuid.UUID, title string, completedByID uuid.UUID, completedByName string) error
	QueueBadgeEarned(userID uuid.UUID, badge *Badge) error
	QueueGoalCompleted(userID uuid.UUID, goalID uuid.UUID, goalName string, reward string) error
	ProcessQueue(batchSize int) error
	RetryFailed() error
}
