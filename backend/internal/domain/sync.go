package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Sync-related errors
var (
	ErrSyncFailed           = errors.New("sync operation failed")
	ErrInvalidChangeSet     = errors.New("invalid change set")
	ErrInvalidSyncTimestamp = errors.New("invalid sync timestamp")
)

// SyncStatus defines the possible statuses for a sync operation
type SyncStatus string

const (
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusInProgress SyncStatus = "in_progress"
	SyncStatusCompleted  SyncStatus = "completed"
	SyncStatusFailed     SyncStatus = "failed"
)

// ValidateSyncStatus checks if the status is valid
func ValidateSyncStatus(s SyncStatus) bool {
	switch s {
	case SyncStatusPending, SyncStatusInProgress, SyncStatusCompleted, SyncStatusFailed:
		return true
	default:
		return false
	}
}

// HLC (Hybrid Logical Clock) provides distributed timestamp ordering
// HLC = max(physical_time, logical_time+1)
type HLC struct {
	PhysicalTime int64 `json:"physical_time"` // Wall clock time (milliseconds)
	LogicalTime  int64 `json:"logical_time"`  // Lamport clock counter
}

// NewHLC creates a new HLC with current physical time
func NewHLC() HLC {
	return HLC{
		PhysicalTime: time.Now().UnixMilli(),
		LogicalTime:  0,
	}
}

// HLCFromTime creates an HLC from a given time
func HLCFromTime(t time.Time) HLC {
	return HLC{
		PhysicalTime: t.UnixMilli(),
		LogicalTime:  0,
	}
}

// String returns string representation of HLC
func (h HLC) String() string {
	return fmt.Sprintf("%d.%d", h.PhysicalTime, h.LogicalTime)
}

// IsZero checks if HLC is zero value
func (h HLC) IsZero() bool {
	return h.PhysicalTime == 0 && h.LogicalTime == 0
}

// Less checks if this HLC is less than another (happened-before)
func (h HLC) Less(other HLC) bool {
	if h.PhysicalTime != other.PhysicalTime {
		return h.PhysicalTime < other.PhysicalTime
	}
	return h.LogicalTime < other.LogicalTime
}

// Equal checks if two HLCs are equal
func (h HLC) Equal(other HLC) bool {
	return h.PhysicalTime == other.PhysicalTime && h.LogicalTime == other.LogicalTime
}

// Compare compares two HLCs: -1 if h < other, 0 if equal, 1 if h > other
func (h HLC) Compare(other HLC) int {
	if h.Less(other) {
		return -1
	}
	if h.Equal(other) {
		return 0
	}
	return 1
}

// ToTime converts HLC to time.Time
func (h HLC) ToTime() time.Time {
	return time.UnixMilli(h.PhysicalTime)
}

// HLCClock manages HLC timestamps with thread safety
type HLCClock struct {
	mu       sync.RWMutex
	lastTime int64
	logical  int64
}

// NewHLCClock creates a new HLC clock
func NewHLCClock() *HLCClock {
	return &HLCClock{
		lastTime: time.Now().UnixMilli(),
		logical:  0,
	}
}

// Now returns the current HLC timestamp
func (c *HLCClock) Now() HLC {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixMilli()

	if now > c.lastTime {
		c.lastTime = now
		c.logical = 0
	} else {
		c.logical++
	}

	return HLC{
		PhysicalTime: c.lastTime,
		LogicalTime:  c.logical,
	}
}

// Update updates the clock with a received HLC
func (c *HLCClock) Update(received HLC) HLC {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixMilli()

	// HLC = max(physical_time, logical_time+1)
	if now > c.lastTime && now > received.PhysicalTime {
		c.lastTime = now
		c.logical = 0
	} else if received.PhysicalTime > c.lastTime {
		c.lastTime = received.PhysicalTime
		c.logical = received.LogicalTime + 1
	} else if received.PhysicalTime == c.lastTime {
		c.logical = max(c.logical, received.LogicalTime) + 1
	} else {
		c.logical++
	}

	return HLC{
		PhysicalTime: c.lastTime,
		LogicalTime:  c.logical,
	}
}

// SyncRecord tracks sync state per user
type SyncRecord struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	LastSyncedAt  HLC        `json:"last_synced_at"`
	Status        SyncStatus `json:"status"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	ClientVersion string     `json:"client_version,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ChangeSet represents a batch of changes for sync
type ChangeSet struct {
	Created []*TodoChange `json:"created"`
	Updated []*TodoChange `json:"updated"`
	Deleted []*TodoChange `json:"deleted"`
}

// TodoChange represents a single todo change for sync
type TodoChange struct {
	ID          uuid.UUID       `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Status      TodoStatus      `json:"status"`
	Priority    TodoPriority    `json:"priority"`
	CreatedBy   uuid.UUID       `json:"created_by"`
	AssignedTo  *uuid.UUID      `json:"assigned_to,omitempty"`
	DueDate     *time.Time      `json:"due_date,omitempty"`
	Version     int             `json:"version"`
	Timestamp   HLC             `json:"timestamp"` // When change was made
	IsDeleted   bool            `json:"is_deleted"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

// ToTodo converts TodoChange to Todo
func (c *TodoChange) ToTodo() *Todo {
	return &Todo{
		ID:          c.ID,
		Title:       c.Title,
		Description: c.Description,
		Status:      c.Status,
		Priority:    c.Priority,
		CreatedBy:   c.CreatedBy,
		AssignedTo:  c.AssignedTo,
		DueDate:     c.DueDate,
		Version:     c.Version,
		UpdatedAt:   c.Timestamp.ToTime(),
	}
}

// FromTodo creates TodoChange from Todo
func TodoChangeFromTodo(todo *Todo) *TodoChange {
	return &TodoChange{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      todo.Status,
		Priority:    todo.Priority,
		CreatedBy:   todo.CreatedBy,
		AssignedTo:  todo.AssignedTo,
		DueDate:     todo.DueDate,
		Version:     todo.Version,
		Timestamp:   HLCFromTime(todo.UpdatedAt),
		IsDeleted:   todo.DeletedAt != nil,
	}
}

// ConnectionChange represents a connection change for sync (read-only from client)
type ConnectionChange struct {
	ID         uuid.UUID        `json:"id"`
	UserAID    uuid.UUID        `json:"user_a_id"`
	UserBID    uuid.UUID        `json:"user_b_id"`
	Status     ConnectionStatus `json:"status"`
	RequestedBy uuid.UUID       `json:"requested_by"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	Timestamp  HLC              `json:"timestamp"`
}

// SyncPullRequest represents a pull request from client
type SyncPullRequest struct {
	LastPulledAt HLC `json:"last_pulled_at"`
}

// SyncPushRequest represents a push request from client
type SyncPushRequest struct {
	Changes   ChangeSet `json:"changes"`
	Timestamp HLC       `json:"timestamp"`
}

// SyncRequest represents a bidirectional sync request
type SyncRequest struct {
	LastPulledAt HLC       `json:"last_pulled_at"`
	Changes      ChangeSet `json:"changes,omitempty"`
}

// SyncResponse represents a sync response
type SyncResponse struct {
	Timestamp  HLC              `json:"timestamp"`
	Changes    ChangeSet        `json:"changes"`
	Conflicts  []*SyncConflict  `json:"conflicts"`
	Status     SyncStatus       `json:"status"`
	ServerTime int64            `json:"server_time"` // Unix milliseconds for client clock skew detection
}

// SyncConflict represents a detected conflict
type SyncConflict struct {
	EntityType   string          `json:"entity_type"`
	EntityID     uuid.UUID       `json:"entity_id"`
	LocalVersion *TodoChange     `json:"local_version"`  // Client's version
	ServerVersion *TodoChange    `json:"server_version"` // Server's version
	ConflictType string          `json:"conflict_type"`  // "both_modified", "delete_modified", etc.
}

// SyncRepository defines the interface for sync persistence
type SyncRepository interface {
	// GetLastSync returns the user's last sync record
	GetLastSync(userID uuid.UUID) (*SyncRecord, error)
	
	// UpdateLastSync updates the user's last sync timestamp
	UpdateLastSync(userID uuid.UUID, timestamp HLC) error
	
	// GetChangesSince returns all changes since the given timestamp for a user
	GetChangesSince(userID uuid.UUID, since HLC) (*ChangeSet, error)
	
	// ApplyChanges applies client changes to the server
	ApplyChanges(userID uuid.UUID, changes ChangeSet, timestamp HLC) ([]*SyncConflict, error)
	
	// RecordConflict records a sync conflict for later resolution
	RecordConflict(conflict *SyncConflictRecord) error
}

// SyncConflictRecord represents a recorded conflict in the database
type SyncConflictRecord struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"user_id"`
	EntityType       string          `json:"entity_type"`
	EntityID         uuid.UUID       `json:"entity_id"`
	LocalVersion     int             `json:"local_version"`
	RemoteVersion    int             `json:"remote_version"`
	LocalData        json.RawMessage `json:"local_data"`
	RemoteData       json.RawMessage `json:"remote_data"`
	ResolutionStrategy string        `json:"resolution_strategy"`
	ResolvedData     json.RawMessage `json:"resolved_data,omitempty"`
	ResolvedAt       *time.Time      `json:"resolved_at,omitempty"`
	ResolvedBy       *uuid.UUID      `json:"resolved_by,omitempty"`
	Status           string          `json:"status"`
	ClientTimestamp  HLC             `json:"client_timestamp"`
	ServerTimestamp  HLC             `json:"server_timestamp"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// ConflictStatus defines the possible statuses for a conflict
type ConflictStatus string

const (
	ConflictStatusPending    ConflictStatus = "pending"
	ConflictStatusResolved   ConflictStatus = "resolved"
	ConflictStatusAutoMerged ConflictStatus = "auto_merged"
)

// ResolutionStrategy defines the strategy for resolving conflicts
type ResolutionStrategy string

const (
	ResolutionStrategyLastWriteWins ResolutionStrategy = "last-write-wins"
	ResolutionStrategyMaxWins       ResolutionStrategy = "max-wins"
	ResolutionStrategyMerge         ResolutionStrategy = "merge"
	ResolutionStrategyPrompt        ResolutionStrategy = "prompt"
)

// FieldConflictResolution defines resolution strategy per field
var FieldConflictResolution = map[string]ResolutionStrategy{
	"title":       ResolutionStrategyPrompt,
	"description": ResolutionStrategyPrompt,
	"status":      ResolutionStrategyLastWriteWins,
	"due_date":    ResolutionStrategyMerge,
	"priority":    ResolutionStrategyMaxWins,
	"assigned_to": ResolutionStrategyPrompt,
}

// ConflictResolution represents a manual conflict resolution
type ConflictResolution struct {
	ConflictID   uuid.UUID       `json:"conflict_id" validate:"required"`
	ResolvedData json.RawMessage `json:"resolved_data" validate:"required"`
	Strategy     ResolutionStrategy `json:"strategy" validate:"omitempty,oneof=local remote merge custom"`
}

// SyncService defines the interface for sync business logic
type SyncService interface {
	// PullChanges gets server changes since the given timestamp
	PullChanges(userID uuid.UUID, lastPulledAt HLC) (*ChangeSet, error)
	
	// PushChanges applies client changes to the server
	PushChanges(userID uuid.UUID, changes ChangeSet) ([]*SyncConflict, error)
	
	// GetLastSync gets the user's last sync timestamp
	GetLastSync(userID uuid.UUID) (HLC, error)
	
	// UpdateLastSync updates the user's last sync timestamp
	UpdateLastSync(userID uuid.UUID, timestamp HLC) error
	
	// Sync performs bidirectional sync (pull + push)
	Sync(userID uuid.UUID, req SyncRequest) (*SyncResponse, error)
}

// ConflictRepository defines the interface for conflict persistence
type ConflictRepository interface {
	// GetByID retrieves a conflict by ID
	GetByID(id uuid.UUID) (*SyncConflictRecord, error)
	
	// GetUnresolvedByUser retrieves all unresolved conflicts for a user
	GetUnresolvedByUser(userID uuid.UUID) ([]*SyncConflictRecord, error)
	
	// UpdateResolution updates a conflict with its resolution
	UpdateResolution(conflictID uuid.UUID, resolution *ConflictResolution, resolvedBy uuid.UUID) error
	
	// RecordConflict records a new conflict
	RecordConflict(conflict *SyncConflictRecord) error
}

// ConflictService defines the interface for conflict resolution business logic
type ConflictService interface {
	// DetectConflicts identifies conflicts between local and remote changes
	DetectConflicts(userID uuid.UUID, changes ChangeSet) ([]*SyncConflict, error)
	
	// ResolveConflict resolves a single conflict using the specified strategy
	ResolveConflict(entityType string, local, remote *TodoChange, strategy ResolutionStrategy) (*TodoChange, error)
	
	// GetUnresolvedConflicts retrieves all pending conflicts for a user
	GetUnresolvedConflicts(userID uuid.UUID) ([]*SyncConflictRecord, error)
	
	// ResolveConflictManually allows user to manually resolve a conflict
	ResolveConflictManually(conflictID uuid.UUID, resolution ConflictResolution, userID uuid.UUID) error
}

// SyncFilters represents filters for sync operations
type SyncFilters struct {
	UserID    *uuid.UUID
	Since     *HLC
	EntityType string
}

// IsValid checks if the change set is valid
func (cs *ChangeSet) IsValid() error {
	for i, created := range cs.Created {
		if created.ID == uuid.Nil {
			return fmt.Errorf("created[%d]: missing ID", i)
		}
		if created.Title == "" {
			return fmt.Errorf("created[%d]: missing title", i)
		}
		if created.CreatedBy == uuid.Nil {
			return fmt.Errorf("created[%d]: missing created_by", i)
		}
	}

	for i, updated := range cs.Updated {
		if updated.ID == uuid.Nil {
			return fmt.Errorf("updated[%d]: missing ID", i)
		}
	}

	for i, deleted := range cs.Deleted {
		if deleted.ID == uuid.Nil {
			return fmt.Errorf("deleted[%d]: missing ID", i)
		}
	}

	return nil
}

// IsEmpty checks if the change set is empty
func (cs *ChangeSet) IsEmpty() bool {
	return len(cs.Created) == 0 && len(cs.Updated) == 0 && len(cs.Deleted) == 0
}

// TotalChanges returns the total number of changes
func (cs *ChangeSet) TotalChanges() int {
	return len(cs.Created) + len(cs.Updated) + len(cs.Deleted)
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
