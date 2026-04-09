package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/user/todo-api/internal/domain"
)

const (
	rateLimitPrefix       = "ratelimit:"
	idempotencyPrefix     = "idempotency:"
	actionTimestampPrefix = "action:"
	statusCyclePrefix     = "statuscycle:"
)

type antiCheatService struct {
	redis      *redis.Client
	todoRepo   domain.TodoRepository
	config     domain.AntiCheatConfig
}

func NewAntiCheatService(redisClient *redis.Client, todoRepo domain.TodoRepository, config domain.AntiCheatConfig) domain.AntiCheatService {
	return &antiCheatService{
		redis:    redisClient,
		todoRepo: todoRepo,
		config:   config,
	}
}

func (s *antiCheatService) ValidateTodoComplete(userID uuid.UUID, todoID uuid.UUID, clientTimestamp time.Time) error {
	ctx := context.Background()

	if err := s.CheckRateLimit(userID, domain.ActionTypeTodoComplete); err != nil {
		s.logSuspiciousActivity(ctx, userID, &todoID, "rate_limit_violation", err.Error(), &clientTimestamp)
		return err
	}

	if err := s.CheckTimestamp(clientTimestamp); err != nil {
		s.logSuspiciousActivity(ctx, userID, &todoID, "invalid_timestamp", err.Error(), &clientTimestamp)
		return err
	}

	if err := s.CheckIdempotency(userID, todoID); err != nil {
		s.logSuspiciousActivity(ctx, userID, &todoID, "duplicate_action", err.Error(), &clientTimestamp)
		return err
	}

	if err := s.CheckMinTimeGap(userID, domain.ActionTypeTodoComplete); err != nil {
		s.logSuspiciousActivity(ctx, userID, &todoID, "action_too_fast", err.Error(), &clientTimestamp)
		return err
	}

	if err := s.CheckStatusCycle(userID, todoID, domain.TodoStatusCompleted); err != nil {
		s.logSuspiciousActivity(ctx, userID, &todoID, "status_cycle", err.Error(), &clientTimestamp)
		return err
	}

	todo, err := s.todoRepo.GetByID(todoID)
	if err != nil {
		return err
	}

	if todo.Status == domain.TodoStatusCompleted {
		s.logSuspiciousActivity(ctx, userID, &todoID, "already_completed", "todo already in completed status", &clientTimestamp)
		return domain.ErrTodoAlreadyCompleted
	}

	if todo.AssignedTo != nil && *todo.AssignedTo == userID && todo.CreatedBy == userID {
		s.logSuspiciousActivity(ctx, userID, &todoID, "self_assignment", "user attempting to complete own assigned todo", &clientTimestamp)
		return domain.ErrSelfAssignmentDetected
	}

	if !clientTimestamp.IsZero() && clientTimestamp.Before(todo.CreatedAt) {
		s.logSuspiciousActivity(ctx, userID, &todoID, "backdated_completion", "completion timestamp before creation", &clientTimestamp)
		return domain.ErrTimestampBackdated
	}

	if err := s.RecordAction(userID, domain.ActionTypeTodoComplete); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to record action timestamp")
	}

	return nil
}

func (s *antiCheatService) CheckRateLimit(userID uuid.UUID, actionType domain.ActionType) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s%s:%s", rateLimitPrefix, userID.String(), actionType)

	pipe := s.redis.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, s.config.RateLimitWindow)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	count := incrCmd.Val()
	if count > int64(s.config.RateLimitMaxActions) {
		return domain.ErrRateLimitExceeded
	}

	return nil
}

func (s *antiCheatService) CheckTimestamp(clientTimestamp time.Time) error {
	if clientTimestamp.IsZero() {
		return nil
	}

	serverTime := time.Now()
	diff := clientTimestamp.Sub(serverTime)

	if diff < 0 {
		diff = -diff
	}

	if diff > s.config.TimestampTolerance {
		return domain.ErrInvalidTimestamp
	}

	return nil
}

func (s *antiCheatService) CheckIdempotency(userID uuid.UUID, todoID uuid.UUID) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s%s:%s", idempotencyPrefix, userID.String(), todoID.String())

	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check idempotency: %w", err)
	}

	if exists > 0 {
		return domain.ErrDuplicateAction
	}

	err = s.redis.Set(ctx, key, "1", s.config.IdempotencyTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set idempotency key: %w", err)
	}

	return nil
}

func (s *antiCheatService) RecordAction(userID uuid.UUID, actionType domain.ActionType) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s%s:%s", actionTimestampPrefix, userID.String(), actionType)

	now := time.Now().Unix()
	err := s.redis.Set(ctx, key, now, s.config.MinActionGap*2).Err()
	if err != nil {
		return fmt.Errorf("failed to record action: %w", err)
	}

	return nil
}

func (s *antiCheatService) CheckMinTimeGap(userID uuid.UUID, actionType domain.ActionType) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s%s:%s", actionTimestampPrefix, userID.String(), actionType)

	lastActionStr, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to check action gap: %w", err)
	}

	var lastAction int64
	if _, err := fmt.Sscanf(lastActionStr, "%d", &lastAction); err != nil {
		return nil
	}

	elapsed := time.Since(time.Unix(lastAction, 0))
	if elapsed < s.config.MinActionGap {
		return domain.ErrActionTooFast
	}

	return nil
}

func (s *antiCheatService) CheckStatusCycle(userID uuid.UUID, todoID uuid.UUID, newStatus domain.TodoStatus) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s%s:%s", statusCyclePrefix, userID.String(), todoID.String())

	statusChange := fmt.Sprintf("%d:%s", time.Now().Unix(), newStatus)

	pipe := s.redis.Pipeline()
	pipe.LPush(ctx, key, statusChange)
	pipe.LTrim(ctx, key, 0, int64(s.config.StatusCycleThreshold-1))
	pipe.Expire(ctx, key, s.config.StatusCycleWindow)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to check status cycle: %w", err)
	}

	changes, err := s.redis.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get status changes: %w", err)
	}

	if len(changes) >= s.config.StatusCycleThreshold {
		var lastStatus domain.TodoStatus
		alternations := 0

		for i := len(changes) - 1; i >= 0; i-- {
			var timestamp int64
			var status domain.TodoStatus
			if _, err := fmt.Sscanf(changes[i], "%d:%s", &timestamp, &status); err != nil {
				continue
			}

			if i < len(changes)-1 && status != lastStatus {
				alternations++
			}
			lastStatus = status
		}

		if alternations >= s.config.StatusCycleThreshold-1 {
			return domain.ErrStatusCycleDetected
		}
	}

	return nil
}

func (s *antiCheatService) logSuspiciousActivity(ctx context.Context, userID uuid.UUID, todoID *uuid.UUID, activityType, reason string, clientTime *time.Time) {
	log.Warn().
		Str("user_id", userID.String()).
		Str("activity_type", activityType).
		Str("reason", reason).
		Time("server_time", time.Now()).
		Func(func(e *zerolog.Event) {
			if todoID != nil {
				e.Str("todo_id", todoID.String())
			}
			if clientTime != nil && !clientTime.IsZero() {
				e.Time("client_time", *clientTime)
			}
		}).
		Msg("suspicious activity detected")
}
