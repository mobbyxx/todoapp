package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/user/todo-api/internal/domain"
)

type sharedGoalService struct {
	sharedGoalRepo   domain.SharedGoalRepository
	connectionRepo   domain.ConnectionRepository
	connectionSvc    domain.ConnectionService
	gamificationSvc  domain.GamificationService
	notificationSvc  domain.NotificationService
}

func NewSharedGoalService(
	sharedGoalRepo domain.SharedGoalRepository,
	connectionRepo domain.ConnectionRepository,
	connectionSvc domain.ConnectionService,
	gamificationSvc domain.GamificationService,
	notificationSvc domain.NotificationService,
) domain.SharedGoalService {
	return &sharedGoalService{
		sharedGoalRepo:  sharedGoalRepo,
		connectionRepo:  connectionRepo,
		connectionSvc:   connectionSvc,
		gamificationSvc: gamificationSvc,
		notificationSvc: notificationSvc,
	}
}

func (s *sharedGoalService) CreateGoal(connectionID uuid.UUID, targetType domain.SharedGoalTargetType, targetValue int, rewardDescription string) (*domain.SharedGoal, error) {
	if err := domain.ValidateTargetType(targetType); err != nil {
		return nil, err
	}

	if err := domain.ValidateTargetValue(targetValue); err != nil {
		return nil, err
	}

	goal := &domain.SharedGoal{
		ConnectionID:      connectionID,
		TargetType:        targetType,
		TargetValue:       targetValue,
		RewardDescription: rewardDescription,
	}

	if err := s.sharedGoalRepo.Create(goal); err != nil {
		return nil, fmt.Errorf("failed to create shared goal: %w", err)
	}

	log.Info().
		Str("goal_id", goal.ID.String()).
		Str("connection_id", connectionID.String()).
		Str("target_type", string(targetType)).
		Int("target_value", targetValue).
		Msg("shared goal created")

	return goal, nil
}

func (s *sharedGoalService) UpdateProgress(connectionID uuid.UUID, amount int) error {
	if amount <= 0 {
		return nil
	}

	goals, err := s.sharedGoalRepo.GetByConnection(connectionID)
	if err != nil {
		return fmt.Errorf("failed to get goals for connection: %w", err)
	}

	for _, goal := range goals {
		if !goal.IsActive() {
			continue
		}

		if err := s.sharedGoalRepo.UpdateProgress(goal.ID, amount); err != nil {
			log.Warn().Err(err).
				Str("goal_id", goal.ID.String()).
				Msg("failed to update goal progress")
			continue
		}

		updatedGoal, err := s.sharedGoalRepo.GetByID(goal.ID)
		if err != nil {
			log.Warn().Err(err).
				Str("goal_id", goal.ID.String()).
				Msg("failed to get updated goal")
			continue
		}

		if updatedGoal.IsCompleted() && !goal.IsCompleted() {
			s.handleGoalCompletion(updatedGoal)
		}
	}

	return nil
}

func (s *sharedGoalService) CheckCompletion(goalID uuid.UUID) (*domain.SharedGoal, error) {
	goal, err := s.sharedGoalRepo.GetByID(goalID)
	if err != nil {
		return nil, err
	}

	if goal.IsCompleted() {
		return goal, nil
	}

	if goal.CurrentValue >= goal.TargetValue {
		goal.MarkAsCompleted()
		if err := s.sharedGoalRepo.Update(goal); err != nil {
			return nil, fmt.Errorf("failed to mark goal as completed: %w", err)
		}
		s.handleGoalCompletion(goal)
	}

	return goal, nil
}

func (s *sharedGoalService) ListGoals(userID uuid.UUID) ([]*domain.SharedGoal, error) {
	return s.sharedGoalRepo.GetByUserID(userID)
}

func (s *sharedGoalService) OnTodoCompleted(userID uuid.UUID) {
	connections, err := s.connectionSvc.GetConnections(userID)
	if err != nil {
		log.Warn().Err(err).
			Str("user_id", userID.String()).
			Msg("failed to get user connections for goal progress")
		return
	}

	for _, conn := range connections {
		goals, err := s.sharedGoalRepo.GetByConnection(conn.ID)
		if err != nil {
			log.Warn().Err(err).
				Str("connection_id", conn.ID.String()).
				Msg("failed to get goals for connection")
			continue
		}

		for _, goal := range goals {
			if !goal.IsActive() {
				continue
			}

			if goal.TargetType == domain.SharedGoalTargetTypeTodosCompleted {
				if err := s.sharedGoalRepo.UpdateProgress(goal.ID, 1); err != nil {
					log.Warn().Err(err).
						Str("goal_id", goal.ID.String()).
						Msg("failed to update goal progress")
					continue
				}

				updatedGoal, err := s.sharedGoalRepo.GetByID(goal.ID)
				if err != nil {
					log.Warn().Err(err).
						Str("goal_id", goal.ID.String()).
						Msg("failed to get updated goal")
					continue
				}

				if updatedGoal.IsCompleted() && !goal.IsCompleted() {
					s.handleGoalCompletion(updatedGoal)
				}
			}
		}
	}
}

func (s *sharedGoalService) handleGoalCompletion(goal *domain.SharedGoal) {
	connection, err := s.getConnection(goal.ConnectionID)
	if err != nil {
		log.Warn().Err(err).
			Str("goal_id", goal.ID.String()).
			Str("connection_id", goal.ConnectionID.String()).
			Msg("failed to get connection for goal completion")
		return
	}

	s.awardCompletionBonus(connection.UserAID, goal)
	s.awardCompletionBonus(connection.UserBID, goal)
	s.queueGoalCompletedNotifications(connection, goal)

	log.Info().
		Str("goal_id", goal.ID.String()).
		Str("connection_id", goal.ConnectionID.String()).
		Int("bonus_xp", domain.SharedGoalCompletionBonusXP).
		Msg("shared goal completed, bonus XP awarded to both users")
}

func (s *sharedGoalService) getConnection(connectionID uuid.UUID) (*domain.Connection, error) {
	return s.connectionRepo.GetByID(connectionID)
}

func (s *sharedGoalService) awardCompletionBonus(userID uuid.UUID, goal *domain.SharedGoal) {
	if s.gamificationSvc == nil {
		return
	}

	err := s.gamificationSvc.AwardXP(
		userID,
		domain.SharedGoalCompletionBonusXP,
		fmt.Sprintf("Shared goal completed: %s", goal.RewardDescription),
	)
	if err != nil {
		log.Warn().Err(err).
			Str("user_id", userID.String()).
			Str("goal_id", goal.ID.String()).
			Msg("failed to award completion bonus XP")
	}
}

func (s *sharedGoalService) queueGoalCompletedNotifications(connection *domain.Connection, goal *domain.SharedGoal) {
	if s.notificationSvc == nil {
		return
	}

	s.notificationSvc.QueueGoalCompleted(
		connection.UserAID,
		goal.ID,
		goal.TargetType.String(),
		goal.RewardDescription,
	)

	s.notificationSvc.QueueGoalCompleted(
		connection.UserBID,
		goal.ID,
		goal.TargetType.String(),
		goal.RewardDescription,
	)
}
