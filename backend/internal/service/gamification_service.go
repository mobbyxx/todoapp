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
	allBadges, err := s.repo.GetAllBadges()
	if err != nil {
		return nil, fmt.Errorf("failed to get badges: %w", err)
	}

	var awarded []*domain.Badge
	for _, badge := range allBadges {
		has, err := s.repo.HasBadge(userID, badge.ID)
		if err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Str("badge_id", badge.ID.String()).Msg("failed to check badge ownership")
			continue
		}
		if has {
			continue
		}

		met, err := s.EvaluateBadgeCriteria(userID, badge)
		if err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Str("badge", badge.Name).Msg("failed to evaluate badge criteria")
			continue
		}
		if !met {
			continue
		}

		userBadge, err := s.AwardBadge(userID, badge.ID)
		if err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Str("badge", badge.Name).Msg("failed to award badge")
			continue
		}
		if userBadge != nil {
			awarded = append(awarded, badge)
			s.queueBadgeEarnedNotification(userID, badge)
		}
	}

	return awarded, nil
}

func (s *gamificationService) EvaluateBadgeCriteria(userID uuid.UUID, badge *domain.Badge) (bool, error) {
	criteria, err := s.repo.GetBadgeCriteria(badge.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get badge criteria: %w", err)
	}

	switch criteria.Type {
	case "todo_completed":
		count, err := s.repo.CountCompletedTodos(userID)
		if err != nil {
			return false, err
		}
		return count >= criteria.Count, nil

	case "streak":
		streakCount, _, _, err := s.repo.GetStreakInfo(userID)
		if err != nil {
			return false, err
		}
		return streakCount >= criteria.Days, nil

	case "shared_tasks":
		count, err := s.repo.CountConnections(userID)
		if err != nil {
			return false, err
		}
		return count >= criteria.Count, nil

	case "early_completion":
		count, err := s.repo.CountEarlyCompletions(userID, criteria.BeforeHour)
		if err != nil {
			return false, err
		}
		return count >= criteria.Count, nil

	default:
		log.Warn().Str("type", criteria.Type).Str("badge", badge.Name).Msg("unknown badge criteria type")
		return false, nil
	}
}

func (s *gamificationService) GetUserBadges(userID uuid.UUID) ([]*domain.BadgeWithEarned, error) {
	allBadges, err := s.repo.GetAllBadges()
	if err != nil {
		return nil, fmt.Errorf("failed to get all badges: %w", err)
	}

	earnedBadges, err := s.repo.GetUserBadges(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user badges: %w", err)
	}

	earnedMap := make(map[uuid.UUID]time.Time, len(earnedBadges))
	for _, ub := range earnedBadges {
		earnedMap[ub.BadgeID] = ub.EarnedAt
	}

	result := make([]*domain.BadgeWithEarned, 0, len(allBadges))
	for _, badge := range allBadges {
		bwe := &domain.BadgeWithEarned{
			Badge: *badge,
		}
		if earnedAt, ok := earnedMap[badge.ID]; ok {
			bwe.Earned = true
			t := earnedAt
			bwe.EarnedAt = &t
		}
		result = append(result, bwe)
	}

	return result, nil
}

func (s *gamificationService) AwardBadge(userID uuid.UUID, badgeID uuid.UUID) (*domain.UserBadge, error) {
	has, err := s.repo.HasBadge(userID, badgeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check badge: %w", err)
	}
	if has {
		return nil, nil
	}

	badge, err := s.repo.GetBadgeByID(badgeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get badge: %w", err)
	}

	userBadge := &domain.UserBadge{
		ID:       uuid.New(),
		UserID:   userID,
		BadgeID:  badgeID,
		EarnedAt: time.Now(),
	}

	if err := s.repo.AwardBadge(userBadge); err != nil {
		return nil, fmt.Errorf("failed to award badge: %w", err)
	}

	if badge.PointsValue > 0 {
		err = s.repo.AddXP(userID, badge.PointsValue, string(domain.TransactionReasonBadgeEarned), "badge", badgeID)
		if err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Str("badge_id", badgeID.String()).Msg("failed to award badge XP")
		}
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("badge_id", badgeID.String()).
		Str("badge_name", badge.Name).
		Msg("badge awarded")

	return userBadge, nil
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
