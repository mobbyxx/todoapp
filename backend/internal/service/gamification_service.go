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

type gamificationService struct {
	repo            domain.GamificationRepository
	antiCheatSvc    domain.AntiCheatService
	notificationSvc domain.NotificationService
	redis           *redis.Client
}

func NewGamificationService(
	repo domain.GamificationRepository,
	antiCheatSvc domain.AntiCheatService,
	notificationSvc domain.NotificationService,
	redis *redis.Client,
) domain.GamificationService {
	return &gamificationService{
		repo:            repo,
		antiCheatSvc:    antiCheatSvc,
		notificationSvc: notificationSvc,
		redis:           redis,
	}
}

func (s *gamificationService) AwardXP(userID uuid.UUID, amount int, reason string) error {
	if amount <= 0 {
		return domain.ErrInvalidPointsAmount
	}

	referenceType := string(domain.TransactionReasonTodoCompleted)
	referenceID := uuid.Nil

	err := s.repo.AddXP(userID, amount, reason, referenceType, referenceID)
	if err != nil {
		return fmt.Errorf("failed to award XP: %w", err)
	}

	s.checkAndHandleLevelUp(userID)

	return nil
}

func (s *gamificationService) CalculateLevel(xp int) int {
	return domain.CalculateLevel(xp)
}

func (s *gamificationService) GetUserStats(userID uuid.UUID) (*domain.UserStats, error) {
	return s.repo.GetUserStats(userID)
}

func (s *gamificationService) checkAndHandleLevelUp(userID uuid.UUID) {
	currentXP, err := s.repo.GetUserXP(userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to get user XP for level check")
		return
	}

	currentLevel := domain.CalculateLevel(currentXP)
	previousLevel := currentLevel - 1

	if previousLevel > 0 {
		previousLevelXP := domain.LevelCurve[previousLevel-1]
		if currentXP >= domain.LevelCurve[currentLevel-1] && previousLevelXP < domain.LevelCurve[currentLevel-1] {
			log.Info().
				Str("user_id", userID.String()).
				Int("new_level", currentLevel).
				Int("total_xp", currentXP).
				Msg("user leveled up")
		}
	}
}

func (s *gamificationService) UpdateStreak(userID uuid.UUID) (*domain.StreakUpdateResult, error) {
	result := &domain.StreakUpdateResult{}

	streakCount, lastStreakDate, freezeTokens, err := s.repo.GetStreakInfo(userID)
	if err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)

	if lastStreakDate == nil {
		result.StreakCount = 1
		result.StreakContinued = true
		err = s.repo.UpdateStreak(userID, 1, today)
		if err != nil {
			return nil, err
		}
		s.checkStreakBonuses(userID, 1)
		return result, nil
	}

	lastDate := lastStreakDate.Truncate(24 * time.Hour)
	daysDiff := int(today.Sub(lastDate).Hours() / 24)

	switch daysDiff {
	case 0:
		result.StreakCount = streakCount
		result.StreakContinued = false

	case 1:
		newStreak := streakCount + 1
		result.StreakCount = newStreak
		result.StreakContinued = true
		err = s.repo.UpdateStreak(userID, newStreak, today)
		if err != nil {
			return nil, err
		}
		s.checkStreakBonuses(userID, newStreak)

	default:
		if freezeTokens > 0 && daysDiff == 2 {
			err = s.repo.UseFreezeToken(userID)
			if err == nil {
				newStreak := streakCount + 1
				result.StreakCount = newStreak
				result.StreakContinued = true
				result.FreezeTokenUsed = true
				err = s.repo.UpdateStreak(userID, newStreak, today)
				if err != nil {
					return nil, err
				}
				s.checkStreakBonuses(userID, newStreak)
				return result, nil
			}
		}

		result.StreakCount = 1
		result.StreakReset = true
		err = s.repo.UpdateStreak(userID, 1, today)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *gamificationService) checkStreakBonuses(userID uuid.UUID, streakDays int) {
	ctx := context.Background()
	bonusKey := fmt.Sprintf("streak_bonus:%s:%d", userID.String(), streakDays)

	exists, err := s.redis.Exists(ctx, bonusKey).Result()
	if err != nil || exists > 0 {
		return
	}

	var bonusXP int
	var bonusReason string

	switch streakDays {
	case 7:
		bonusXP = domain.XPRewardStreakBonus7Days
		bonusReason = "7-day streak bonus"
	case 30:
		bonusXP = domain.XPRewardStreakBonus30Days
		bonusReason = "30-day streak bonus"
	default:
		return
	}

	err = s.repo.AddXP(userID, bonusXP, bonusReason, "streak_bonus", uuid.Nil)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Int("streak", streakDays).Msg("failed to award streak bonus")
		return
	}

	s.redis.Set(ctx, bonusKey, "1", 24*time.Hour)

	log.Info().
		Str("user_id", userID.String()).
		Int("streak_days", streakDays).
		Int("bonus_xp", bonusXP).
		Msg("streak bonus awarded")
}

func (s *gamificationService) OnTodoCompleted(userID uuid.UUID, completedAt time.Time) {
	err := s.AwardXP(userID, domain.XPRewardTodoCompleted, string(domain.TransactionReasonTodoCompleted))
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to award XP for todo completion")
		return
	}

	_, err = s.UpdateStreak(userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to update streak")
	}
}

func (s *gamificationService) OnStreakUpdated(userID uuid.UUID, streakDays int) {
	s.checkStreakBonuses(userID, streakDays)
}

func (s *gamificationService) OnConnectionAdded(userID uuid.UUID) {
}

func (s *gamificationService) CheckAndAwardBadges(userID uuid.UUID) ([]*domain.Badge, error) {
	return []*domain.Badge{}, nil
}

func (s *gamificationService) EvaluateBadgeCriteria(userID uuid.UUID, badge *domain.Badge) (bool, error) {
	return false, nil
}

func (s *gamificationService) GetUserBadges(userID uuid.UUID) ([]*domain.BadgeWithEarned, error) {
	return []*domain.BadgeWithEarned{}, nil
}

func (s *gamificationService) AwardBadge(userID uuid.UUID, badgeID uuid.UUID) (*domain.UserBadge, error) {
	return nil, nil
}

func (s *gamificationService) queueBadgeEarnedNotification(userID uuid.UUID, badge *domain.Badge) {
	if s.notificationSvc == nil || badge == nil {
		return
	}

	s.notificationSvc.QueueBadgeEarned(userID, badge)
}

func (s *gamificationService) GetPointsHistory(userID uuid.UUID, limit int) ([]*domain.PointsTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetPointsHistory(userID, limit)
}
