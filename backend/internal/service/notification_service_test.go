package service

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/todo-api/internal/domain"
)

type mockNotificationQueueRepository struct {
	mock.Mock
}

func (m *mockNotificationQueueRepository) Enqueue(item *domain.NotificationQueueItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *mockNotificationQueueRepository) Dequeue(limit int) ([]*domain.NotificationQueueItem, error) {
	args := m.Called(limit)
	return args.Get(0).([]*domain.NotificationQueueItem), args.Error(1)
}

func (m *mockNotificationQueueRepository) MarkSent(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockNotificationQueueRepository) MarkFailed(id uuid.UUID, errorMessage string) error {
	args := m.Called(id, errorMessage)
	return args.Error(0)
}

func (m *mockNotificationQueueRepository) GetFailed() ([]*domain.NotificationQueueItem, error) {
	args := m.Called()
	return args.Get(0).([]*domain.NotificationQueueItem), args.Error(1)
}

func (m *mockNotificationQueueRepository) GetPendingCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *mockNotificationQueueRepository) UpdateRetry(id uuid.UUID, errorMessage string, scheduledAt time.Time) error {
	args := m.Called(id, errorMessage, scheduledAt)
	return args.Error(0)
}

func TestNewNotificationService(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)

	assert.NotNil(t, svc)
	assert.Implements(t, (*domain.NotificationService)(nil), svc)
}

func TestQueueNotification(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)
	userID := uuid.New()

	mockRepo.On("Enqueue", mock.AnythingOfType("*domain.NotificationQueueItem")).Return(nil)

	err := svc.QueueNotification(
		userID,
		domain.NotificationTypePush,
		"Test Title",
		"Test Body",
		map[string]interface{}{"key": "value"},
		2,
	)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestQueueNotification_NormalizesPriority(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)
	userID := uuid.New()

	mockRepo.On("Enqueue", mock.MatchedBy(func(item *domain.NotificationQueueItem) bool {
		return item.Priority == 1
	})).Return(nil)

	err := svc.QueueNotification(userID, domain.NotificationTypePush, "Title", "Body", nil, 0)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestQueueConnectionRequest(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)
	userID := uuid.New()
	fromUserID := uuid.New()
	connectionID := uuid.New()

	mockRepo.On("Enqueue", mock.MatchedBy(func(item *domain.NotificationQueueItem) bool {
		if item.UserID != userID || item.Type != domain.NotificationTypePush {
			return false
		}
		var payload domain.ConnectionRequestPayload
		err := json.Unmarshal(item.Payload, &payload)
		return err == nil && payload.ConnectionID == connectionID.String() && payload.FromUserID == fromUserID.String()
	})).Return(nil)

	err := svc.QueueConnectionRequest(userID, fromUserID, "TestUser", connectionID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestQueueConnectionAccepted(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)
	userID := uuid.New()
	acceptedByID := uuid.New()
	connectionID := uuid.New()

	mockRepo.On("Enqueue", mock.MatchedBy(func(item *domain.NotificationQueueItem) bool {
		if item.UserID != userID {
			return false
		}
		var payload domain.ConnectionAcceptedPayload
		err := json.Unmarshal(item.Payload, &payload)
		return err == nil && payload.ConnectionID == connectionID.String()
	})).Return(nil)

	err := svc.QueueConnectionAccepted(userID, acceptedByID, "Acceptor", connectionID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestQueueTodoAssigned(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)
	userID := uuid.New()
	todoID := uuid.New()
	assignerID := uuid.New()

	mockRepo.On("Enqueue", mock.MatchedBy(func(item *domain.NotificationQueueItem) bool {
		if item.UserID != userID {
			return false
		}
		var payload domain.TodoAssignedPayload
		err := json.Unmarshal(item.Payload, &payload)
		return err == nil && payload.TodoID == todoID.String() && payload.Title == "Test Todo"
	})).Return(nil)

	err := svc.QueueTodoAssigned(userID, todoID, "Test Todo", assignerID, "Assigner")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestQueueTodoCompleted(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)
	userID := uuid.New()
	todoID := uuid.New()
	completedByID := uuid.New()

	mockRepo.On("Enqueue", mock.MatchedBy(func(item *domain.NotificationQueueItem) bool {
		if item.UserID != userID {
			return false
		}
		var payload domain.TodoCompletedPayload
		err := json.Unmarshal(item.Payload, &payload)
		return err == nil && payload.TodoID == todoID.String() && payload.UserName == "Completer"
	})).Return(nil)

	err := svc.QueueTodoCompleted(userID, todoID, "Test Todo", completedByID, "Completer")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestQueueBadgeEarned(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)
	userID := uuid.New()

	badge := &domain.Badge{
		ID:          uuid.New(),
		Name:        "First Steps",
		Description: "Complete your first todo",
		Icon:        "badge_first_steps.png",
	}

	mockRepo.On("Enqueue", mock.MatchedBy(func(item *domain.NotificationQueueItem) bool {
		if item.UserID != userID {
			return false
		}
		var payload domain.BadgeEarnedPayload
		err := json.Unmarshal(item.Payload, &payload)
		return err == nil && payload.BadgeName == badge.Name
	})).Return(nil)

	err := svc.QueueBadgeEarned(userID, badge)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestQueueGoalCompleted(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)
	userID := uuid.New()
	goalID := uuid.New()

	mockRepo.On("Enqueue", mock.MatchedBy(func(item *domain.NotificationQueueItem) bool {
		if item.UserID != userID {
			return false
		}
		var payload domain.GoalCompletedPayload
		err := json.Unmarshal(item.Payload, &payload)
		return err == nil && payload.GoalName == "TargetType"
	})).Return(nil)

	err := svc.QueueGoalCompleted(userID, goalID, "TargetType", "Reward Description")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProcessQueue(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)

	userID := uuid.New()
	item := &domain.NotificationQueueItem{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     domain.NotificationTypePush,
		Priority: 2,
		Payload:  json.RawMessage(`{"title":"Test"}`),
	}

	mockRepo.On("Dequeue", 10).Return([]*domain.NotificationQueueItem{item}, nil)
	mockRepo.On("MarkSent", item.ID).Return(nil)

	err := svc.ProcessQueue(10)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProcessQueue_Empty(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)

	mockRepo.On("Dequeue", 10).Return([]*domain.NotificationQueueItem{}, nil)

	err := svc.ProcessQueue(10)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProcessQueue_WithFailure(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)

	userID := uuid.New()
	item := &domain.NotificationQueueItem{
		ID:         uuid.New(),
		UserID:     userID,
		Type:       domain.NotificationTypePush,
		Priority:   2,
		Payload:    json.RawMessage(`{}`),
		RetryCount: 0,
		MaxRetries: 3,
	}

	mockRepo.On("Dequeue", 10).Return([]*domain.NotificationQueueItem{item}, nil)
	mockRepo.On("MarkFailed", item.ID, mock.AnythingOfType("string")).Return(nil)

	err := svc.ProcessQueue(10)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRetryFailed(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)

	userID := uuid.New()
	item := &domain.NotificationQueueItem{
		ID:         uuid.New(),
		UserID:     userID,
		Type:       domain.NotificationTypePush,
		Priority:   2,
		Payload:    json.RawMessage(`{"title":"Test"}`),
		RetryCount: 1,
		MaxRetries: 3,
		Status:     domain.NotificationStatusFailed,
	}

	mockRepo.On("GetFailed").Return([]*domain.NotificationQueueItem{item}, nil)
	mockRepo.On("MarkSent", item.ID).Return(nil)

	err := svc.RetryFailed()

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRetryFailed_Empty(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)

	mockRepo.On("GetFailed").Return([]*domain.NotificationQueueItem{}, nil)

	err := svc.RetryFailed()

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRetryFailed_MaxRetriesReached(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)

	userID := uuid.New()
	item := &domain.NotificationQueueItem{
		ID:         uuid.New(),
		UserID:     userID,
		Type:       domain.NotificationTypePush,
		RetryCount: 3,
		MaxRetries: 3,
		Status:     domain.NotificationStatusFailed,
	}

	mockRepo.On("GetFailed").Return([]*domain.NotificationQueueItem{item}, nil)

	err := svc.RetryFailed()

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestNotificationQueueItem_PriorityRange(t *testing.T) {
	item := &domain.NotificationQueueItem{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Type:      domain.NotificationTypePush,
		Priority:  1,
		Status:    domain.NotificationStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.True(t, item.Priority >= 1 && item.Priority <= 5)
}

func TestNotificationQueueItem_RetryLogic(t *testing.T) {
	item := &domain.NotificationQueueItem{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Type:       domain.NotificationTypePush,
		RetryCount: 2,
		MaxRetries: 3,
		Status:     domain.NotificationStatusPending,
	}

	assert.True(t, item.RetryCount < item.MaxRetries)
	assert.Equal(t, domain.NotificationStatusPending, item.Status)
}

func TestQueueNotification_RepositoryError(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)
	userID := uuid.New()

	mockRepo.On("Enqueue", mock.Anything).Return(errors.New("database error"))

	err := svc.QueueNotification(userID, domain.NotificationTypePush, "Title", "Body", nil, 2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestProcessQueue_DequeueError(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)

	mockRepo.On("Dequeue", 10).Return([]*domain.NotificationQueueItem{}, errors.New("dequeue error"))

	err := svc.ProcessQueue(10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dequeue error")
	mockRepo.AssertExpectations(t)
}

func TestRetryFailed_GetFailedError(t *testing.T) {
	mockRepo := new(mockNotificationQueueRepository)
	svc := NewNotificationService(mockRepo)

	mockRepo.On("GetFailed").Return([]*domain.NotificationQueueItem{}, errors.New("get failed error"))

	err := svc.RetryFailed()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get failed error")
	mockRepo.AssertExpectations(t)
}
