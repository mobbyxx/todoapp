package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Todo-related errors
var (
	ErrTodoNotFound          = errors.New("todo not found")
	ErrTodoAlreadyExists     = errors.New("todo already exists")
	ErrInvalidTodoTitle      = errors.New("invalid todo title")
	ErrInvalidTodoDescription = errors.New("invalid todo description")
	ErrVersionMismatch       = errors.New("version mismatch - todo was modified by another request")
	ErrInvalidAssignee       = errors.New("invalid assignee")
)

// TodoStatus defines the possible statuses for a todo
type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"
)

// TodoPriority defines the possible priorities for a todo
type TodoPriority string

const (
	TodoPriorityLow    TodoPriority = "low"
	TodoPriorityMedium TodoPriority = "medium"
	TodoPriorityHigh   TodoPriority = "high"
	TodoPriorityUrgent TodoPriority = "urgent"
)

// ValidStatusTransitions defines allowed status transitions
var ValidStatusTransitions = map[TodoStatus][]TodoStatus{
	TodoStatusPending:    {TodoStatusInProgress, TodoStatusCompleted},
	TodoStatusInProgress: {TodoStatusCompleted, TodoStatusPending},
	TodoStatusCompleted:  {TodoStatusInProgress, TodoStatusPending},
}

// Todo represents a todo item in the system
type Todo struct {
	ID          uuid.UUID    `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Status      TodoStatus   `json:"status"`
	Priority    TodoPriority `json:"priority"`
	CreatedBy   uuid.UUID    `json:"created_by"`
	AssignedTo  *uuid.UUID   `json:"assigned_to,omitempty"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	Version     int          `json:"version"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty"`
}

// TodoRepository defines the interface for todo persistence
type TodoRepository interface {
	Create(todo *Todo) error
	GetByID(id uuid.UUID) (*Todo, error)
	GetByUserID(userID uuid.UUID, filters TodoFilters) ([]*Todo, error)
	Update(todo *Todo) error
	Delete(id uuid.UUID) error
	List(filters TodoFilters) ([]*Todo, int, error)
}

// TodoService defines the interface for todo business logic
type TodoService interface {
	Create(userID uuid.UUID, input CreateTodoInput) (*Todo, error)
	Get(userID uuid.UUID, todoID uuid.UUID) (*Todo, error)
	List(userID uuid.UUID, filters TodoFilters) ([]*Todo, int, error)
	Update(userID uuid.UUID, todoID uuid.UUID, input UpdateTodoInput, version int) (*Todo, error)
	Delete(userID uuid.UUID, todoID uuid.UUID) error
	Assign(todoID uuid.UUID, userID uuid.UUID, assignToID uuid.UUID) (*Todo, error)
	Complete(userID uuid.UUID, todoID uuid.UUID, version int) (*Todo, error)
}

// CreateTodoInput represents the input for creating a todo
type CreateTodoInput struct {
	Title       string       `json:"title" validate:"required,min=1,max=200"`
	Description string       `json:"description" validate:"omitempty,max=2000"`
	Priority    TodoPriority `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	AssignedTo  *uuid.UUID   `json:"assigned_to,omitempty"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
}

// UpdateTodoInput represents the input for updating a todo
type UpdateTodoInput struct {
	Title       string       `json:"title" validate:"omitempty,min=1,max=200"`
	Description string       `json:"description" validate:"omitempty,max=2000"`
	Status      TodoStatus   `json:"status" validate:"omitempty,oneof=pending in_progress completed"`
	Priority    TodoPriority `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	AssignedTo  *uuid.UUID   `json:"assigned_to,omitempty"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
}

// TodoFilters represents filters for listing todos
type TodoFilters struct {
	UserID     *uuid.UUID
	Status     *TodoStatus
	Priority   *TodoPriority
	AssignedTo *uuid.UUID
	Search     string
	Page       int
	PageSize   int
}

// TodoResponse represents a todo in API responses
type TodoResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	CreatedBy   string     `json:"created_by"`
	AssignedTo  string     `json:"assigned_to,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse converts a Todo to a TodoResponse
func (t *Todo) ToResponse() TodoResponse {
	resp := TodoResponse{
		ID:          t.ID.String(),
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		CreatedBy:   t.CreatedBy.String(),
		Version:     t.Version,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}

	if t.AssignedTo != nil {
		assignedTo := t.AssignedTo.String()
		resp.AssignedTo = assignedTo
	}

	if t.DueDate != nil {
		resp.DueDate = t.DueDate
	}

	return resp
}

// CanTransitionTo checks if the todo can transition to the given status
func (t *Todo) CanTransitionTo(newStatus TodoStatus) bool {
	allowedTransitions, ok := ValidStatusTransitions[t.Status]
	if !ok {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}
	return false
}

// IsOwnedBy checks if the todo is owned by the given user
func (t *Todo) IsOwnedBy(userID uuid.UUID) bool {
	return t.CreatedBy == userID
}

// IsAssignedTo checks if the todo is assigned to the given user
func (t *Todo) IsAssignedTo(userID uuid.UUID) bool {
	if t.AssignedTo == nil {
		return false
	}
	return *t.AssignedTo == userID
}

// IsAccessibleBy checks if the todo is accessible by the given user (owner or assignee)
func (t *Todo) IsAccessibleBy(userID uuid.UUID) bool {
	return t.IsOwnedBy(userID) || t.IsAssignedTo(userID)
}

// IsDeleted checks if the todo is soft deleted
func (t *Todo) IsDeleted() bool {
	return t.DeletedAt != nil
}

// DefaultPriority returns the default priority for a new todo
func DefaultPriority() TodoPriority {
	return TodoPriorityMedium
}

// ValidatePriority checks if the priority is valid
func ValidatePriority(p TodoPriority) bool {
	switch p {
	case TodoPriorityLow, TodoPriorityMedium, TodoPriorityHigh, TodoPriorityUrgent:
		return true
	default:
		return false
	}
}

// ValidateStatus checks if the status is valid
func ValidateStatus(s TodoStatus) bool {
	switch s {
	case TodoStatusPending, TodoStatusInProgress, TodoStatusCompleted:
		return true
	default:
		return false
	}
}

// TodoListResponse represents a list of todos response
type TodoListResponse struct {
	Todos      []TodoResponse `json:"todos"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}
