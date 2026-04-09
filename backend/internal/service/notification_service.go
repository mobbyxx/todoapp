package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/user/todo-api/internal/domain"
)

type notificationService struct {
	repo domain.NotificationQueueRepository
}

func NewNotificationService(repo domain.NotificationQueueRepository) domain.NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) QueueNotification(
	userID uuid.UUID,
	notificationType domain.NotificationType,
	title string,
	body string,
	data map[string]interface{},
	priority int,
) error {
	payload := map[string]interface{}{
		"title": title,
		"body":  body,
		"data":  data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}

	if priority < 1 {
		priority = 1
	}
	if priority > 5 {
		priority = 5
	}

	item := &domain.NotificationQueueItem{
		UserID:     userID,
		Type:       notificationType,
		Priority:   priority,
		Payload:    payloadBytes,
		Status:     domain.NotificationStatusPending,
		ScheduledAt: time.Now(),
	}

	return s.repo.Enqueue(item)
}

func (s *notificationService) QueueConnectionRequest(
	userID uuid.UUID,
	fromUserID uuid.UUID,
	fromUserName string,
	connectionID uuid.UUID,
) error {
	payload := domain.ConnectionRequestPayload{
		ConnectionID:   connectionID.String(),
		FromUserID:     fromUserID.String(),
		FromUserName:   fromUserName,
		InvitationLink: fmt.Sprintf("/connections/accept/%s", connectionID.String()),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal connection request payload: %w", err)
	}

	item := &domain.NotificationQueueItem{
		UserID:     userID,
		Type:       domain.NotificationTypePush,
		Priority:   2,
		Payload:    payloadBytes,
		Status:     domain.NotificationStatusPending,
		ScheduledAt: time.Now(),
	}

	return s.repo.Enqueue(item)
}

func (s *notificationService) QueueConnectionAccepted(
	userID uuid.UUID,
	acceptedByID uuid.UUID,
	acceptedByName string,
	connectionID uuid.UUID,
) error {
	payload := domain.ConnectionAcceptedPayload{
		ConnectionID: connectionID.String(),
		UserID:       acceptedByID.String(),
		UserName:     acceptedByName,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal connection accepted payload: %w", err)
	}

	item := &domain.NotificationQueueItem{
		UserID:     userID,
		Type:       domain.NotificationTypePush,
		Priority:   3,
		Payload:    payloadBytes,
		Status:     domain.NotificationStatusPending,
		ScheduledAt: time.Now(),
	}

	return s.repo.Enqueue(item)
}

func (s *notificationService) QueueTodoAssigned(
	userID uuid.UUID,
	todoID uuid.UUID,
	title string,
	assignerID uuid.UUID,
	assignerName string,
) error {
	payload := domain.TodoAssignedPayload{
		TodoID:       todoID.String(),
		Title:        title,
		AssignerID:   assignerID.String(),
		AssignerName: assignerName,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal todo assigned payload: %w", err)
	}

	item := &domain.NotificationQueueItem{
		UserID:     userID,
		Type:       domain.NotificationTypePush,
		Priority:   2,
		Payload:    payloadBytes,
		Status:     domain.NotificationStatusPending,
		ScheduledAt: time.Now(),
	}

	return s.repo.Enqueue(item)
}

func (s *notificationService) QueueTodoCompleted(
	userID uuid.UUID,
	todoID uuid.UUID,
	title string,
	completedByID uuid.UUID,
	completedByName string,
) error {
	payload := domain.TodoCompletedPayload{
		TodoID:      todoID.String(),
		Title:       title,
		CompletedBy: completedByID.String(),
		UserName:    completedByName,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal todo completed payload: %w", err)
	}

	item := &domain.NotificationQueueItem{
		UserID:     userID,
		Type:       domain.NotificationTypePush,
		Priority:   3,
		Payload:    payloadBytes,
		Status:     domain.NotificationStatusPending,
		ScheduledAt: time.Now(),
	}

	return s.repo.Enqueue(item)
}

func (s *notificationService) QueueBadgeEarned(
	userID uuid.UUID,
	badge *domain.Badge,
) error {
	payload := domain.BadgeEarnedPayload{
		BadgeID:     badge.ID.String(),
		BadgeName:   badge.Name,
		Description: badge.Description,
		Icon:        badge.Icon,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal badge earned payload: %w", err)
	}

	item := &domain.NotificationQueueItem{
		UserID:     userID,
		Type:       domain.NotificationTypePush,
		Priority:   2,
		Payload:    payloadBytes,
		Status:     domain.NotificationStatusPending,
		ScheduledAt: time.Now(),
	}

	return s.repo.Enqueue(item)
}

func (s *notificationService) QueueGoalCompleted(
	userID uuid.UUID,
	goalID uuid.UUID,
	goalName string,
	reward string,
) error {
	payload := domain.GoalCompletedPayload{
		GoalID:   goalID.String(),
		GoalName: goalName,
		Reward:   reward,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal goal completed payload: %w", err)
	}

	item := &domain.NotificationQueueItem{
		UserID:     userID,
		Type:       domain.NotificationTypePush,
		Priority:   3,
		Payload:    payloadBytes,
		Status:     domain.NotificationStatusPending,
		ScheduledAt: time.Now(),
	}

	return s.repo.Enqueue(item)
}

func (s *notificationService) ProcessQueue(batchSize int) error {
	if batchSize <= 0 {
		batchSize = 10
	}

	items, err := s.repo.Dequeue(batchSize)
	if err != nil {
		return fmt.Errorf("failed to dequeue notifications: %w", err)
	}

	if len(items) == 0 {
		return nil
	}

	log.Info().Int("count", len(items)).Msg("processing notification queue batch")

	for _, item := range items {
		if err := s.processItem(item); err != nil {
			log.Warn().
				Err(err).
				Str("notification_id", item.ID.String()).
				Int("retry_count", item.RetryCount).
				Msg("failed to process notification")

			if markErr := s.repo.MarkFailed(item.ID, err.Error()); markErr != nil {
				log.Error().
					Err(markErr).
					Str("notification_id", item.ID.String()).
					Msg("failed to mark notification as failed")
			}
		} else {
			if markErr := s.repo.MarkSent(item.ID); markErr != nil {
				log.Error().
					Err(markErr).
					Str("notification_id", item.ID.String()).
					Msg("failed to mark notification as sent")
			}
		}
	}

	return nil
}

func (s *notificationService) RetryFailed() error {
	items, err := s.repo.GetFailed()
	if err != nil {
		return fmt.Errorf("failed to get failed notifications: %w", err)
	}

	if len(items) == 0 {
		return nil
	}

	log.Info().Int("count", len(items)).Msg("retrying failed notifications")

	for _, item := range items {
		if item.RetryCount >= item.MaxRetries {
			continue
		}

		item.Status = domain.NotificationStatusPending
		item.ErrorMessage = ""
		item.UpdatedAt = time.Now()

		if err := s.processItem(item); err != nil {
			log.Warn().
				Err(err).
				Str("notification_id", item.ID.String()).
				Int("retry_count", item.RetryCount).
				Msg("retry failed for notification")

			if markErr := s.repo.MarkFailed(item.ID, err.Error()); markErr != nil {
				log.Error().
					Err(markErr).
					Str("notification_id", item.ID.String()).
					Msg("failed to mark notification as failed after retry")
			}
		} else {
			if markErr := s.repo.MarkSent(item.ID); markErr != nil {
				log.Error().
					Err(markErr).
					Str("notification_id", item.ID.String()).
					Msg("failed to mark notification as sent after retry")
			}
		}
	}

	return nil
}

func (s *notificationService) processItem(item *domain.NotificationQueueItem) error {
	log.Info().
		Str("notification_id", item.ID.String()).
		Str("user_id", item.UserID.String()).
		Str("type", string(item.Type)).
		Int("priority", item.Priority).
		Msg("processing notification")

	return nil
}
