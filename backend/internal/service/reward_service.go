package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

type rewardService struct {
	rewardRepo   domain.RewardRepository
	gamification domain.GamificationService
}

func NewRewardService(
	rewardRepo domain.RewardRepository,
	gamification domain.GamificationService,
) domain.RewardService {
	return &rewardService{
		rewardRepo:   rewardRepo,
		gamification: gamification,
	}
}

func (s *rewardService) CreateReward(userID uuid.UUID, input domain.CreateRewardInput) (*domain.Reward, error) {
	if input.Name == "" {
		return nil, domain.ErrInvalidRewardName
	}

	if len(input.Name) > 100 {
		return nil, domain.ErrInvalidRewardName
	}

	if len(input.Description) > 500 {
		return nil, domain.ErrInvalidRewardDescription
	}

	if input.Cost <= 0 {
		return nil, domain.ErrInvalidRewardCost
	}

	reward := &domain.Reward{
		UserID:      &userID,
		Name:        input.Name,
		Description: input.Description,
		Cost:        input.Cost,
		IsActive:    true,
	}

	if err := s.rewardRepo.Create(reward); err != nil {
		return nil, fmt.Errorf("failed to create reward: %w", err)
	}

	return reward, nil
}

func (s *rewardService) GetRewards(userID uuid.UUID) ([]*domain.Reward, error) {
	rewards, err := s.rewardRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rewards: %w", err)
	}
	return rewards, nil
}

func (s *rewardService) GetMyRewards(userID uuid.UUID) ([]*domain.RewardRedemption, error) {
	redemptions, err := s.rewardRepo.GetRedemptionsByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get redemptions: %w", err)
	}
	return redemptions, nil
}

func (s *rewardService) RedeemReward(userID uuid.UUID, rewardID uuid.UUID) (*domain.RewardRedemption, error) {
	reward, err := s.rewardRepo.GetByID(rewardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reward: %w", err)
	}

	if !reward.IsActive {
		return nil, domain.ErrRewardNotActive
	}

	userStats, err := s.gamification.GetUserStats(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	if userStats.Points < reward.Cost {
		return nil, domain.ErrInsufficientXP
	}

	xpDeduction := -reward.Cost
	err = s.gamification.AwardXP(userID, xpDeduction, fmt.Sprintf("Redeemed reward: %s", reward.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to deduct XP: %w", err)
	}

	redemption := &domain.RewardRedemption{
		UserID:   userID,
		RewardID: rewardID,
		Status:   domain.RedemptionStatusCompleted,
	}

	if err := s.rewardRepo.CreateRedemption(redemption); err != nil {
		return nil, fmt.Errorf("failed to create redemption: %w", err)
	}

	return redemption, nil
}

func (s *rewardService) ApproveRedemption(redemptionID uuid.UUID) error {
	_, err := s.rewardRepo.GetRedemptionByID(redemptionID)
	if err != nil {
		return fmt.Errorf("failed to get redemption: %w", err)
	}

	if err := s.rewardRepo.UpdateRedemptionStatus(redemptionID, domain.RedemptionStatusApproved); err != nil {
		return fmt.Errorf("failed to approve redemption: %w", err)
	}

	return nil
}

func (s *rewardService) GetRewardByID(userID uuid.UUID, rewardID uuid.UUID) (*domain.Reward, error) {
	reward, err := s.rewardRepo.GetByID(rewardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reward: %w", err)
	}
	return reward, nil
}
