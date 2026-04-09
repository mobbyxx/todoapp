package service

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

type mockConflictRepository struct {
	conflicts map[uuid.UUID]*domain.SyncConflictRecord
}

func newMockConflictRepository() *mockConflictRepository {
	return &mockConflictRepository{
		conflicts: make(map[uuid.UUID]*domain.SyncConflictRecord),
	}
}

func (m *mockConflictRepository) GetByID(id uuid.UUID) (*domain.SyncConflictRecord, error) {
	if conflict, ok := m.conflicts[id]; ok {
		return conflict, nil
	}
	return nil, nil
}

func (m *mockConflictRepository) GetUnresolvedByUser(userID uuid.UUID) ([]*domain.SyncConflictRecord, error) {
	var result []*domain.SyncConflictRecord
	for _, conflict := range m.conflicts {
		if conflict.UserID == userID && conflict.Status == string(domain.ConflictStatusPending) {
			result = append(result, conflict)
		}
	}
	return result, nil
}

func (m *mockConflictRepository) UpdateResolution(conflictID uuid.UUID, resolution *domain.ConflictResolution, resolvedBy uuid.UUID) error {
	if conflict, ok := m.conflicts[conflictID]; ok {
		now := time.Now()
		conflict.Status = string(domain.ConflictStatusResolved)
		conflict.ResolvedData = resolution.ResolvedData
		conflict.ResolutionStrategy = string(resolution.Strategy)
		conflict.ResolvedBy = &resolvedBy
		conflict.ResolvedAt = &now
		conflict.UpdatedAt = now
	}
	return nil
}

func (m *mockConflictRepository) RecordConflict(conflict *domain.SyncConflictRecord) error {
	if conflict.ID == uuid.Nil {
		conflict.ID = uuid.New()
	}
	m.conflicts[conflict.ID] = conflict
	return nil
}

func setupTestConflictService(t *testing.T) (domain.ConflictService, *mockConflictRepository, *mockTodoRepository, *mockSyncRepository) {
	conflictRepo := newMockConflictRepository()
	todoRepo := newMockTodoRepository()
	syncRepo := newMockSyncRepository()
	service := NewConflictService(conflictRepo, todoRepo, syncRepo)
	return service, conflictRepo, todoRepo, syncRepo
}

func TestConflictService_DetectConflicts_NoConflicts(t *testing.T) {
	service, _, todoRepo, _ := setupTestConflictService(t)

	userID := uuid.New()
	todoID := uuid.New()

	todo := &domain.Todo{
		ID:        todoID,
		Title:     "Test Todo",
		CreatedBy: userID,
		Status:    domain.TodoStatusPending,
		Priority:  domain.TodoPriorityMedium,
		Version:   1,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	todoRepo.todos[todoID] = todo

	changes := domain.ChangeSet{
		Updated: []*domain.TodoChange{
			{
				ID:        todoID,
				Title:     "Test Todo",
				Status:    domain.TodoStatusPending,
				Priority:  domain.TodoPriorityMedium,
				CreatedBy: userID,
				Version:   2,
				Timestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
			},
		},
	}

	conflicts, err := service.DetectConflicts(userID, changes)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("Expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestConflictService_DetectConflicts_WithConflicts(t *testing.T) {
	service, _, todoRepo, _ := setupTestConflictService(t)

	userID := uuid.New()
	todoID := uuid.New()

	todo := &domain.Todo{
		ID:        todoID,
		Title:     "Server Title",
		CreatedBy: userID,
		Status:    domain.TodoStatusInProgress,
		Priority:  domain.TodoPriorityHigh,
		Version:   2,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}
	todoRepo.todos[todoID] = todo

	changes := domain.ChangeSet{
		Updated: []*domain.TodoChange{
			{
				ID:        todoID,
				Title:     "Client Title",
				Status:    domain.TodoStatusPending,
				Priority:  domain.TodoPriorityMedium,
				CreatedBy: userID,
				Version:   1,
				Timestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
			},
		},
	}

	conflicts, err := service.DetectConflicts(userID, changes)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(conflicts) != 1 {
		t.Fatalf("Expected 1 conflict, got %d", len(conflicts))
	}

	if conflicts[0].EntityID != todoID {
		t.Errorf("Expected conflict for todo %s, got %s", todoID, conflicts[0].EntityID)
	}

	if conflicts[0].ConflictType != "both_modified" {
		t.Errorf("Expected conflict type 'both_modified', got %s", conflicts[0].ConflictType)
	}
}

func TestConflictService_ResolveConflict_LastWriteWins_LocalWins(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	now := time.Now()
	localTime := now.UnixMilli()
	remoteTime := now.Add(-time.Hour).UnixMilli()

	local := &domain.TodoChange{
		ID:        uuid.New(),
		Title:     "Local Title",
		Status:    domain.TodoStatusCompleted,
		Priority:  domain.TodoPriorityHigh,
		Timestamp: domain.HLC{PhysicalTime: localTime},
	}

	remote := &domain.TodoChange{
		ID:        local.ID,
		Title:     "Remote Title",
		Status:    domain.TodoStatusPending,
		Priority:  domain.TodoPriorityMedium,
		Timestamp: domain.HLC{PhysicalTime: remoteTime},
	}

	resolved, err := service.ResolveConflict("todo", local, remote, domain.ResolutionStrategyLastWriteWins)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resolved.Title != "Local Title" {
		t.Errorf("Expected 'Local Title', got '%s'", resolved.Title)
	}

	if resolved.Status != domain.TodoStatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", resolved.Status)
	}
}

func TestConflictService_ResolveConflict_LastWriteWins_RemoteWins(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	now := time.Now()
	localTime := now.Add(-time.Hour).UnixMilli()
	remoteTime := now.UnixMilli()

	local := &domain.TodoChange{
		ID:        uuid.New(),
		Title:     "Local Title",
		Status:    domain.TodoStatusPending,
		Priority:  domain.TodoPriorityLow,
		Timestamp: domain.HLC{PhysicalTime: localTime},
	}

	remote := &domain.TodoChange{
		ID:        local.ID,
		Title:     "Remote Title",
		Status:    domain.TodoStatusInProgress,
		Priority:  domain.TodoPriorityHigh,
		Timestamp: domain.HLC{PhysicalTime: remoteTime},
	}

	resolved, err := service.ResolveConflict("todo", local, remote, domain.ResolutionStrategyLastWriteWins)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resolved.Title != "Remote Title" {
		t.Errorf("Expected 'Remote Title', got '%s'", resolved.Title)
	}

	if resolved.Status != domain.TodoStatusInProgress {
		t.Errorf("Expected status 'in_progress', got '%s'", resolved.Status)
	}
}

func TestConflictService_ResolveConflict_MaxWins_LocalHigher(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	local := &domain.TodoChange{
		ID:       uuid.New(),
		Title:    "Local",
		Priority: domain.TodoPriorityUrgent,
	}

	remote := &domain.TodoChange{
		ID:       local.ID,
		Title:    "Remote",
		Priority: domain.TodoPriorityMedium,
	}

	resolved, err := service.ResolveConflict("todo", local, remote, domain.ResolutionStrategyMaxWins)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resolved.Priority != domain.TodoPriorityUrgent {
		t.Errorf("Expected priority 'urgent', got '%s'", resolved.Priority)
	}
}

func TestConflictService_ResolveConflict_MaxWins_RemoteHigher(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	local := &domain.TodoChange{
		ID:       uuid.New(),
		Title:    "Local",
		Priority: domain.TodoPriorityLow,
	}

	remote := &domain.TodoChange{
		ID:       local.ID,
		Title:    "Remote",
		Priority: domain.TodoPriorityHigh,
	}

	resolved, err := service.ResolveConflict("todo", local, remote, domain.ResolutionStrategyMaxWins)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resolved.Priority != domain.TodoPriorityHigh {
		t.Errorf("Expected priority 'high', got '%s'", resolved.Priority)
	}
}

func TestConflictService_ResolveConflict_MaxWins_EqualPriority(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	local := &domain.TodoChange{
		ID:       uuid.New(),
		Title:    "Local",
		Priority: domain.TodoPriorityMedium,
	}

	remote := &domain.TodoChange{
		ID:       local.ID,
		Title:    "Remote",
		Priority: domain.TodoPriorityMedium,
	}

	resolved, err := service.ResolveConflict("todo", local, remote, domain.ResolutionStrategyMaxWins)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resolved.Priority != domain.TodoPriorityMedium {
		t.Errorf("Expected priority 'medium', got '%s'", resolved.Priority)
	}
}

func TestConflictService_ResolveConflict_Merge_LaterDate(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	earlierDate := time.Now().Add(-48 * time.Hour)
	laterDate := time.Now().Add(-24 * time.Hour)

	local := &domain.TodoChange{
		ID:      uuid.New(),
		Title:   "Local",
		DueDate: &earlierDate,
	}

	remote := &domain.TodoChange{
		ID:      local.ID,
		Title:   "Remote",
		DueDate: &laterDate,
	}

	resolved, err := service.ResolveConflict("todo", local, remote, domain.ResolutionStrategyMerge)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resolved.DueDate == nil || !resolved.DueDate.Equal(laterDate) {
		t.Errorf("Expected due date %v, got %v", laterDate, resolved.DueDate)
	}
}

func TestConflictService_ResolveConflict_Merge_LocalHasDate(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	localDate := time.Now().Add(24 * time.Hour)

	local := &domain.TodoChange{
		ID:      uuid.New(),
		Title:   "Local",
		DueDate: &localDate,
	}

	remote := &domain.TodoChange{
		ID:      local.ID,
		Title:   "Remote",
		DueDate: nil,
	}

	resolved, err := service.ResolveConflict("todo", local, remote, domain.ResolutionStrategyMerge)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resolved.DueDate == nil || !resolved.DueDate.Equal(localDate) {
		t.Errorf("Expected due date %v, got %v", localDate, resolved.DueDate)
	}
}

func TestConflictService_ResolveConflict_Merge_RemoteHasDate(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	remoteDate := time.Now().Add(48 * time.Hour)

	local := &domain.TodoChange{
		ID:      uuid.New(),
		Title:   "Local",
		DueDate: nil,
	}

	remote := &domain.TodoChange{
		ID:      local.ID,
		Title:   "Remote",
		DueDate: &remoteDate,
	}

	resolved, err := service.ResolveConflict("todo", local, remote, domain.ResolutionStrategyMerge)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resolved.DueDate == nil || !resolved.DueDate.Equal(remoteDate) {
		t.Errorf("Expected due date %v, got %v", remoteDate, resolved.DueDate)
	}
}

func TestConflictService_ResolveConflict_Prompt_ReturnsError(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	local := &domain.TodoChange{
		ID:    uuid.New(),
		Title: "Local",
	}

	remote := &domain.TodoChange{
		ID:    local.ID,
		Title: "Remote",
	}

	resolved, err := service.ResolveConflict("todo", local, remote, domain.ResolutionStrategyPrompt)

	if err != domain.ErrConflictDetected {
		t.Errorf("Expected ErrConflictDetected, got %v", err)
	}

	if resolved != nil {
		t.Error("Expected nil resolved change for prompt strategy")
	}
}

func TestConflictService_ResolveConflict_UnsupportedEntityType(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	local := &domain.TodoChange{ID: uuid.New()}
	remote := &domain.TodoChange{ID: local.ID}

	_, err := service.ResolveConflict("user", local, remote, domain.ResolutionStrategyLastWriteWins)

	if err == nil {
		t.Error("Expected error for unsupported entity type")
	}
}

func TestConflictService_GetUnresolvedConflicts(t *testing.T) {
	service, conflictRepo, _, _ := setupTestConflictService(t)

	userID := uuid.New()
	conflictID := uuid.New()

	conflict := &domain.SyncConflictRecord{
		ID:              conflictID,
		UserID:          userID,
		EntityType:      "todo",
		EntityID:        uuid.New(),
		LocalVersion:    1,
		RemoteVersion:   2,
		Status:          string(domain.ConflictStatusPending),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		LocalData:       []byte(`{"title":"local"}`),
		RemoteData:      []byte(`{"title":"remote"}`),
		ClientTimestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
		ServerTimestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
	}
	conflictRepo.conflicts[conflictID] = conflict

	conflicts, err := service.GetUnresolvedConflicts(userID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(conflicts) != 1 {
		t.Fatalf("Expected 1 conflict, got %d", len(conflicts))
	}

	if conflicts[0].ID != conflictID {
		t.Errorf("Expected conflict ID %s, got %s", conflictID, conflicts[0].ID)
	}
}

func TestConflictService_ResolveConflictManually_Local(t *testing.T) {
	service, conflictRepo, todoRepo, _ := setupTestConflictService(t)

	userID := uuid.New()
	todoID := uuid.New()
	conflictID := uuid.New()

	todo := &domain.Todo{
		ID:        todoID,
		Title:     "Server Title",
		CreatedBy: userID,
		Status:    domain.TodoStatusPending,
		Priority:  domain.TodoPriorityMedium,
		Version:   2,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}
	todoRepo.todos[todoID] = todo

	localData, _ := json.Marshal(&domain.TodoChange{
		ID:       todoID,
		Title:    "Local Title",
		Status:   domain.TodoStatusCompleted,
		Priority: domain.TodoPriorityHigh,
		Version:  1,
	})

	remoteData, _ := json.Marshal(&domain.TodoChange{
		ID:       todoID,
		Title:    "Server Title",
		Status:   domain.TodoStatusPending,
		Priority: domain.TodoPriorityMedium,
		Version:  2,
	})

	conflict := &domain.SyncConflictRecord{
		ID:              conflictID,
		UserID:          userID,
		EntityType:      "todo",
		EntityID:        todoID,
		LocalVersion:    1,
		RemoteVersion:   2,
		LocalData:       localData,
		RemoteData:      remoteData,
		Status:          string(domain.ConflictStatusPending),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ClientTimestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
		ServerTimestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
	}
	conflictRepo.conflicts[conflictID] = conflict

	resolution := domain.ConflictResolution{
		ConflictID:   conflictID,
		Strategy:     "local",
		ResolvedData: localData,
	}

	err := service.ResolveConflictManually(conflictID, resolution, userID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updatedConflict := conflictRepo.conflicts[conflictID]
	if updatedConflict.Status != string(domain.ConflictStatusResolved) {
		t.Errorf("Expected status 'resolved', got '%s'", updatedConflict.Status)
	}

	updatedTodo := todoRepo.todos[todoID]
	if updatedTodo.Title != "Local Title" {
		t.Errorf("Expected todo title 'Local Title', got '%s'", updatedTodo.Title)
	}

	if updatedTodo.Version != 3 {
		t.Errorf("Expected version 3, got %d", updatedTodo.Version)
	}
}

func TestConflictService_ResolveConflictManually_Remote(t *testing.T) {
	service, conflictRepo, todoRepo, _ := setupTestConflictService(t)

	userID := uuid.New()
	todoID := uuid.New()
	conflictID := uuid.New()

	todo := &domain.Todo{
		ID:        todoID,
		Title:     "Server Title",
		CreatedBy: userID,
		Status:    domain.TodoStatusPending,
		Priority:  domain.TodoPriorityMedium,
		Version:   2,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}
	todoRepo.todos[todoID] = todo

	localData, _ := json.Marshal(&domain.TodoChange{
		ID:       todoID,
		Title:    "Local Title",
		Status:   domain.TodoStatusCompleted,
		Priority: domain.TodoPriorityHigh,
		Version:  1,
	})

	remoteData, _ := json.Marshal(&domain.TodoChange{
		ID:       todoID,
		Title:    "Server Title",
		Status:   domain.TodoStatusPending,
		Priority: domain.TodoPriorityMedium,
		Version:  2,
	})

	conflict := &domain.SyncConflictRecord{
		ID:              conflictID,
		UserID:          userID,
		EntityType:      "todo",
		EntityID:        todoID,
		LocalVersion:    1,
		RemoteVersion:   2,
		LocalData:       localData,
		RemoteData:      remoteData,
		Status:          string(domain.ConflictStatusPending),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ClientTimestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
		ServerTimestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
	}
	conflictRepo.conflicts[conflictID] = conflict

	resolution := domain.ConflictResolution{
		ConflictID:   conflictID,
		Strategy:     "remote",
		ResolvedData: remoteData,
	}

	err := service.ResolveConflictManually(conflictID, resolution, userID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updatedTodo := todoRepo.todos[todoID]
	if updatedTodo.Title != "Server Title" {
		t.Errorf("Expected todo title 'Server Title', got '%s'", updatedTodo.Title)
	}
}

func TestConflictService_ResolveConflictManually_AlreadyResolved(t *testing.T) {
	service, conflictRepo, _, _ := setupTestConflictService(t)

	userID := uuid.New()
	conflictID := uuid.New()

	conflict := &domain.SyncConflictRecord{
		ID:              conflictID,
		UserID:          userID,
		EntityType:      "todo",
		EntityID:        uuid.New(),
		LocalVersion:    1,
		RemoteVersion:   2,
		Status:          string(domain.ConflictStatusResolved),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		LocalData:       []byte(`{"title":"local"}`),
		RemoteData:      []byte(`{"title":"remote"}`),
		ClientTimestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
		ServerTimestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
	}
	conflictRepo.conflicts[conflictID] = conflict

	resolution := domain.ConflictResolution{
		ConflictID: conflictID,
		Strategy:   "local",
	}

	err := service.ResolveConflictManually(conflictID, resolution, userID)

	if err == nil {
		t.Error("Expected error for already resolved conflict")
	}
}

func TestConflictService_ResolveConflictManually_WrongUser(t *testing.T) {
	service, conflictRepo, _, _ := setupTestConflictService(t)

	userID := uuid.New()
	wrongUserID := uuid.New()
	conflictID := uuid.New()

	conflict := &domain.SyncConflictRecord{
		ID:              conflictID,
		UserID:          userID,
		EntityType:      "todo",
		EntityID:        uuid.New(),
		LocalVersion:    1,
		RemoteVersion:   2,
		Status:          string(domain.ConflictStatusPending),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		LocalData:       []byte(`{"title":"local"}`),
		RemoteData:      []byte(`{"title":"remote"}`),
		ClientTimestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
		ServerTimestamp: domain.HLC{PhysicalTime: time.Now().UnixMilli()},
	}
	conflictRepo.conflicts[conflictID] = conflict

	resolution := domain.ConflictResolution{
		ConflictID: conflictID,
		Strategy:   "local",
	}

	err := service.ResolveConflictManually(conflictID, resolution, wrongUserID)

	if err != domain.ErrUnauthorizedAction {
		t.Errorf("Expected ErrUnauthorizedAction, got %v", err)
	}
}

func TestConflictService_ResolveWithFieldStrategies_AllAutoResolve(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	now := time.Now()
	laterDate := now.Add(48 * time.Hour)

	local := &domain.TodoChange{
		ID:          uuid.New(),
		Title:       "Local Title",
		Description: "Local Desc",
		Status:      domain.TodoStatusPending,
		Priority:    domain.TodoPriorityMedium,
		DueDate:     &now,
		Timestamp:   domain.HLC{PhysicalTime: now.UnixMilli()},
	}

	remote := &domain.TodoChange{
		ID:          local.ID,
		Title:       "Remote Title",
		Description: "Remote Desc",
		Status:      domain.TodoStatusInProgress,
		Priority:    domain.TodoPriorityLow,
		DueDate:     &laterDate,
		Timestamp:   domain.HLC{PhysicalTime: now.Add(-time.Hour).UnixMilli()},
	}

	resolved, promptFields, err := service.ResolveWithFieldStrategies(local, remote)

	if !errors.Is(err, domain.ErrConflictDetected) {
		t.Fatalf("Expected ErrConflictDetected, got %v", err)
	}

	if len(promptFields) != 2 {
		t.Errorf("Expected 2 prompt fields (title, description), got %d: %v", len(promptFields), promptFields)
	}

	if resolved.Status != domain.TodoStatusPending {
		t.Errorf("Expected status 'pending' (last-write-wins), got '%s'", resolved.Status)
	}

	if resolved.Priority != domain.TodoPriorityMedium {
		t.Errorf("Expected priority 'medium' (max-wins), got '%s'", resolved.Priority)
	}

	if resolved.DueDate == nil || !resolved.DueDate.Equal(laterDate) {
		t.Errorf("Expected due date %v (merge), got %v", laterDate, resolved.DueDate)
	}
}

func TestConflictService_ResolveWithFieldStrategies_NoConflicts(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	now := time.Now()

	local := &domain.TodoChange{
		ID:          uuid.New(),
		Title:       "Same Title",
		Description: "Same Desc",
		Status:      domain.TodoStatusPending,
		Priority:    domain.TodoPriorityMedium,
		DueDate:     &now,
	}

	remote := &domain.TodoChange{
		ID:          local.ID,
		Title:       "Same Title",
		Description: "Same Desc",
		Status:      domain.TodoStatusPending,
		Priority:    domain.TodoPriorityMedium,
		DueDate:     &now,
	}

	resolved, promptFields, err := service.ResolveWithFieldStrategies(local, remote)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(promptFields) != 0 {
		t.Errorf("Expected 0 prompt fields, got %d", len(promptFields))
	}

	if resolved == nil {
		t.Error("Expected non-nil resolved change")
	}
}

func TestConflictService_fieldsDiffer(t *testing.T) {
	service, _, _, _ := setupTestConflictService(t)

	now := time.Now()

	tests := []struct {
		name     string
		local    *domain.TodoChange
		remote   *domain.TodoChange
		expected bool
	}{
		{
			name:     "identical",
			local:    &domain.TodoChange{ID: uuid.New(), Title: "Test", Status: domain.TodoStatusPending, Priority: domain.TodoPriorityMedium},
			remote:   &domain.TodoChange{ID: uuid.New(), Title: "Test", Status: domain.TodoStatusPending, Priority: domain.TodoPriorityMedium},
			expected: false,
		},
		{
			name:     "different title",
			local:    &domain.TodoChange{ID: uuid.New(), Title: "Local"},
			remote:   &domain.TodoChange{ID: uuid.New(), Title: "Remote"},
			expected: true,
		},
		{
			name:     "different status",
			local:    &domain.TodoChange{ID: uuid.New(), Status: domain.TodoStatusPending},
			remote:   &domain.TodoChange{ID: uuid.New(), Status: domain.TodoStatusCompleted},
			expected: true,
		},
		{
			name:     "different priority",
			local:    &domain.TodoChange{ID: uuid.New(), Priority: domain.TodoPriorityLow},
			remote:   &domain.TodoChange{ID: uuid.New(), Priority: domain.TodoPriorityHigh},
			expected: true,
		},
		{
			name:     "different due date",
			local:    &domain.TodoChange{ID: uuid.New(), DueDate: &now},
			remote:   &domain.TodoChange{ID: uuid.New(), DueDate: nil},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.(*conflictService).fieldsDiffer(tt.local, tt.remote)
			if result != tt.expected {
				t.Errorf("fieldsDiffer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPriorityOrder(t *testing.T) {
	tests := []struct {
		priority domain.TodoPriority
		expected int
	}{
		{domain.TodoPriorityLow, 1},
		{domain.TodoPriorityMedium, 2},
		{domain.TodoPriorityHigh, 3},
		{domain.TodoPriorityUrgent, 4},
	}

	for _, tt := range tests {
		if priorityOrder[tt.priority] != tt.expected {
			t.Errorf("priorityOrder[%s] = %d, want %d", tt.priority, priorityOrder[tt.priority], tt.expected)
		}
	}
}
