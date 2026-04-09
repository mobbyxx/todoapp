package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Gamification-related errors
var (
	ErrInvalidPointsAmount = errors.New("invalid points amount")
	ErrInvalidLevelNumber  = errors.New("invalid level number")
	ErrMaxLevelReached     = errors.New("maximum level reached")
	ErrInvalidTransaction  = errors.New("invalid points transaction")
)

// XP Reward Constants
const (
	XPRewardTodoCompleted       = 10
	XPRewardStreakBonus7Days    = 50
	XPRewardStreakBonus30Days   = 200
	XPRewardPerfectDay          = 25
)

var LevelCurve = []int{
	0,
	100,
	300,
	700,
	1500,
	3000,
	6000,
	12000,
}

const MaxLevel = 8

// UserStats represents a user's gamification statistics
type UserStats struct {
	Points            int       `json:"points"`
	Level             int       `json:"level"`
	Streak            int       `json:"streak"`
	LastActiveAt      time.Time `json:"last_active_at"`
	TotalTodosCompleted int     `json:"total_todos_completed"`
}

// Badge represents a badge definition
type Badge struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Icon         string    `json:"icon"`
	PointsValue  int       `json:"points_value"`
	CreatedAt    time.Time `json:"created_at"`
}

// Level represents a level definition
type Level struct {
	ID           uuid.UUID `json:"id"`
	LevelNumber  int       `json:"level_number"`
	Name         string    `json:"name"`
	MinPoints    int       `json:"min_points"`
	MaxPoints    int       `json:"max_points"`
	CreatedAt    time.Time `json:"created_at"`
}

// PointsTransaction represents an immutable ledger entry for points
type PointsTransaction struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Amount        int       `json:"amount"`
	Reason        string    `json:"reason"`
	ReferenceType string    `json:"reference_type"`
	ReferenceID   uuid.UUID `json:"reference_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// TransactionReason defines common reasons for points transactions
type TransactionReason string

const (
	TransactionReasonTodoCompleted     TransactionReason = "todo_completed"
	TransactionReasonStreakBonus       TransactionReason = "streak_bonus"
	TransactionReasonPerfectDay        TransactionReason = "perfect_day"
	TransactionReasonBadgeEarned       TransactionReason = "badge_earned"
	TransactionReasonLevelUp           TransactionReason = "level_up"
)

func CalculateLevel(xp int) int {
	if xp < 0 {
		return 1
	}

	for level := len(LevelCurve); level > 0; level-- {
		if xp >= LevelCurve[level-1] {
			return level
		}
	}

	return 1
}

func GetXPForNextLevel(currentXP int) int {
	currentLevel := CalculateLevel(currentXP)

	if currentLevel >= MaxLevel {
		return 0
	}

	nextLevelMinXP := LevelCurve[currentLevel]
	xpNeeded := nextLevelMinXP - currentXP

	if xpNeeded < 0 {
		return 0
	}

	return xpNeeded
}

// GetLevelProgress returns the progress towards the next level as a percentage (0-100)
func GetLevelProgress(currentXP int) float64 {
	currentLevel := CalculateLevel(currentXP)

	if currentLevel >= MaxLevel {
		return 100.0
	}

	currentLevelMinXP := LevelCurve[currentLevel-1]
	nextLevelMinXP := LevelCurve[currentLevel]

	if nextLevelMinXP == currentLevelMinXP {
		return 100.0
	}

	progress := float64(currentXP-currentLevelMinXP) / float64(nextLevelMinXP-currentLevelMinXP)
	if progress > 1.0 {
		progress = 1.0
	}

	return progress * 100.0
}

// IsMaxLevel checks if the given level is the maximum level
func IsMaxLevel(level int) bool {
	return level >= MaxLevel
}

// GetLevelName returns a display name for a level
func GetLevelName(level int) string {
	names := map[int]string{
		1: "Novice",
		2: "Beginner",
		3: "Apprentice",
		4: "Practitioner",
		5: "Expert",
		6: "Master",
		7: "Grandmaster",
		8: "Legend",
	}

	if name, ok := names[level]; ok {
		return name
	}
	return "Unknown"
}

// UserBadge represents a badge earned by a user
type UserBadge struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	BadgeID   uuid.UUID `json:"badge_id"`
	EarnedAt  time.Time `json:"earned_at"`
}

// UserLevel represents a user's current level progression
type UserLevel struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	CurrentLevelID uuid.UUID `json:"current_level_id"`
	CurrentXP      int       `json:"current_xp"`
	TotalXP        int       `json:"total_xp"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type StreakUpdateResult struct {
	StreakCount     int  `json:"streak_count"`
	StreakContinued bool `json:"streak_continued"`
	StreakReset     bool `json:"streak_reset"`
	FreezeTokenUsed bool `json:"freeze_token_used"`
	BonusXPAwarded  int  `json:"bonus_xp_awarded"`
}

// BadgeCriteria represents the criteria for earning a badge
type BadgeCriteria struct {
	Type        string `json:"type"`
	Count       int    `json:"count,omitempty"`
	Days        int    `json:"days,omitempty"`
	BeforeHour  int    `json:"before_hour,omitempty"`
}

// BadgeWithEarned extends Badge with earned status
type BadgeWithEarned struct {
	Badge
	Earned   bool       `json:"earned"`
	EarnedAt *time.Time `json:"earned_at,omitempty"`
}

// GamificationRepository defines the interface for gamification data access
type GamificationRepository interface {
	GetAllBadges() ([]*Badge, error)
	GetBadgeByID(id uuid.UUID) (*Badge, error)
	GetUserBadges(userID uuid.UUID) ([]*UserBadge, error)
	HasBadge(userID, badgeID uuid.UUID) (bool, error)
	AwardBadge(userBadge *UserBadge) error
	GetUserStats(userID uuid.UUID) (*UserStats, error)
	CountCompletedTodos(userID uuid.UUID) (int, error)
	CountEarlyCompletions(userID uuid.UUID, beforeHour int) (int, error)
	CountConnections(userID uuid.UUID) (int, error)
	AddXP(userID uuid.UUID, amount int, reason string, referenceType string, referenceID uuid.UUID) error
	GetUserXP(userID uuid.UUID) (int, error)
	GetPointsHistory(userID uuid.UUID, limit int) ([]*PointsTransaction, error)
	UpdateStreak(userID uuid.UUID, newStreak int, lastStreakDate time.Time) error
	GetStreakInfo(userID uuid.UUID) (streakCount int, lastStreakDate *time.Time, freezeTokens int, err error)
	UseFreezeToken(userID uuid.UUID) error
	AwardFreezeToken(userID uuid.UUID) error
}

// GamificationService defines the interface for gamification business logic
type GamificationService interface {
	CheckAndAwardBadges(userID uuid.UUID) ([]*Badge, error)
	EvaluateBadgeCriteria(userID uuid.UUID, badge *Badge) (bool, error)
	GetUserBadges(userID uuid.UUID) ([]*BadgeWithEarned, error)
	AwardBadge(userID uuid.UUID, badgeID uuid.UUID) (*UserBadge, error)
	AwardXP(userID uuid.UUID, amount int, reason string) error
	GetUserStats(userID uuid.UUID) (*UserStats, error)
	OnTodoCompleted(userID uuid.UUID, completedAt time.Time)
	OnStreakUpdated(userID uuid.UUID, streakDays int)
	OnConnectionAdded(userID uuid.UUID)
}
