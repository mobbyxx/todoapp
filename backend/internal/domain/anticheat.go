package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Anti-cheat related errors
var (
	ErrRateLimitExceeded      = errors.New("rate limit exceeded")
	ErrInvalidTimestamp       = errors.New("invalid timestamp")
	ErrDuplicateAction        = errors.New("duplicate action detected")
	ErrTodoAlreadyCompleted   = errors.New("todo already completed")
	ErrActionTooFast          = errors.New("action performed too quickly")
	ErrTimestampBackdated     = errors.New("timestamp cannot be backdated")
	ErrStatusCycleDetected    = errors.New("rapid status cycling detected")
	ErrSelfAssignmentDetected = errors.New("cannot complete own assigned todo")
)

// ActionType defines the type of action for rate limiting
type ActionType string

const (
	ActionTypeTodoComplete   ActionType = "todo_complete"
	ActionTypeTodoUncomplete ActionType = "todo_uncomplete"
	ActionTypeTodoCreate     ActionType = "todo_create"
	ActionTypeTodoUpdate     ActionType = "todo_update"
)

// AntiCheatService defines the interface for anti-cheat validation
type AntiCheatService interface {
	// ValidateTodoComplete performs all validation checks for todo completion
	ValidateTodoComplete(userID uuid.UUID, todoID uuid.UUID, clientTimestamp time.Time) error

	// CheckRateLimit checks if the user has exceeded rate limits for an action type
	CheckRateLimit(userID uuid.UUID, actionType ActionType) error

	// CheckTimestamp validates client timestamp is within acceptable range
	CheckTimestamp(clientTimestamp time.Time) error

	// CheckIdempotency prevents duplicate XP awards for the same action
	CheckIdempotency(userID uuid.UUID, todoID uuid.UUID) error

	// RecordAction records an action timestamp for minimum gap enforcement
	RecordAction(userID uuid.UUID, actionType ActionType) error

	// CheckMinTimeGap checks if enough time has passed since last action
	CheckMinTimeGap(userID uuid.UUID, actionType ActionType) error

	// CheckStatusCycle detects rapid complete/uncomplete cycling
	CheckStatusCycle(userID uuid.UUID, todoID uuid.UUID, newStatus TodoStatus) error
}

// AntiCheatConfig holds configuration for anti-cheat validation
type AntiCheatConfig struct {
	// RateLimitMaxActions is the maximum number of actions allowed per window
	RateLimitMaxActions int

	// RateLimitWindow is the duration for rate limiting
	RateLimitWindow time.Duration

	// TimestampTolerance is the allowed clock skew between client and server
	TimestampTolerance time.Duration

	// IdempotencyTTL is how long to track idempotency keys
	IdempotencyTTL time.Duration

	// MinActionGap is the minimum time between actions
	MinActionGap time.Duration

	// StatusCycleWindow is the window to check for rapid status cycling
	StatusCycleWindow time.Duration

	// StatusCycleThreshold is the number of status changes to trigger detection
	StatusCycleThreshold int
}

// DefaultAntiCheatConfig returns the default anti-cheat configuration
func DefaultAntiCheatConfig() AntiCheatConfig {
	return AntiCheatConfig{
		RateLimitMaxActions:  10,
		RateLimitWindow:      time.Minute,
		TimestampTolerance:   5 * time.Minute,
		IdempotencyTTL:       24 * time.Hour,
		MinActionGap:         5 * time.Second,
		StatusCycleWindow:    time.Minute,
		StatusCycleThreshold: 5,
	}
}

// SuspiciousActivity represents a logged suspicious action for monitoring
type SuspiciousActivity struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TodoID        *uuid.UUID
	ActivityType  string
	Reason        string
	ClientTime    *time.Time
	ServerTime    time.Time
	Details       map[string]interface{}
}
