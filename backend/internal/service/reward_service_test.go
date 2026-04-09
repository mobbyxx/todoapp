package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/todo-api/internal/domain"
)

type mockRewardRepository struct {
	mock.Mock
	rewards     map[uuid.UUID]*domain.Reward
	redemptions map[uuid.UUID]*domain.RewardRedemption
}

func newMockRewardRepository() *mockRewardRepository {
	return &mockRewardRepository{
		rewards:     make(map[uuid.UUID]*domain.Reward),
		redemptions: make(map[uuid.UUID]*domain.RewardRedemption),
	}
}

func (m *mockRewardRepository) Create(reward *domain.Reward) error {
	args := m.Called(reward)
	if args.Error(0) == nil {
		m.rewards[reward.ID] = reward
	}
	return args.Error(0)
}

func (m *mockRewardRepository) GetByID(id uuid.UUID) (*domain.Reward, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Reward), args.Error(1)
}

func (m *mockRewardRepository) GetByUserID(userID uuid.UUID) ([]*domain.Reward, error) {
	args := m.Called(userID)
	return args.Get(0).([]*domain.Reward), args.Error(1)
}

func (m *mockRewardRepository) GetAllActive() ([]*domain.Reward, error) {
	args := m.Called()
	return args.Get(0).([]*domain.Reward), args.Error(1)
}

func (m *mockRewardRepository) Update(reward *domain.Reward) error {
	args := m.Called(reward)
	return args.Error(0)
}

func (m *mockRewardRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockRewardRepository) CreateRedemption(redemption *domain.RewardRedemption) error {
	args := m.Called(redemption)
	if args.Error(0) == nil {
		m.redemptions[redemption.ID] = redemption
	}
	return args.Error(0)
}

func (m *mockRewardRepository) GetRedemptionByID(id uuid.UUID) (*domain.RewardRedemption, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RewardRedemption), args.Error(1)
}

func (m *mockRewardRepository) GetRedemptionsByUser(userID uuid.UUID) ([]*domain.RewardRedemption, error) {
	args := m.Called(userID)
	return args.Get(0).([]*domain.RewardRedemption), args.Error(1)
}

func (m *mockRewardRepository) GetRedemptionsByReward(rewardID uuid.UUID) ([]*domain.RewardRedemption, error) {
	args := m.Called(rewardID)
	return args.Get(0).([]*domain.RewardRedemption), args.Error(1)
}

func (m *mockRewardRepository) UpdateRedemptionStatus(id uuid.UUID, status domain.RedemptionStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

type mockGamificationServiceForRewards struct {
	mock.Mock
}

func (m *mockGamificationServiceForRewards) CheckAndAwardBadges(userID uuid.UUID) ([]*domain.Badge, error) {
	args := m.Called(userID)
	return args.Get(0).([]*domain.Badge), args.Error(1)
}

func (m *mockGamificationServiceForRewards) EvaluateBadgeCriteria(userID uuid.UUID, badge *domain.Badge) (bool, error) {
	args := m.Called(userID, badge)
	return args.Bool(0), args.Error(1)
}

func (m *mockGamificationServiceForRewards) GetUserBadges(userID uuid.UUID) ([]*domain.BadgeWithEarned, error) {
	args := m.Called(userID)
	return args.Get(0).([]*domain.BadgeWithEarned), args.Error(1)
}

func (m *mockGamificationServiceForRewards) AwardBadge(userID uuid.UUID, badgeID uuid.UUID) (*domain.UserBadge, error) {
	args := m.Called(userID, badgeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserBadge), args.Error(1)
}

func (m *mockGamificationServiceForRewards) AwardXP(userID uuid.UUID, amount int, reason string) error {
	args := m.Called(userID, amount, reason)
	return args.Error(0)
}

func (m *mockGamificationServiceForRewards) GetUserStats(userID uuid.UUID) (*domain.UserStats, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserStats), args.Error(1)
}

func (m *mockGamificationServiceForRewards) GetPointsHistory(userID uuid.UUID, limit int) ([]*domain.PointsTransaction, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.PointsTransaction), args.Error(1)
}

func (m *mockGamificationServiceForRewards) OnTodoCompleted(userID uuid.UUID, completedAt time.Time) {
	m.Called(userID, completedAt)
}

func (m *mockGamificationServiceForRewards) OnStreakUpdated(userID uuid.UUID, streakDays int) {
	m.Called(userID, streakDays)
}

func (m *mockGamificationServiceForRewards) OnConnectionAdded(userID uuid.UUID) {
	m.Called(userID)
}

func setupRewardService() (*rewardService, *mockRewardRepository, *mockGamificationServiceForRewards) {
	mockRepo := newMockRewardRepository()
	mockGamification := &mockGamificationServiceForRewards{}

	svc := &rewardService{
		rewardRepo:   mockRepo,
		gamification: mockGamification,
	}

	return svc, mockRepo, mockGamification
}

func TestRewardCreate_Success(t *testing.T) {
	svc, mockRepo, _ := setupRewardService()
	userID := uuid.New()

	input := domain.CreateRewardInput{
		Name:        "Test Reward",
		Description: "A test reward",
		Cost:        100,
	}

	mockRepo.On("Create", mock.AnythingOfType("*domain.Reward")).Return(nil)

	reward, err := svc.CreateReward(userID, input)

	assert.NoError(t, err)
	assert.NotNil(t, reward)
	assert.Equal(t, input.Name, reward.Name)
	assert.Equal(t, input.Description, reward.Description)
	assert.Equal(t, input.Cost, reward.Cost)
	assert.True(t, reward.IsActive)
	assert.Equal(t, &userID, reward.UserID)
	mockRepo.AssertExpectations(t)
}

func TestRewardCreate_InvalidName(t *testing.T) {
	svc, _, _ := setupRewardService()
	userID := uuid.New()

	input := domain.CreateRewardInput{
		Name:        "",
		Description: "A test reward",
		Cost:        100,
	}

	reward, err := svc.CreateReward(userID, input)

	assert.ErrorIs(t, err, domain.ErrInvalidRewardName)
	assert.Nil(t, reward)
}

func TestRewardCreate_NameTooLong(t *testing.T) {
	svc, _, _ := setupRewardService()
	userID := uuid.New()

	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}

	input := domain.CreateRewardInput{
		Name:        longName,
		Description: "A test reward",
		Cost:        100,
	}

	reward, err := svc.CreateReward(userID, input)

	assert.ErrorIs(t, err, domain.ErrInvalidRewardName)
	assert.Nil(t, reward)
}

func TestRewardCreate_InvalidCost(t *testing.T) {
	svc, _, _ := setupRewardService()
	userID := uuid.New()

	input := domain.CreateRewardInput{
		Name:        "Test Reward",
		Description: "A test reward",
		Cost:        0,
	}

	reward, err := svc.CreateReward(userID, input)

	assert.ErrorIs(t, err, domain.ErrInvalidRewardCost)
	assert.Nil(t, reward)
}

func TestRewardCreate_NegativeCost(t *testing.T) {
	svc, _, _ := setupRewardService()
	userID := uuid.New()

	input := domain.CreateRewardInput{
		Name:        "Test Reward",
		Description: "A test reward",
		Cost:        -50,
	}

	reward, err := svc.CreateReward(userID, input)

	assert.ErrorIs(t, err, domain.ErrInvalidRewardCost)
	assert.Nil(t, reward)
}

func TestGetRewards(t *testing.T) {
	svc, mockRepo, _ := setupRewardService()
	userID := uuid.New()

	expectedRewards := []*domain.Reward{
		{
			ID:          uuid.New(),
			Name:        "Reward 1",
			Description: "First reward",
			Cost:        100,
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			Name:        "Reward 2",
			Description: "Second reward",
			Cost:        200,
			IsActive:    true,
		},
	}

	mockRepo.On("GetByUserID", userID).Return(expectedRewards, nil)

	rewards, err := svc.GetRewards(userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedRewards, rewards)
	mockRepo.AssertExpectations(t)
}

func TestGetMyRewards(t *testing.T) {
	svc, mockRepo, _ := setupRewardService()
	userID := uuid.New()

	expectedRedemptions := []*domain.RewardRedemption{
		{
			ID:       uuid.New(),
			UserID:   userID,
			RewardID: uuid.New(),
			Status:   domain.RedemptionStatusCompleted,
		},
		{
			ID:       uuid.New(),
			UserID:   userID,
			RewardID: uuid.New(),
			Status:   domain.RedemptionStatusPending,
		},
	}

	mockRepo.On("GetRedemptionsByUser", userID).Return(expectedRedemptions, nil)

	redemptions, err := svc.GetMyRewards(userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedRedemptions, redemptions)
	mockRepo.AssertExpectations(t)
}

func TestRedeemReward_Success(t *testing.T) {
	svc, mockRepo, mockGamification := setupRewardService()
	userID := uuid.New()
	rewardID := uuid.New()

	reward := &domain.Reward{
		ID:          rewardID,
		Name:        "Test Reward",
		Description: "A test reward",
		Cost:        100,
		IsActive:    true,
	}

	userStats := &domain.UserStats{
		Points: 200,
		Level:  2,
	}

	mockRepo.On("GetByID", rewardID).Return(reward, nil)
	mockGamification.On("GetUserStats", userID).Return(userStats, nil)
	mockGamification.On("AwardXP", userID, -100, "Redeemed reward: Test Reward").Return(nil)
	mockRepo.On("CreateRedemption", mock.AnythingOfType("*domain.RewardRedemption")).Return(nil)

	redemption, err := svc.RedeemReward(userID, rewardID)

	assert.NoError(t, err)
	assert.NotNil(t, redemption)
	assert.Equal(t, userID, redemption.UserID)
	assert.Equal(t, rewardID, redemption.RewardID)
	assert.Equal(t, domain.RedemptionStatusCompleted, redemption.Status)
	mockRepo.AssertExpectations(t)
	mockGamification.AssertExpectations(t)
}

func TestRedeemReward_RewardNotFound(t *testing.T) {
	svc, mockRepo, _ := setupRewardService()
	userID := uuid.New()
	rewardID := uuid.New()

	mockRepo.On("GetByID", rewardID).Return(nil, domain.ErrRewardNotFound)

	redemption, err := svc.RedeemReward(userID, rewardID)

	assert.Error(t, err)
	assert.Nil(t, redemption)
	mockRepo.AssertExpectations(t)
}

func TestRedeemReward_NotActive(t *testing.T) {
	svc, mockRepo, _ := setupRewardService()
	userID := uuid.New()
	rewardID := uuid.New()

	reward := &domain.Reward{
		ID:          rewardID,
		Name:        "Test Reward",
		Description: "A test reward",
		Cost:        100,
		IsActive:    false,
	}

	mockRepo.On("GetByID", rewardID).Return(reward, nil)

	redemption, err := svc.RedeemReward(userID, rewardID)

	assert.ErrorIs(t, err, domain.ErrRewardNotActive)
	assert.Nil(t, redemption)
	mockRepo.AssertExpectations(t)
}

func TestRedeemReward_InsufficientXP(t *testing.T) {
	svc, mockRepo, mockGamification := setupRewardService()
	userID := uuid.New()
	rewardID := uuid.New()

	reward := &domain.Reward{
		ID:          rewardID,
		Name:        "Test Reward",
		Description: "A test reward",
		Cost:        100,
		IsActive:    true,
	}

	userStats := &domain.UserStats{
		Points: 50,
		Level:  1,
	}

	mockRepo.On("GetByID", rewardID).Return(reward, nil)
	mockGamification.On("GetUserStats", userID).Return(userStats, nil)

	redemption, err := svc.RedeemReward(userID, rewardID)

	assert.ErrorIs(t, err, domain.ErrInsufficientXP)
	assert.Nil(t, redemption)
	mockRepo.AssertExpectations(t)
	mockGamification.AssertExpectations(t)
}

func TestApproveRedemption_Success(t *testing.T) {
	svc, mockRepo, _ := setupRewardService()
	redemptionID := uuid.New()

	redemption := &domain.RewardRedemption{
		ID:     redemptionID,
		Status: domain.RedemptionStatusPending,
	}

	mockRepo.On("GetRedemptionByID", redemptionID).Return(redemption, nil)
	mockRepo.On("UpdateRedemptionStatus", redemptionID, domain.RedemptionStatusApproved).Return(nil)

	err := svc.ApproveRedemption(redemptionID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestApproveRedemption_NotFound(t *testing.T) {
	svc, mockRepo, _ := setupRewardService()
	redemptionID := uuid.New()

	mockRepo.On("GetRedemptionByID", redemptionID).Return(nil, domain.ErrRedemptionNotFound)

	err := svc.ApproveRedemption(redemptionID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetRewardByID_Success(t *testing.T) {
	svc, mockRepo, _ := setupRewardService()
	userID := uuid.New()
	rewardID := uuid.New()

	reward := &domain.Reward{
		ID:          rewardID,
		Name:        "Test Reward",
		Description: "A test reward",
		Cost:        100,
		IsActive:    true,
	}

	mockRepo.On("GetByID", rewardID).Return(reward, nil)

	result, err := svc.GetRewardByID(userID, rewardID)

	assert.NoError(t, err)
	assert.Equal(t, reward, result)
	mockRepo.AssertExpectations(t)
}

func TestGetRewardByID_NotFound(t *testing.T) {
	svc, mockRepo, _ := setupRewardService()
	userID := uuid.New()
	rewardID := uuid.New()

	mockRepo.On("GetByID", rewardID).Return(nil, domain.ErrRewardNotFound)

	result, err := svc.GetRewardByID(userID, rewardID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}
