package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSharedGoalNotFound      = errors.New("shared goal not found")
	ErrInvalidTargetType       = errors.New("invalid target type")
	ErrInvalidTargetValue      = errors.New("invalid target value")
	ErrGoalAlreadyCompleted    = errors.New("goal already completed")
	ErrGoalCancelled           = errors.New("goal is cancelled")
	ErrNotConnectionParticipant = errors.New("user is not a participant in this connection")
)

// SharedGoalTargetType defines the possible target types for shared goals
type SharedGoalTargetType string

const (
	SharedGoalTargetTypeTodosCompleted SharedGoalTargetType = "todos_completed"
	SharedGoalTargetTypeStreakDays     SharedGoalTargetType = "streak_days"
)

// SharedGoalStatus defines the possible statuses for shared goals
type SharedGoalStatus string

const (
	SharedGoalStatusActive     SharedGoalStatus = "active"
	SharedGoalStatusCompleted  SharedGoalStatus = "completed"
	SharedGoalStatusCancelled  SharedGoalStatus = "cancelled"
)

// SharedGoal represents a collaborative goal between two connected users
type SharedGoal struct {
	ID                uuid.UUID            `json:"id"`
	ConnectionID      uuid.UUID            `json:"connection_id"`
	TargetType        SharedGoalTargetType `json:"target_type"`
	TargetValue       int                  `json:"target_value"`
	CurrentValue      int                  `json:"current_value"`
	RewardDescription string               `json:"reward_description"`
	Status            SharedGoalStatus     `json:"status"`
	CreatedAt         time.Time            `json:"created_at"`
	CompletedAt       *time.Time           `json:"completed_at,omitempty"`
}

func (g *SharedGoal) IsActive() bool {
	return g.Status == SharedGoalStatusActive
}

func (g *SharedGoal) IsCompleted() bool {
	return g.Status == SharedGoalStatusCompleted
}

func (g *SharedGoal) IsCancelled() bool {
	return g.Status == SharedGoalStatusCancelled
}

// GetProgressPercentage returns the progress as a percentage (0-100)
func (g *SharedGoal) GetProgressPercentage() float64 {
	if g.TargetValue <= 0 {
		return 0.0
	}
	progress := float64(g.CurrentValue) / float64(g.TargetValue) * 100.0
	if progress > 100.0 {
		return 100.0
	}
	return progress
}

// CanBeUpdated checks if the goal can be updated with progress
func (g *SharedGoal) CanBeUpdated() error {
	if g.IsCompleted() {
		return ErrGoalAlreadyCompleted
	}
	if g.IsCancelled() {
		return ErrGoalCancelled
	}
	return nil
}

// MarkAsCompleted marks the goal as completed
func (g *SharedGoal) MarkAsCompleted() {
	now := time.Now()
	g.Status = SharedGoalStatusCompleted
	g.CompletedAt = &now
	if g.CurrentValue < g.TargetValue {
		g.CurrentValue = g.TargetValue
	}
}

func ValidateTargetType(targetType SharedGoalTargetType) error {
	switch targetType {
	case SharedGoalTargetTypeTodosCompleted, SharedGoalTargetTypeStreakDays:
		return nil
	default:
		return ErrInvalidTargetType
	}
}

func ValidateTargetValue(targetValue int) error {
	if targetValue <= 0 {
		return ErrInvalidTargetValue
	}
	return nil
}

// SharedGoalWithConnection extends SharedGoal with connection details
type SharedGoalWithConnection struct {
	SharedGoal
	Connection *Connection `json:"connection"`
}

// SharedGoalRepository defines the interface for shared goal persistence
type SharedGoalRepository interface {
	Create(goal *SharedGoal) error
	GetByID(id uuid.UUID) (*SharedGoal, error)
	GetByConnection(connectionID uuid.UUID) ([]*SharedGoal, error)
	GetByUserID(userID uuid.UUID) ([]*SharedGoal, error)
	UpdateProgress(id uuid.UUID, amount int) error
	Update(goal *SharedGoal) error
}

// SharedGoalService defines the interface for shared goal business logic
type SharedGoalService interface {
	CreateGoal(connectionID uuid.UUID, targetType SharedGoalTargetType, targetValue int, rewardDescription string) (*SharedGoal, error)
	UpdateProgress(connectionID uuid.UUID, amount int) error
	CheckCompletion(goalID uuid.UUID) (*SharedGoal, error)
	ListGoals(userID uuid.UUID) ([]*SharedGoal, error)
	OnTodoCompleted(userID uuid.UUID)
}

// CreateSharedGoalRequest represents a request to create a shared goal
type CreateSharedGoalRequest struct {
	ConnectionID      string `json:"connection_id" validate:"required,uuid"`
	TargetType        string `json:"target_type" validate:"required,oneof=todos_completed streak_days"`
	TargetValue       int    `json:"target_value" validate:"required,min=1"`
	RewardDescription string `json:"reward_description" validate:"required,max=255"`
}

// SharedGoalResponse represents a shared goal in API responses
type SharedGoalResponse struct {
	ID                string    `json:"id"`
	ConnectionID      string    `json:"connection_id"`
	TargetType        string    `json:"target_type"`
	TargetValue       int       `json:"target_value"`
	CurrentValue      int       `json:"current_value"`
	ProgressPercent   float64   `json:"progress_percent"`
	RewardDescription string    `json:"reward_description"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

func (g *SharedGoal) ToResponse() SharedGoalResponse {
	return SharedGoalResponse{
		ID:                g.ID.String(),
		ConnectionID:      g.ConnectionID.String(),
		TargetType:        string(g.TargetType),
		TargetValue:       g.TargetValue,
		CurrentValue:      g.CurrentValue,
		ProgressPercent:   g.GetProgressPercentage(),
		RewardDescription: g.RewardDescription,
		Status:            string(g.Status),
		CreatedAt:         g.CreatedAt,
		CompletedAt:       g.CompletedAt,
	}
}

// ListSharedGoalsResponse represents the response for listing shared goals
type ListSharedGoalsResponse struct {
	Goals []SharedGoalResponse `json:"goals"`
}

const SharedGoalCompletionBonusXP = 100
