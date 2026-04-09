package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

type syncService struct {
	syncRepo       domain.SyncRepository
	todoRepo       domain.TodoRepository
	connectionRepo domain.ConnectionRepository
	hlcClock       *domain.HLCClock
}

func NewSyncService(
	syncRepo domain.SyncRepository,
	todoRepo domain.TodoRepository,
	connectionRepo domain.ConnectionRepository,
) domain.SyncService {
	return &syncService{
		syncRepo:       syncRepo,
		todoRepo:       todoRepo,
		connectionRepo: connectionRepo,
		hlcClock:       domain.NewHLCClock(),
	}
}

func (s *syncService) PullChanges(userID uuid.UUID, lastPulledAt domain.HLC) (*domain.ChangeSet, error) {
	changes, err := s.syncRepo.GetChangesSince(userID, lastPulledAt)
	if err != nil {
		return nil, err
	}

	return changes, nil
}

func (s *syncService) PushChanges(userID uuid.UUID, changes domain.ChangeSet) ([]*domain.SyncConflict, error) {
	if err := changes.IsValid(); err != nil {
		return nil, domain.ErrInvalidChangeSet
	}

	timestamp := s.hlcClock.Now()

	conflicts, err := s.syncRepo.ApplyChanges(userID, changes, timestamp)
	if err != nil {
		return nil, err
	}

	return conflicts, nil
}

func (s *syncService) GetLastSync(userID uuid.UUID) (domain.HLC, error) {
	record, err := s.syncRepo.GetLastSync(userID)
	if err != nil {
		return domain.HLC{}, err
	}

	if record == nil {
		return domain.HLC{PhysicalTime: 0, LogicalTime: 0}, nil
	}

	return record.LastSyncedAt, nil
}

func (s *syncService) UpdateLastSync(userID uuid.UUID, timestamp domain.HLC) error {
	return s.syncRepo.UpdateLastSync(userID, timestamp)
}

func (s *syncService) Sync(userID uuid.UUID, req domain.SyncRequest) (*domain.SyncResponse, error) {
	if req.LastPulledAt.IsZero() {
		req.LastPulledAt = domain.HLC{PhysicalTime: 0, LogicalTime: 0}
	}

	syncTimestamp := s.hlcClock.Now()

	serverChanges, err := s.PullChanges(userID, req.LastPulledAt)
	if err != nil {
		return nil, err
	}

	var conflicts []*domain.SyncConflict
	if !req.Changes.IsEmpty() {
		conflicts, err = s.PushChanges(userID, req.Changes)
		if err != nil {
			return nil, err
		}
	}

	if err := s.UpdateLastSync(userID, syncTimestamp); err != nil {
		return nil, err
	}

	status := domain.SyncStatusCompleted
	if len(conflicts) > 0 {
		status = domain.SyncStatusPending
	}

	return &domain.SyncResponse{
		Timestamp:  syncTimestamp,
		Changes:    *serverChanges,
		Conflicts:  conflicts,
		Status:     status,
		ServerTime: time.Now().UnixMilli(),
	}, nil
}
