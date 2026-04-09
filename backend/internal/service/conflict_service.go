package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

var priorityOrder = map[domain.TodoPriority]int{
	domain.TodoPriorityLow:    1,
	domain.TodoPriorityMedium: 2,
	domain.TodoPriorityHigh:   3,
	domain.TodoPriorityUrgent: 4,
}

type conflictService struct {
	conflictRepo domain.ConflictRepository
	todoRepo     domain.TodoRepository
	syncRepo     domain.SyncRepository
}

func NewConflictService(
	conflictRepo domain.ConflictRepository,
	todoRepo domain.TodoRepository,
	syncRepo domain.SyncRepository,
) domain.ConflictService {
	return &conflictService{
		conflictRepo: conflictRepo,
		todoRepo:     todoRepo,
		syncRepo:     syncRepo,
	}
}

func (s *conflictService) DetectConflicts(userID uuid.UUID, changes domain.ChangeSet) ([]*domain.SyncConflict, error) {
	conflicts := []*domain.SyncConflict{}

	for _, updated := range changes.Updated {
		existing, err := s.todoRepo.GetByID(updated.ID)
		if err != nil {
			if errors.Is(err, domain.ErrTodoNotFound) {
				continue
			}
			return nil, err
		}

		if existing != nil && s.hasConflict(updated, existing) {
			conflict := &domain.SyncConflict{
				EntityType:    "todo",
				EntityID:      updated.ID,
				LocalVersion:  updated,
				ServerVersion: domain.TodoChangeFromTodo(existing),
				ConflictType:  "both_modified",
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts, nil
}

func (s *conflictService) hasConflict(local *domain.TodoChange, remote *domain.Todo) bool {
	if remote.Version >= local.Version {
		return true
	}

	remoteChange := domain.TodoChangeFromTodo(remote)
	return s.fieldsDiffer(local, remoteChange)
}

func (s *conflictService) fieldsDiffer(local, remote *domain.TodoChange) bool {
	if local.Title != remote.Title {
		return true
	}
	if local.Description != remote.Description {
		return true
	}
	if local.Status != remote.Status {
		return true
	}
	if local.Priority != remote.Priority {
		return true
	}
	if !s.ptrUUIDEqual(local.AssignedTo, remote.AssignedTo) {
		return true
	}
	if !s.ptrTimeEqual(local.DueDate, remote.DueDate) {
		return true
	}
	return false
}

func (s *conflictService) ptrUUIDEqual(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func (s *conflictService) ptrTimeEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

func (s *conflictService) ResolveConflict(
	entityType string,
	local, remote *domain.TodoChange,
	strategy domain.ResolutionStrategy,
) (*domain.TodoChange, error) {
	if entityType != "todo" {
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}

	switch strategy {
	case domain.ResolutionStrategyLastWriteWins:
		return s.resolveLastWriteWins(local, remote)
	case domain.ResolutionStrategyMaxWins:
		return s.resolveMaxWins(local, remote)
	case domain.ResolutionStrategyMerge:
		return s.resolveMerge(local, remote)
	case domain.ResolutionStrategyPrompt:
		return nil, domain.ErrConflictDetected
	default:
		return nil, fmt.Errorf("unknown resolution strategy: %s", strategy)
	}
}

func (s *conflictService) resolveLastWriteWins(local, remote *domain.TodoChange) (*domain.TodoChange, error) {
	if local.Timestamp.Compare(remote.Timestamp) >= 0 {
		return local, nil
	}
	return remote, nil
}

func (s *conflictService) resolveMaxWins(local, remote *domain.TodoChange) (*domain.TodoChange, error) {
	localPriority := priorityOrder[local.Priority]
	remotePriority := priorityOrder[remote.Priority]

	if localPriority >= remotePriority {
		return local, nil
	}
	return remote, nil
}

func (s *conflictService) resolveMerge(local, remote *domain.TodoChange) (*domain.TodoChange, error) {
	merged := *local

	if remote.DueDate != nil {
		if local.DueDate == nil {
			merged.DueDate = remote.DueDate
		} else if remote.DueDate.After(*local.DueDate) {
			merged.DueDate = remote.DueDate
		}
	}

	return &merged, nil
}

func (s *conflictService) GetUnresolvedConflicts(userID uuid.UUID) ([]*domain.SyncConflictRecord, error) {
	return s.conflictRepo.GetUnresolvedByUser(userID)
}

func (s *conflictService) ResolveConflictManually(
	conflictID uuid.UUID,
	resolution domain.ConflictResolution,
	userID uuid.UUID,
) error {
	conflict, err := s.conflictRepo.GetByID(conflictID)
	if err != nil {
		return err
	}

	if conflict == nil {
		return errors.New("conflict not found")
	}

	if conflict.Status != string(domain.ConflictStatusPending) {
		return errors.New("conflict is already resolved")
	}

	if conflict.UserID != userID {
		return domain.ErrUnauthorizedAction
	}

	var resolvedTodo *domain.TodoChange
	switch resolution.Strategy {
	case "local":
		resolvedTodo = &domain.TodoChange{}
		if err := json.Unmarshal(conflict.LocalData, resolvedTodo); err != nil {
			return err
		}
	case "remote":
		resolvedTodo = &domain.TodoChange{}
		if err := json.Unmarshal(conflict.RemoteData, resolvedTodo); err != nil {
			return err
		}
	case "custom":
		resolvedTodo = &domain.TodoChange{}
		if err := json.Unmarshal(resolution.ResolvedData, resolvedTodo); err != nil {
			return err
		}
	case "merge":
		local := &domain.TodoChange{}
		remote := &domain.TodoChange{}
		if err := json.Unmarshal(conflict.LocalData, local); err != nil {
			return err
		}
		if err := json.Unmarshal(conflict.RemoteData, remote); err != nil {
			return err
		}
		merged, err := s.resolveMerge(local, remote)
		if err != nil {
			return err
		}
		if resolution.ResolvedData != nil {
			if err := json.Unmarshal(resolution.ResolvedData, merged); err != nil {
				return err
			}
		}
		resolvedTodo = merged
	default:
		return fmt.Errorf("invalid resolution strategy: %s", resolution.Strategy)
	}

	todo := resolvedTodo.ToTodo()
	todo.Version = conflict.RemoteVersion

	if err := s.todoRepo.Update(todo); err != nil {
		return err
	}

	return s.conflictRepo.UpdateResolution(conflictID, &resolution, userID)
}

func (s *conflictService) ResolveWithFieldStrategies(
	local, remote *domain.TodoChange,
) (*domain.TodoChange, []string, error) {
	merged := *local
	promptFields := []string{}

	for field, strategy := range domain.FieldConflictResolution {
		switch field {
		case "title":
			if local.Title != remote.Title {
				if strategy == domain.ResolutionStrategyPrompt {
					promptFields = append(promptFields, field)
				} else {
					resolved, _ := s.ResolveConflict("todo", local, remote, strategy)
					if resolved != nil {
						merged.Title = resolved.Title
					}
				}
			}
		case "description":
			if local.Description != remote.Description {
				if strategy == domain.ResolutionStrategyPrompt {
					promptFields = append(promptFields, field)
				} else {
					resolved, _ := s.ResolveConflict("todo", local, remote, strategy)
					if resolved != nil {
						merged.Description = resolved.Description
					}
				}
			}
		case "status":
			if local.Status != remote.Status {
				resolved, _ := s.resolveLastWriteWins(local, remote)
				if resolved != nil {
					merged.Status = resolved.Status
				}
			}
		case "due_date":
			if !s.ptrTimeEqual(local.DueDate, remote.DueDate) {
				resolved, _ := s.resolveMerge(local, remote)
				if resolved != nil {
					merged.DueDate = resolved.DueDate
				}
			}
		case "priority":
			if local.Priority != remote.Priority {
				resolved, _ := s.resolveMaxWins(local, remote)
				if resolved != nil {
					merged.Priority = resolved.Priority
				}
			}
		case "assigned_to":
			if !s.ptrUUIDEqual(local.AssignedTo, remote.AssignedTo) {
				promptFields = append(promptFields, field)
			}
		}
	}

	if len(promptFields) > 0 {
		return &merged, promptFields, domain.ErrConflictDetected
	}

	return &merged, nil, nil
}
