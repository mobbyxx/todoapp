package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRewardNotFound         = errors.New("reward not found")
	ErrRedemptionNotFound     = errors.New("redemption not found")
	ErrInsufficientXP         = errors.New("insufficient XP")
	ErrRewardNotActive        = errors.New("reward is not active")
	ErrInvalidRewardCost      = errors.New("reward cost must be positive")
	ErrInvalidRewardName      = errors.New("reward name must be 1-100 characters")
	ErrInvalidRewardDescription = errors.New("reward description must be max 500 characters")
)

// RedemptionStatus represents the status of a reward redemption
type RedemptionStatus string

const (
	RedemptionStatusPending   RedemptionStatus = "pending"
	RedemptionStatusApproved  RedemptionStatus = "approved"
	RedemptionStatusRejected  RedemptionStatus = "rejected"
	RedemptionStatusCompleted RedemptionStatus = "completed"
)

func (s RedemptionStatus) IsValid() bool {
	switch s {
	case RedemptionStatusPending, RedemptionStatusApproved, RedemptionStatusRejected, RedemptionStatusCompleted:
		return true
	}
	return false
}

// Reward represents a user-defined reward that can be redeemed with XP
type Reward struct {
	ID          uuid.UUID  `json:"id"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Cost        int        `json:"cost"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// RewardRedemption represents a redemption of a reward by a user
type RewardRedemption struct {
	ID          uuid.UUID        `json:"id"`
	UserID      uuid.UUID        `json:"user_id"`
	RewardID    uuid.UUID        `json:"reward_id"`
	Status      RedemptionStatus `json:"status"`
	RedeemedAt  time.Time        `json:"redeemed_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// RewardWithRedemption extends Reward with redemption status for the current user
type RewardWithRedemption struct {
	Reward
	HasRedeemed    bool                   `json:"has_redeemed"`
	RedemptionStatus *RedemptionStatus    `json:"redemption_status,omitempty"`
}

// CreateRewardInput represents the input for creating a new reward
type CreateRewardInput struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
	Cost        int    `json:"cost" validate:"required,min=1"`
}

// UpdateRewardInput represents the input for updating a reward
type UpdateRewardInput struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Cost        *int    `json:"cost,omitempty" validate:"omitempty,min=1"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// RewardRepository defines the interface for reward data access
type RewardRepository interface {
	Create(reward *Reward) error
	GetByID(id uuid.UUID) (*Reward, error)
	GetByUserID(userID uuid.UUID) ([]*Reward, error)
	GetAllActive() ([]*Reward, error)
	Update(reward *Reward) error
	Delete(id uuid.UUID) error
	CreateRedemption(redemption *RewardRedemption) error
	GetRedemptionByID(id uuid.UUID) (*RewardRedemption, error)
	GetRedemptionsByUser(userID uuid.UUID) ([]*RewardRedemption, error)
	GetRedemptionsByReward(rewardID uuid.UUID) ([]*RewardRedemption, error)
	UpdateRedemptionStatus(id uuid.UUID, status RedemptionStatus) error
}

// RewardService defines the interface for reward business logic
type RewardService interface {
	CreateReward(userID uuid.UUID, input CreateRewardInput) (*Reward, error)
	GetRewards(userID uuid.UUID) ([]*Reward, error)
	GetMyRewards(userID uuid.UUID) ([]*RewardRedemption, error)
	RedeemReward(userID uuid.UUID, rewardID uuid.UUID) (*RewardRedemption, error)
	ApproveRedemption(redemptionID uuid.UUID) error
	GetRewardByID(userID uuid.UUID, rewardID uuid.UUID) (*Reward, error)
}
