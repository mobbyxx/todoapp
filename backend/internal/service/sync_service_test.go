package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

type mockSyncRepository struct {
	lastSync  map[uuid.UUID]*domain.SyncRecord
	changes   map[uuid.UUID][]*domain.Todo
	conflicts []*domain.SyncConflictRecord
}

func newMockSyncRepository() *mockSyncRepository {
	return &mockSyncRepository{
		lastSync:  make(map[uuid.UUID]*domain.SyncRecord),
		changes:   make(map[uuid.UUID][]*domain.Todo),
		conflicts: []*domain.SyncConflictRecord{},
	}
}

func (m *mockSyncRepository) GetLastSync(userID uuid.UUID) (*domain.SyncRecord, error) {
	if record, ok := m.lastSync[userID]; ok {
		return record, nil
	}
	return nil, nil
}

func (m *mockSyncRepository) UpdateLastSync(userID uuid.UUID, timestamp domain.HLC) error {
	m.lastSync[userID] = &domain.SyncRecord{
		UserID:       userID,
		LastSyncedAt: timestamp,
		Status:       domain.SyncStatusCompleted,
		UpdatedAt:    time.Now(),
	}
	return nil
}

func (m *mockSyncRepository) GetChangesSince(userID uuid.UUID, since domain.HLC) (*domain.ChangeSet, error) {
	changes := &domain.ChangeSet{
		Created: []*domain.TodoChange{},
		Updated: []*domain.TodoChange{},
		Deleted: []*domain.TodoChange{},
	}

	for _, todo := range m.changes[userID] {
		updatedMs := todo.UpdatedAt.UnixMilli()
		if updatedMs > since.PhysicalTime {
			change := domain.TodoChangeFromTodo(todo)
			change.Timestamp = domain.HLC{
				PhysicalTime: updatedMs,
				LogicalTime:  0,
			}
			if todo.DeletedAt != nil {
				change.IsDeleted = true
				changes.Deleted = append(changes.Deleted, change)
			} else if todo.CreatedAt.UnixMilli() == updatedMs {
				changes.Created = append(changes.Created, change)
			} else {
				changes.Updated = append(changes.Updated, change)
			}
		}
	}

	return changes, nil
}

func (m *mockSyncRepository) ApplyChanges(userID uuid.UUID, changes domain.ChangeSet, timestamp domain.HLC) ([]*domain.SyncConflict, error) {
	conflicts := []*domain.SyncConflict{}

	for _, created := range changes.Created {
		m.changes[userID] = append(m.changes[userID], created.ToTodo())
	}

	for _, updated := range changes.Updated {
		todo := updated.ToTodo()
		todo.UpdatedAt = time.Now()
		m.changes[userID] = append(m.changes[userID], todo)
	}

	for _, deleted := range changes.Deleted {
		todo := deleted.ToTodo()
		now := time.Now()
		todo.DeletedAt = &now
		m.changes[userID] = append(m.changes[userID], todo)
	}

	return conflicts, nil
}

func (m *mockSyncRepository) RecordConflict(conflict *domain.SyncConflictRecord) error {
	m.conflicts = append(m.conflicts, conflict)
	return nil
}

func setupTestSyncService(t *testing.T) (domain.SyncService, *mockSyncRepository, *mockTodoRepository, *mockConnectionRepositoryForTodo) {
	syncRepo := newMockSyncRepository()
	todoRepo := newMockTodoRepository()
	connRepo := newMockConnectionRepositoryForTodo()
	service := NewSyncService(syncRepo, todoRepo, connRepo)
	return service, syncRepo, todoRepo, connRepo
}

func TestSyncService_GetLastSync_Initial(t *testing.T) {
	service, _, _, _ := setupTestSyncService(t)

	userID := uuid.New()

	lastSync, err := service.GetLastSync(userID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !lastSync.IsZero() {
		t.Errorf("Expected zero HLC for initial sync, got %v", lastSync)
	}
}

func TestSyncService_UpdateLastSync(t *testing.T) {
	service, _, _, _ := setupTestSyncService(t)

	userID := uuid.New()
	timestamp := domain.HLC{
		PhysicalTime: time.Now().UnixMilli(),
		LogicalTime:  1,
	}

	err := service.UpdateLastSync(userID, timestamp)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	lastSync, err := service.GetLastSync(userID)
	if err != nil {
		t.Fatalf("Expected no error getting last sync, got %v", err)
	}

	if lastSync.PhysicalTime != timestamp.PhysicalTime {
		t.Errorf("Expected physical time %d, got %d", timestamp.PhysicalTime, lastSync.PhysicalTime)
	}

	if lastSync.LogicalTime != timestamp.LogicalTime {
		t.Errorf("Expected logical time %d, got %d", timestamp.LogicalTime, lastSync.LogicalTime)
	}
}

func TestSyncService_PullChanges_Empty(t *testing.T) {
	service, _, _, _ := setupTestSyncService(t)

	userID := uuid.New()
	lastPulledAt := domain.HLC{PhysicalTime: 0}

	changes, err := service.PullChanges(userID, lastPulledAt)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if changes.TotalChanges() != 0 {
		t.Errorf("Expected 0 changes, got %d", changes.TotalChanges())
	}
}

func TestSyncService_PullChanges_WithChanges(t *testing.T) {
	service, syncRepo, _, _ := setupTestSyncService(t)

	userID := uuid.New()
	now := time.Now()

	todo := &domain.Todo{
		ID:        uuid.New(),
		Title:     "Test Todo",
		CreatedBy: userID,
		Status:    domain.TodoStatusPending,
		Priority:  domain.TodoPriorityMedium,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
	syncRepo.changes[userID] = append(syncRepo.changes[userID], todo)

	lastPulledAt := domain.HLC{PhysicalTime: now.UnixMilli() - 1000}

	changes, err := service.PullChanges(userID, lastPulledAt)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if changes.TotalChanges() != 1 {
		t.Errorf("Expected 1 change, got %d", changes.TotalChanges())
	}

	if len(changes.Created) != 1 {
		t.Errorf("Expected 1 created change, got %d", len(changes.Created))
	}
}

func TestSyncService_PushChanges_Success(t *testing.T) {
	service, _, _, _ := setupTestSyncService(t)

	userID := uuid.New()
	now := time.Now()

	changes := domain.ChangeSet{
		Created: []*domain.TodoChange{
			{
				ID:        uuid.New(),
				Title:     "New Todo",
				Status:    domain.TodoStatusPending,
				Priority:  domain.TodoPriorityHigh,
				CreatedBy: userID,
				Version:   1,
				Timestamp: domain.HLC{PhysicalTime: now.UnixMilli()},
			},
		},
	}

	conflicts, err := service.PushChanges(userID, changes)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("Expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestSyncService_PushChanges_InvalidChangeSet(t *testing.T) {
	service, _, _, _ := setupTestSyncService(t)

	userID := uuid.New()

	changes := domain.ChangeSet{
		Created: []*domain.TodoChange{
			{
				ID:        uuid.Nil,
				Title:     "",
				CreatedBy: uuid.Nil,
			},
		},
	}

	_, err := service.PushChanges(userID, changes)

	if err == nil {
		t.Fatal("Expected error for invalid change set")
	}

	if err != domain.ErrInvalidChangeSet {
		t.Errorf("Expected ErrInvalidChangeSet, got %v", err)
	}
}

func TestSyncService_Sync_Bidirectional(t *testing.T) {
	service, syncRepo, _, _ := setupTestSyncService(t)

	userID := uuid.New()
	now := time.Now()

	serverTodo := &domain.Todo{
		ID:        uuid.New(),
		Title:     "Server Todo",
		CreatedBy: userID,
		Status:    domain.TodoStatusPending,
		Priority:  domain.TodoPriorityMedium,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
		Version:   1,
	}
	syncRepo.changes[userID] = append(syncRepo.changes[userID], serverTodo)

	clientChanges := domain.ChangeSet{
		Created: []*domain.TodoChange{
			{
				ID:        uuid.New(),
				Title:     "Client Todo",
				Status:    domain.TodoStatusPending,
				Priority:  domain.TodoPriorityHigh,
				CreatedBy: userID,
				Version:   1,
				Timestamp: domain.HLC{PhysicalTime: now.UnixMilli()},
			},
		},
	}

	req := domain.SyncRequest{
		LastPulledAt: domain.HLC{PhysicalTime: now.Add(-2 * time.Hour).UnixMilli()},
		Changes:      clientChanges,
	}

	resp, err := service.Sync(userID, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Changes.TotalChanges() != 1 {
		t.Errorf("Expected 1 server change, got %d", resp.Changes.TotalChanges())
	}

	if resp.Status != domain.SyncStatusCompleted {
		t.Errorf("Expected status %s, got %s", domain.SyncStatusCompleted, resp.Status)
	}

	if resp.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if resp.ServerTime == 0 {
		t.Error("Expected non-zero server time")
	}
}

func TestSyncService_Sync_NoChanges(t *testing.T) {
	service, _, _, _ := setupTestSyncService(t)

	userID := uuid.New()

	req := domain.SyncRequest{
		LastPulledAt: domain.HLC{PhysicalTime: 0},
		Changes:      domain.ChangeSet{},
	}

	resp, err := service.Sync(userID, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Changes.TotalChanges() != 0 {
		t.Errorf("Expected 0 changes, got %d", resp.Changes.TotalChanges())
	}

	if resp.Status != domain.SyncStatusCompleted {
		t.Errorf("Expected status %s, got %s", domain.SyncStatusCompleted, resp.Status)
	}
}

func TestHLCClock_Now(t *testing.T) {
	clock := domain.NewHLCClock()

	hlc1 := clock.Now()
	time.Sleep(1 * time.Millisecond)
	hlc2 := clock.Now()

	if hlc1.IsZero() {
		t.Error("Expected non-zero HLC")
	}

	if hlc2.Less(hlc1) {
		t.Error("Second HLC should not be less than first")
	}
}

func TestHLCClock_Update(t *testing.T) {
	clock := domain.NewHLCClock()

	localHLC := clock.Now()

	receivedHLC := domain.HLC{
		PhysicalTime: localHLC.PhysicalTime + 1000,
		LogicalTime:  5,
	}

	newHLC := clock.Update(receivedHLC)

	if newHLC.PhysicalTime < receivedHLC.PhysicalTime {
		t.Errorf("Expected physical time >= %d, got %d", receivedHLC.PhysicalTime, newHLC.PhysicalTime)
	}

	if newHLC.LogicalTime <= receivedHLC.LogicalTime {
		t.Errorf("Expected logical time > %d, got %d", receivedHLC.LogicalTime, newHLC.LogicalTime)
	}
}

func TestHLC_Compare(t *testing.T) {
	tests := []struct {
		h1       domain.HLC
		h2       domain.HLC
		expected int
	}{
		{domain.HLC{PhysicalTime: 1000, LogicalTime: 0}, domain.HLC{PhysicalTime: 2000, LogicalTime: 0}, -1},
		{domain.HLC{PhysicalTime: 2000, LogicalTime: 0}, domain.HLC{PhysicalTime: 1000, LogicalTime: 0}, 1},
		{domain.HLC{PhysicalTime: 1000, LogicalTime: 0}, domain.HLC{PhysicalTime: 1000, LogicalTime: 0}, 0},
		{domain.HLC{PhysicalTime: 1000, LogicalTime: 1}, domain.HLC{PhysicalTime: 1000, LogicalTime: 2}, -1},
		{domain.HLC{PhysicalTime: 1000, LogicalTime: 2}, domain.HLC{PhysicalTime: 1000, LogicalTime: 1}, 1},
	}

	for _, tt := range tests {
		result := tt.h1.Compare(tt.h2)
		if result != tt.expected {
			t.Errorf("Compare(%v, %v) = %d, want %d", tt.h1, tt.h2, result, tt.expected)
		}
	}
}

func TestHLC_Less(t *testing.T) {
	tests := []struct {
		h1       domain.HLC
		h2       domain.HLC
		expected bool
	}{
		{domain.HLC{PhysicalTime: 1000, LogicalTime: 0}, domain.HLC{PhysicalTime: 2000, LogicalTime: 0}, true},
		{domain.HLC{PhysicalTime: 2000, LogicalTime: 0}, domain.HLC{PhysicalTime: 1000, LogicalTime: 0}, false},
		{domain.HLC{PhysicalTime: 1000, LogicalTime: 0}, domain.HLC{PhysicalTime: 1000, LogicalTime: 0}, false},
		{domain.HLC{PhysicalTime: 1000, LogicalTime: 1}, domain.HLC{PhysicalTime: 1000, LogicalTime: 2}, true},
	}

	for _, tt := range tests {
		result := tt.h1.Less(tt.h2)
		if result != tt.expected {
			t.Errorf("Less(%v, %v) = %v, want %v", tt.h1, tt.h2, result, tt.expected)
		}
	}
}

func TestChangeSet_IsValid(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name    string
		changes domain.ChangeSet
		wantErr bool
	}{
		{
			name: "valid changeset",
			changes: domain.ChangeSet{
				Created: []*domain.TodoChange{
					{ID: uuid.New(), Title: "Test", CreatedBy: userID},
				},
			},
			wantErr: false,
		},
		{
			name: "missing title",
			changes: domain.ChangeSet{
				Created: []*domain.TodoChange{
					{ID: uuid.New(), Title: "", CreatedBy: userID},
				},
			},
			wantErr: true,
		},
		{
			name: "missing ID",
			changes: domain.ChangeSet{
				Created: []*domain.TodoChange{
					{ID: uuid.Nil, Title: "Test", CreatedBy: userID},
				},
			},
			wantErr: true,
		},
		{
			name: "missing created_by",
			changes: domain.ChangeSet{
				Created: []*domain.TodoChange{
					{ID: uuid.New(), Title: "Test", CreatedBy: uuid.Nil},
				},
			},
			wantErr: true,
		},
		{
			name: "empty changeset",
			changes: domain.ChangeSet{
				Created: []*domain.TodoChange{},
				Updated: []*domain.TodoChange{},
				Deleted: []*domain.TodoChange{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.changes.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	_ = now
}

func TestChangeSet_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		changes  domain.ChangeSet
		expected bool
	}{
		{
			name:     "empty",
			changes:  domain.ChangeSet{},
			expected: true,
		},
		{
			name: "with created",
			changes: domain.ChangeSet{
				Created: []*domain.TodoChange{{ID: uuid.New()}},
			},
			expected: false,
		},
		{
			name: "with updated",
			changes: domain.ChangeSet{
				Updated: []*domain.TodoChange{{ID: uuid.New()}},
			},
			expected: false,
		},
		{
			name: "with deleted",
			changes: domain.ChangeSet{
				Deleted: []*domain.TodoChange{{ID: uuid.New()}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.changes.IsEmpty()
			if result != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChangeSet_TotalChanges(t *testing.T) {
	changes := domain.ChangeSet{
		Created: []*domain.TodoChange{{ID: uuid.New()}, {ID: uuid.New()}},
		Updated: []*domain.TodoChange{{ID: uuid.New()}},
		Deleted: []*domain.TodoChange{{ID: uuid.New()}, {ID: uuid.New()}, {ID: uuid.New()}},
	}

	if changes.TotalChanges() != 6 {
		t.Errorf("Expected 6 total changes, got %d", changes.TotalChanges())
	}
}
