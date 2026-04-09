package service

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/user/todo-api/internal/domain"
)

type mockTodoRepoForAntiCheat struct {
	todos map[uuid.UUID]*domain.Todo
}

func newMockTodoRepoForAntiCheat() *mockTodoRepoForAntiCheat {
	return &mockTodoRepoForAntiCheat{
		todos: make(map[uuid.UUID]*domain.Todo),
	}
}

func (m *mockTodoRepoForAntiCheat) Create(todo *domain.Todo) error {
	m.todos[todo.ID] = todo
	return nil
}

func (m *mockTodoRepoForAntiCheat) GetByID(id uuid.UUID) (*domain.Todo, error) {
	todo, ok := m.todos[id]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	return todo, nil
}

func (m *mockTodoRepoForAntiCheat) GetByUserID(userID uuid.UUID, filters domain.TodoFilters) ([]*domain.Todo, error) {
	var result []*domain.Todo
	for _, todo := range m.todos {
		if todo.CreatedBy == userID || (todo.AssignedTo != nil && *todo.AssignedTo == userID) {
			result = append(result, todo)
		}
	}
	return result, nil
}

func (m *mockTodoRepoForAntiCheat) Update(todo *domain.Todo) error {
	if _, ok := m.todos[todo.ID]; !ok {
		return domain.ErrTodoNotFound
	}
	m.todos[todo.ID] = todo
	return nil
}

func (m *mockTodoRepoForAntiCheat) Delete(id uuid.UUID) error {
	delete(m.todos, id)
	return nil
}

func (m *mockTodoRepoForAntiCheat) List(filters domain.TodoFilters) ([]*domain.Todo, int, error) {
	var result []*domain.Todo
	for _, todo := range m.todos {
		result = append(result, todo)
	}
	return result, len(result), nil
}

func setupAntiCheatTest(t *testing.T) (*antiCheatService, *miniredis.Miniredis, *mockTodoRepoForAntiCheat) {
	mr := miniredis.RunT(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	todoRepo := newMockTodoRepoForAntiCheat()
	config := domain.DefaultAntiCheatConfig()

	service := NewAntiCheatService(rdb, todoRepo, config).(*antiCheatService)

	return service, mr, todoRepo
}

func createTestTodo(repo *mockTodoRepoForAntiCheat, userID uuid.UUID, status domain.TodoStatus, assignedTo *uuid.UUID) *domain.Todo {
	todo := &domain.Todo{
		ID:          uuid.New(),
		Title:       "Test Todo",
		Description: "Test Description",
		Status:      status,
		Priority:    domain.TodoPriorityMedium,
		CreatedBy:   userID,
		AssignedTo:  assignedTo,
		Version:     1,
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now(),
	}
	repo.Create(todo)
	return todo
}

func TestNewAntiCheatService(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	todoRepo := newMockTodoRepository()
	config := domain.DefaultAntiCheatConfig()

	service := NewAntiCheatService(rdb, todoRepo, config)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}
}

func TestAntiCheatService_ValidateTodoComplete_Success(t *testing.T) {
	service, mr, todoRepo := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	todo := createTestTodo(todoRepo, userID, domain.TodoStatusPending, nil)

	err := service.ValidateTodoComplete(userID, todo.ID, time.Now())
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}
}

func TestAntiCheatService_ValidateTodoComplete_AlreadyCompleted(t *testing.T) {
	service, mr, todoRepo := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	todo := createTestTodo(todoRepo, userID, domain.TodoStatusCompleted, nil)

	err := service.ValidateTodoComplete(userID, todo.ID, time.Now())
	if err != domain.ErrTodoAlreadyCompleted {
		t.Errorf("Expected ErrTodoAlreadyCompleted, got: %v", err)
	}
}

func TestAntiCheatService_ValidateTodoComplete_RateLimitExceeded(t *testing.T) {
	service, mr, todoRepo := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()

	for i := 0; i < 11; i++ {
		todo := createTestTodo(todoRepo, uuid.New(), domain.TodoStatusPending, nil)

		err := service.ValidateTodoComplete(userID, todo.ID, time.Now())

		if i < 10 && err != nil {
			t.Errorf("Expected no error on iteration %d, got: %v", i, err)
		}
		if i >= 10 && err != domain.ErrRateLimitExceeded {
			t.Errorf("Expected ErrRateLimitExceeded on iteration %d, got: %v", i, err)
		}
	}
}

func TestAntiCheatService_ValidateTodoComplete_InvalidTimestamp(t *testing.T) {
	service, mr, todoRepo := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	todo := createTestTodo(todoRepo, userID, domain.TodoStatusPending, nil)

	futureTime := time.Now().Add(10 * time.Minute)
	err := service.ValidateTodoComplete(userID, todo.ID, futureTime)
	if err != domain.ErrInvalidTimestamp {
		t.Errorf("Expected ErrInvalidTimestamp, got: %v", err)
	}

	pastTime := time.Now().Add(-10 * time.Minute)
	err = service.ValidateTodoComplete(uuid.New(), createTestTodo(todoRepo, uuid.New(), domain.TodoStatusPending, nil).ID, pastTime)
	if err != domain.ErrInvalidTimestamp {
		t.Errorf("Expected ErrInvalidTimestamp for past time, got: %v", err)
	}
}

func TestAntiCheatService_ValidateTodoComplete_DuplicateAction(t *testing.T) {
	service, mr, todoRepo := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	todo := createTestTodo(todoRepo, userID, domain.TodoStatusPending, nil)

	err := service.ValidateTodoComplete(userID, todo.ID, time.Now())
	if err != nil {
		t.Fatalf("First validation should pass, got: %v", err)
	}

	err = service.ValidateTodoComplete(userID, todo.ID, time.Now())
	if err != domain.ErrDuplicateAction {
		t.Errorf("Expected ErrDuplicateAction, got: %v", err)
	}
}

func TestAntiCheatService_ValidateTodoComplete_ActionTooFast(t *testing.T) {
	service, mr, todoRepo := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	todo1 := createTestTodo(todoRepo, userID, domain.TodoStatusPending, nil)

	err := service.ValidateTodoComplete(userID, todo1.ID, time.Now())
	if err != nil {
		t.Fatalf("First validation should pass, got: %v", err)
	}

	todo2 := createTestTodo(todoRepo, userID, domain.TodoStatusPending, nil)
	err = service.ValidateTodoComplete(userID, todo2.ID, time.Now())
	if err != domain.ErrActionTooFast {
		t.Errorf("Expected ErrActionTooFast, got: %v", err)
	}
}

func TestAntiCheatService_ValidateTodoComplete_BackdatedTimestamp(t *testing.T) {
	service, mr, todoRepo := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	todo := createTestTodo(todoRepo, userID, domain.TodoStatusPending, nil)

	backdatedTime := todo.CreatedAt.Add(-time.Hour)
	err := service.ValidateTodoComplete(userID, todo.ID, backdatedTime)
	if err != domain.ErrTimestampBackdated {
		t.Errorf("Expected ErrTimestampBackdated, got: %v", err)
	}
}

func TestAntiCheatService_ValidateTodoComplete_SelfAssignment(t *testing.T) {
	service, mr, todoRepo := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	todo := createTestTodo(todoRepo, userID, domain.TodoStatusPending, &userID)

	err := service.ValidateTodoComplete(userID, todo.ID, time.Now())
	if err != domain.ErrSelfAssignmentDetected {
		t.Errorf("Expected ErrSelfAssignmentDetected, got: %v", err)
	}
}

func TestAntiCheatService_ValidateTodoComplete_AssignedToOther(t *testing.T) {
	service, mr, todoRepo := setupAntiCheatTest(t)
	defer mr.Close()

	ownerID := uuid.New()
	assigneeID := uuid.New()
	todo := createTestTodo(todoRepo, ownerID, domain.TodoStatusPending, &assigneeID)

	err := service.ValidateTodoComplete(assigneeID, todo.ID, time.Now())
	if err != nil {
		t.Errorf("Expected validation to pass for valid assignee, got: %v", err)
	}
}

func TestAntiCheatService_ValidateTodoComplete_TodoNotFound(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	fakeTodoID := uuid.New()

	err := service.ValidateTodoComplete(userID, fakeTodoID, time.Now())
	if err != domain.ErrTodoNotFound {
		t.Errorf("Expected ErrTodoNotFound, got: %v", err)
	}
}

func TestAntiCheatService_CheckRateLimit_Success(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()

	for i := 0; i < 10; i++ {
		err := service.CheckRateLimit(userID, domain.ActionTypeTodoComplete)
		if err != nil {
			t.Errorf("Expected no error on iteration %d, got: %v", i, err)
		}
	}
}

func TestAntiCheatService_CheckRateLimit_Exceeded(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()

	for i := 0; i < 10; i++ {
		service.CheckRateLimit(userID, domain.ActionTypeTodoComplete)
	}

	err := service.CheckRateLimit(userID, domain.ActionTypeTodoComplete)
	if err != domain.ErrRateLimitExceeded {
		t.Errorf("Expected ErrRateLimitExceeded, got: %v", err)
	}
}

func TestAntiCheatService_CheckRateLimit_ResetsAfterWindow(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	service.config.RateLimitWindow = time.Millisecond * 100

	userID := uuid.New()

	for i := 0; i < 10; i++ {
		service.CheckRateLimit(userID, domain.ActionTypeTodoComplete)
	}

	err := service.CheckRateLimit(userID, domain.ActionTypeTodoComplete)
	if err != domain.ErrRateLimitExceeded {
		t.Errorf("Expected ErrRateLimitExceeded, got: %v", err)
	}

	time.Sleep(time.Millisecond * 150)

	err = service.CheckRateLimit(userID, domain.ActionTypeTodoComplete)
	if err != nil {
		t.Errorf("Expected rate limit to reset, got: %v", err)
	}
}

func TestAntiCheatService_CheckTimestamp_WithinTolerance(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	testCases := []struct {
		name       string
		clientTime time.Time
		wantErr    bool
	}{
		{
			name:       "exact server time",
			clientTime: time.Now(),
			wantErr:    false,
		},
		{
			name:       "2 minutes behind",
			clientTime: time.Now().Add(-2 * time.Minute),
			wantErr:    false,
		},
		{
			name:       "2 minutes ahead",
			clientTime: time.Now().Add(2 * time.Minute),
			wantErr:    false,
		},
		{
			name:       "6 minutes behind",
			clientTime: time.Now().Add(-6 * time.Minute),
			wantErr:    true,
		},
		{
			name:       "6 minutes ahead",
			clientTime: time.Now().Add(6 * time.Minute),
			wantErr:    true,
		},
		{
			name:       "zero time",
			clientTime: time.Time{},
			wantErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.CheckTimestamp(tc.clientTime)
			if tc.wantErr && err != domain.ErrInvalidTimestamp {
				t.Errorf("Expected ErrInvalidTimestamp, got: %v", err)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestAntiCheatService_CheckIdempotency_Success(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	todoID := uuid.New()

	err := service.CheckIdempotency(userID, todoID)
	if err != nil {
		t.Errorf("Expected first check to pass, got: %v", err)
	}

	err = service.CheckIdempotency(userID, todoID)
	if err != domain.ErrDuplicateAction {
		t.Errorf("Expected ErrDuplicateAction, got: %v", err)
	}
}

func TestAntiCheatService_CheckIdempotency_DifferentUsers(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	user1 := uuid.New()
	user2 := uuid.New()
	todoID := uuid.New()

	err := service.CheckIdempotency(user1, todoID)
	if err != nil {
		t.Errorf("Expected user1 check to pass, got: %v", err)
	}

	err = service.CheckIdempotency(user2, todoID)
	if err != nil {
		t.Errorf("Expected user2 check to pass, got: %v", err)
	}
}

func TestAntiCheatService_CheckIdempotency_DifferentTodos(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	todo1 := uuid.New()
	todo2 := uuid.New()

	err := service.CheckIdempotency(userID, todo1)
	if err != nil {
		t.Errorf("Expected first todo check to pass, got: %v", err)
	}

	err = service.CheckIdempotency(userID, todo2)
	if err != nil {
		t.Errorf("Expected second todo check to pass, got: %v", err)
	}
}

func TestAntiCheatService_RecordAction(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()

	err := service.RecordAction(userID, domain.ActionTypeTodoComplete)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestAntiCheatService_CheckMinTimeGap_NoPreviousAction(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()

	err := service.CheckMinTimeGap(userID, domain.ActionTypeTodoComplete)
	if err != nil {
		t.Errorf("Expected no error with no previous action, got: %v", err)
	}
}

func TestAntiCheatService_CheckMinTimeGap_WithPreviousAction(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	service.config.MinActionGap = time.Millisecond * 100

	userID := uuid.New()

	err := service.RecordAction(userID, domain.ActionTypeTodoComplete)
	if err != nil {
		t.Fatalf("Failed to record action: %v", err)
	}

	err = service.CheckMinTimeGap(userID, domain.ActionTypeTodoComplete)
	if err != domain.ErrActionTooFast {
		t.Errorf("Expected ErrActionTooFast, got: %v", err)
	}

	time.Sleep(time.Millisecond * 150)

	err = service.CheckMinTimeGap(userID, domain.ActionTypeTodoComplete)
	if err != nil {
		t.Errorf("Expected gap to pass after wait, got: %v", err)
	}
}

func TestAntiCheatService_CheckStatusCycle_NoCycle(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	userID := uuid.New()
	todoID := uuid.New()

	err := service.CheckStatusCycle(userID, todoID, domain.TodoStatusCompleted)
	if err != nil {
		t.Errorf("Expected no error on first status change, got: %v", err)
	}
}

func TestAntiCheatService_CheckStatusCycle_Detected(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	service.config.StatusCycleThreshold = 3

	userID := uuid.New()
	todoID := uuid.New()

	service.CheckStatusCycle(userID, todoID, domain.TodoStatusCompleted)
	service.CheckStatusCycle(userID, todoID, domain.TodoStatusPending)
	service.CheckStatusCycle(userID, todoID, domain.TodoStatusCompleted)

	err := service.CheckStatusCycle(userID, todoID, domain.TodoStatusPending)
	if err != domain.ErrStatusCycleDetected {
		t.Errorf("Expected ErrStatusCycleDetected, got: %v", err)
	}
}

func TestAntiCheatService_CheckStatusCycle_DifferentTodos(t *testing.T) {
	service, mr, _ := setupAntiCheatTest(t)
	defer mr.Close()

	service.config.StatusCycleThreshold = 2

	userID := uuid.New()
	todo1 := uuid.New()
	todo2 := uuid.New()

	service.CheckStatusCycle(userID, todo1, domain.TodoStatusCompleted)
	service.CheckStatusCycle(userID, todo1, domain.TodoStatusPending)

	err := service.CheckStatusCycle(userID, todo2, domain.TodoStatusCompleted)
	if err != nil {
		t.Errorf("Expected different todo to pass, got: %v", err)
	}
}

func TestAntiCheatService_DefaultConfig(t *testing.T) {
	config := domain.DefaultAntiCheatConfig()

	if config.RateLimitMaxActions != 10 {
		t.Errorf("Expected RateLimitMaxActions=10, got: %d", config.RateLimitMaxActions)
	}

	if config.RateLimitWindow != time.Minute {
		t.Errorf("Expected RateLimitWindow=1m, got: %v", config.RateLimitWindow)
	}

	if config.TimestampTolerance != 5*time.Minute {
		t.Errorf("Expected TimestampTolerance=5m, got: %v", config.TimestampTolerance)
	}

	if config.IdempotencyTTL != 24*time.Hour {
		t.Errorf("Expected IdempotencyTTL=24h, got: %v", config.IdempotencyTTL)
	}

	if config.MinActionGap != 5*time.Second {
		t.Errorf("Expected MinActionGap=5s, got: %v", config.MinActionGap)
	}

	if config.StatusCycleThreshold != 5 {
		t.Errorf("Expected StatusCycleThreshold=5, got: %d", config.StatusCycleThreshold)
	}

	if config.StatusCycleWindow != time.Minute {
		t.Errorf("Expected StatusCycleWindow=1m, got: %v", config.StatusCycleWindow)
	}
}
