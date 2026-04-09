package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/todo-api/internal/domain"
)

type mockGamificationRepository struct {
	mock.Mock
	users       map[uuid.UUID]*domain.User
	badges      map[uuid.UUID]*domain.Badge
	userBadges  map[string]*domain.UserBadge
	transactions []*domain.PointsTransaction
}

func newMockGamificationRepository() *mockGamificationRepository {
	return &mockGamificationRepository{
		users:        make(map[uuid.UUID]*domain.User),
		badges:       make(map[uuid.UUID]*domain.Badge),
		userBadges:   make(map[string]*domain.UserBadge),
		transactions: make([]*domain.PointsTransaction, 0),
	}
}

func (m *mockGamificationRepository) GetAllBadges() ([]*domain.Badge, error) {
	args := m.Called()
	return args.Get(0).([]*domain.Badge), args.Error(1)
}

func (m *mockGamificationRepository) GetBadgeByID(id uuid.UUID) (*domain.Badge, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Badge), args.Error(1)
}

func (m *mockGamificationRepository) GetUserBadges(userID uuid.UUID) ([]*domain.UserBadge, error) {
	args := m.Called(userID)
	return args.Get(0).([]*domain.UserBadge), args.Error(1)
}

func (m *mockGamificationRepository) HasBadge(userID, badgeID uuid.UUID) (bool, error) {
	args := m.Called(userID, badgeID)
	return args.Bool(0), args.Error(1)
}

func (m *mockGamificationRepository) AwardBadge(userBadge *domain.UserBadge) error {
	args := m.Called(userBadge)
	return args.Error(0)
}

func (m *mockGamificationRepository) GetUserStats(userID uuid.UUID) (*domain.UserStats, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserStats), args.Error(1)
}

func (m *mockGamificationRepository) CountCompletedTodos(userID uuid.UUID) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *mockGamificationRepository) CountEarlyCompletions(userID uuid.UUID, beforeHour int) (int, error) {
	args := m.Called(userID, beforeHour)
	return args.Int(0), args.Error(1)
}

func (m *mockGamificationRepository) CountConnections(userID uuid.UUID) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *mockGamificationRepository) AddXP(userID uuid.UUID, amount int, reason string, referenceType string, referenceID uuid.UUID) error {
	args := m.Called(userID, amount, reason, referenceType, referenceID)
	if user, ok := m.users[userID]; ok {
		user.TotalPoints += amount
	}
	return args.Error(0)
}

func (m *mockGamificationRepository) GetUserXP(userID uuid.UUID) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *mockGamificationRepository) GetPointsHistory(userID uuid.UUID, limit int) ([]*domain.PointsTransaction, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]*domain.PointsTransaction), args.Error(1)
}

func (m *mockGamificationRepository) UpdateStreak(userID uuid.UUID, newStreak int, lastStreakDate time.Time) error {
	args := m.Called(userID, newStreak, lastStreakDate)
	if user, ok := m.users[userID]; ok {
		user.StreakCount = newStreak
		user.LastStreakDate = &lastStreakDate
	}
	return args.Error(0)
}

func (m *mockGamificationRepository) GetStreakInfo(userID uuid.UUID) (int, *time.Time, int, error) {
	args := m.Called(userID)
	var lastDate *time.Time
	if args.Get(1) != nil {
		t := args.Get(1).(time.Time)
		lastDate = &t
	}
	return args.Int(0), lastDate, args.Int(2), args.Error(3)
}

func (m *mockGamificationRepository) UseFreezeToken(userID uuid.UUID) error {
	args := m.Called(userID)
	if user, ok := m.users[userID]; ok && user.StreakFreezeTokens > 0 {
		user.StreakFreezeTokens--
	}
	return args.Error(0)
}

func (m *mockGamificationRepository) AwardFreezeToken(userID uuid.UUID) error {
	args := m.Called(userID)
	if user, ok := m.users[userID]; ok {
		user.StreakFreezeTokens++
	}
	return args.Error(0)
}

type mockAntiCheatService struct {
	mock.Mock
}

func (m *mockAntiCheatService) ValidateTodoComplete(userID uuid.UUID, todoID uuid.UUID, clientTimestamp time.Time) error {
	args := m.Called(userID, todoID, clientTimestamp)
	return args.Error(0)
}

func (m *mockAntiCheatService) CheckRateLimit(userID uuid.UUID, actionType domain.ActionType) error {
	args := m.Called(userID, actionType)
	return args.Error(0)
}

func (m *mockAntiCheatService) CheckTimestamp(clientTimestamp time.Time) error {
	args := m.Called(clientTimestamp)
	return args.Error(0)
}

func (m *mockAntiCheatService) CheckIdempotency(userID uuid.UUID, todoID uuid.UUID) error {
	args := m.Called(userID, todoID)
	return args.Error(0)
}

func (m *mockAntiCheatService) RecordAction(userID uuid.UUID, actionType domain.ActionType) error {
	args := m.Called(userID, actionType)
	return args.Error(0)
}

func (m *mockAntiCheatService) CheckMinTimeGap(userID uuid.UUID, actionType domain.ActionType) error {
	args := m.Called(userID, actionType)
	return args.Error(0)
}

func (m *mockAntiCheatService) CheckStatusCycle(userID uuid.UUID, todoID uuid.UUID, newStatus domain.TodoStatus) error {
	args := m.Called(userID, todoID, newStatus)
	return args.Error(0)
}

func setupGamificationService() (*gamificationService, *mockGamificationRepository, *mockAntiCheatService) {
	mockRepo := newMockGamificationRepository()
	mockAntiCheat := &mockAntiCheatService{}
	
	svc := &gamificationService{
		repo:         mockRepo,
		antiCheatSvc: mockAntiCheat,
		redis:        &redis.Client{},
	}
	
	return svc, mockRepo, mockAntiCheat
}

func TestAwardXP_Success(t *testing.T) {
	svc, mockRepo, _ := setupGamificationService()
	userID := uuid.New()

	mockRepo.On("AddXP", userID, 10, "todo completed", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("GetUserXP", userID).Return(10, nil)

	err := svc.AwardXP(userID, 10, "todo completed")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAwardXP_InvalidAmount(t *testing.T) {
	svc, _, _ := setupGamificationService()
	userID := uuid.New()

	err := svc.AwardXP(userID, 0, "test")
	assert.ErrorIs(t, err, domain.ErrInvalidPointsAmount)

	err = svc.AwardXP(userID, -5, "test")
	assert.ErrorIs(t, err, domain.ErrInvalidPointsAmount)
}

func TestCalculateLevel(t *testing.T) {
	svc, _, _ := setupGamificationService()

	tests := []struct {
		xp        int
		expectedLevel int
	}{
		{0, 1},
		{50, 1},
		{100, 2},
		{299, 2},
		{300, 3},
		{699, 3},
		{700, 4},
		{1499, 4},
		{1500, 5},
		{2999, 5},
		{3000, 6},
		{5999, 6},
		{6000, 7},
		{11999, 7},
		{12000, 8},
		{20000, 8},
	}

	for _, tt := range tests {
		level := svc.CalculateLevel(tt.xp)
		assert.Equal(t, tt.expectedLevel, level, "XP: %d", tt.xp)
	}
}

func TestUpdateStreak_FirstLogin(t *testing.T) {
	svc, mockRepo, _ := setupGamificationService()
	userID := uuid.New()

	mockRepo.On("GetStreakInfo", userID).Return(0, nil, 0, nil)
	mockRepo.On("UpdateStreak", userID, 1, mock.AnythingOfType("time.Time")).Return(nil)

	result, err := svc.UpdateStreak(userID)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.StreakCount)
	assert.True(t, result.StreakContinued)
	assert.False(t, result.StreakReset)
	mockRepo.AssertExpectations(t)
}

func TestUpdateStreak_ContinueStreak(t *testing.T) {
	svc, mockRepo, _ := setupGamificationService()
	userID := uuid.New()
	lastStreak := time.Now().Add(-24 * time.Hour)

	mockRepo.On("GetStreakInfo", userID).Return(5, lastStreak, 0, nil)
	mockRepo.On("UpdateStreak", userID, 6, mock.AnythingOfType("time.Time")).Return(nil)

	result, err := svc.UpdateStreak(userID)

	assert.NoError(t, err)
	assert.Equal(t, 6, result.StreakCount)
	assert.True(t, result.StreakContinued)
	assert.False(t, result.StreakReset)
	mockRepo.AssertExpectations(t)
}

func TestUpdateStreak_SameDay(t *testing.T) {
	svc, mockRepo, _ := setupGamificationService()
	userID := uuid.New()
	lastStreak := time.Now()

	mockRepo.On("GetStreakInfo", userID).Return(5, lastStreak, 0, nil)

	result, err := svc.UpdateStreak(userID)

	assert.NoError(t, err)
	assert.Equal(t, 5, result.StreakCount)
	assert.False(t, result.StreakContinued)
	assert.False(t, result.StreakReset)
	mockRepo.AssertExpectations(t)
}

func TestUpdateStreak_UseFreezeToken(t *testing.T) {
	svc, mockRepo, _ := setupGamificationService()
	userID := uuid.New()
	lastStreak := time.Now().Add(-48 * time.Hour)

	mockRepo.On("GetStreakInfo", userID).Return(5, lastStreak, 1, nil)
	mockRepo.On("UseFreezeToken", userID).Return(nil)
	mockRepo.On("UpdateStreak", userID, 6, mock.AnythingOfType("time.Time")).Return(nil)

	result, err := svc.UpdateStreak(userID)

	assert.NoError(t, err)
	assert.Equal(t, 6, result.StreakCount)
	assert.True(t, result.StreakContinued)
	assert.True(t, result.FreezeTokenUsed)
	assert.False(t, result.StreakReset)
	mockRepo.AssertExpectations(t)
}

func TestUpdateStreak_ResetStreak(t *testing.T) {
	svc, mockRepo, _ := setupGamificationService()
	userID := uuid.New()
	lastStreak := time.Now().Add(-72 * time.Hour)

	mockRepo.On("GetStreakInfo", userID).Return(5, lastStreak, 0, nil)
	mockRepo.On("UpdateStreak", userID, 1, mock.AnythingOfType("time.Time")).Return(nil)

	result, err := svc.UpdateStreak(userID)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.StreakCount)
	assert.False(t, result.StreakContinued)
	assert.True(t, result.StreakReset)
	assert.False(t, result.FreezeTokenUsed)
	mockRepo.AssertExpectations(t)
}

func TestGetUserStats(t *testing.T) {
	svc, mockRepo, _ := setupGamificationService()
	userID := uuid.New()

	expectedStats := &domain.UserStats{
		Points:              150,
		Level:               2,
		Streak:              5,
		TotalTodosCompleted: 10,
	}

	mockRepo.On("GetUserStats", userID).Return(expectedStats, nil)

	stats, err := svc.GetUserStats(userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedStats, stats)
	mockRepo.AssertExpectations(t)
}

func TestGetPointsHistory(t *testing.T) {
	svc, mockRepo, _ := setupGamificationService()
	userID := uuid.New()

	expectedHistory := []*domain.PointsTransaction{
		{
			ID:     uuid.New(),
			UserID: userID,
			Amount: 10,
			Reason: "todo completed",
		},
		{
			ID:     uuid.New(),
			UserID: userID,
			Amount: 50,
			Reason: "streak bonus",
		},
	}

	mockRepo.On("GetPointsHistory", userID, 50).Return(expectedHistory, nil)

	history, err := svc.GetPointsHistory(userID, 50)

	assert.NoError(t, err)
	assert.Equal(t, expectedHistory, history)
	mockRepo.AssertExpectations(t)
}

func TestOnTodoCompleted(t *testing.T) {
	svc, mockRepo, _ := setupGamificationService()
	userID := uuid.New()
	completedAt := time.Now()

	mockRepo.On("AddXP", userID, domain.XPRewardTodoCompleted, string(domain.TransactionReasonTodoCompleted), mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("GetUserXP", userID).Return(10, nil)
	mockRepo.On("GetStreakInfo", userID).Return(0, nil, 0, nil)
	mockRepo.On("UpdateStreak", userID, 1, mock.AnythingOfType("time.Time")).Return(nil)

	svc.OnTodoCompleted(userID, completedAt)

	mockRepo.AssertExpectations(t)
}
