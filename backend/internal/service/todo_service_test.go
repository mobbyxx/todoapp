package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

type mockTodoRepository struct {
	todos map[uuid.UUID]*domain.Todo
}

func newMockTodoRepository() *mockTodoRepository {
	return &mockTodoRepository{
		todos: make(map[uuid.UUID]*domain.Todo),
	}
}

func (m *mockTodoRepository) Create(todo *domain.Todo) error {
	todo.ID = uuid.New()
	todo.Version = 1
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()
	m.todos[todo.ID] = todo
	return nil
}

func (m *mockTodoRepository) GetByID(id uuid.UUID) (*domain.Todo, error) {
	if todo, ok := m.todos[id]; ok && !todo.IsDeleted() {
		return todo, nil
	}
	return nil, domain.ErrTodoNotFound
}

func (m *mockTodoRepository) GetByUserID(userID uuid.UUID, filters domain.TodoFilters) ([]*domain.Todo, error) {
	var result []*domain.Todo
	for _, todo := range m.todos {
		if !todo.IsDeleted() && todo.IsAccessibleBy(userID) {
			result = append(result, todo)
		}
	}
	return result, nil
}

func (m *mockTodoRepository) Update(todo *domain.Todo) error {
	existing, ok := m.todos[todo.ID]
	if !ok || existing.IsDeleted() {
		return domain.ErrTodoNotFound
	}
	if existing.Version != todo.Version {
		return domain.ErrVersionMismatch
	}
	todo.Version = existing.Version + 1
	todo.UpdatedAt = time.Now()
	m.todos[todo.ID] = todo
	return nil
}

func (m *mockTodoRepository) Delete(id uuid.UUID) error {
	if todo, ok := m.todos[id]; ok && !todo.IsDeleted() {
		now := time.Now()
		todo.DeletedAt = &now
		todo.UpdatedAt = now
		return nil
	}
	return domain.ErrTodoNotFound
}

func (m *mockTodoRepository) List(filters domain.TodoFilters) ([]*domain.Todo, int, error) {
	var result []*domain.Todo
	for _, todo := range m.todos {
		if todo.IsDeleted() {
			continue
		}
		if filters.UserID != nil && !todo.IsAccessibleBy(*filters.UserID) {
			continue
		}
		if filters.Status != nil && todo.Status != *filters.Status {
			continue
		}
		if filters.Priority != nil && todo.Priority != *filters.Priority {
			continue
		}
		result = append(result, todo)
	}
	return result, len(result), nil
}

type mockConnectionRepositoryForTodo struct {
	connections map[string]*domain.Connection
}

func newMockConnectionRepositoryForTodo() *mockConnectionRepositoryForTodo {
	return &mockConnectionRepositoryForTodo{
		connections: make(map[string]*domain.Connection),
	}
}

func (m *mockConnectionRepositoryForTodo) Create(connection *domain.Connection) error {
	connection.ID = uuid.New()
	key := connection.UserAID.String() + ":" + connection.UserBID.String()
	m.connections[key] = connection
	return nil
}

func (m *mockConnectionRepositoryForTodo) GetByID(id uuid.UUID) (*domain.Connection, error) {
	for _, conn := range m.connections {
		if conn.ID == id {
			return conn, nil
		}
	}
	return nil, domain.ErrConnectionNotFound
}

func (m *mockConnectionRepositoryForTodo) GetByToken(token string) (*domain.Connection, error) {
	for _, conn := range m.connections {
		if conn.InvitationToken == token {
			return conn, nil
		}
	}
	return nil, domain.ErrConnectionNotFound
}

func (m *mockConnectionRepositoryForTodo) GetByUserID(userID uuid.UUID) ([]*domain.Connection, error) {
	var result []*domain.Connection
	for _, conn := range m.connections {
		if conn.UserAID == userID || conn.UserBID == userID {
			result = append(result, conn)
		}
	}
	return result, nil
}

func (m *mockConnectionRepositoryForTodo) GetByUserPair(userAID, userBID uuid.UUID) (*domain.Connection, error) {
	aid, bid := domain.NormalizeUserPair(userAID, userBID)
	key := aid.String() + ":" + bid.String()
	if conn, ok := m.connections[key]; ok {
		return conn, nil
	}
	return nil, domain.ErrConnectionNotFound
}

func (m *mockConnectionRepositoryForTodo) Update(connection *domain.Connection) error {
	key := connection.UserAID.String() + ":" + connection.UserBID.String()
	m.connections[key] = connection
	return nil
}

func (m *mockConnectionRepositoryForTodo) Delete(id uuid.UUID) error {
	for key, conn := range m.connections {
		if conn.ID == id {
			delete(m.connections, key)
			return nil
		}
	}
	return domain.ErrConnectionNotFound
}

type mockUserRepositoryForTodo struct {
	users map[uuid.UUID]*domain.User
}

func newMockUserRepositoryForTodo() *mockUserRepositoryForTodo {
	return &mockUserRepositoryForTodo{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (m *mockUserRepositoryForTodo) Create(user *domain.User) error {
	user.ID = uuid.New()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepositoryForTodo) GetByID(id uuid.UUID) (*domain.User, error) {
	if user, ok := m.users[id]; ok && user.IsActive {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepositoryForTodo) GetByEmail(email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email && user.IsActive {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepositoryForTodo) Update(user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepositoryForTodo) Delete(id uuid.UUID) error {
	if user, ok := m.users[id]; ok {
		user.IsActive = false
		return nil
	}
	return domain.ErrUserNotFound
}

func (m *mockUserRepositoryForTodo) UpdateLastSeen(id uuid.UUID) error {
	return nil
}

func setupTestTodoService(t *testing.T) (domain.TodoService, *mockTodoRepository, *mockConnectionRepositoryForTodo, *mockUserRepositoryForTodo) {
	todoRepo := newMockTodoRepository()
	connRepo := newMockConnectionRepositoryForTodo()
	userRepo := newMockUserRepositoryForTodo()
	service := NewTodoService(todoRepo, connRepo, userRepo, nil, nil)
	return service, todoRepo, connRepo, userRepo
}

func TestTodoService_Create_Success(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title:       "Test Todo",
		Description: "Test Description",
		Priority:    domain.TodoPriorityHigh,
	}

	todo, err := service.Create(user.ID, input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if todo == nil {
		t.Fatal("Expected todo, got nil")
	}

	if todo.Title != input.Title {
		t.Errorf("Expected title %s, got %s", input.Title, todo.Title)
	}

	if todo.Description != input.Description {
		t.Errorf("Expected description %s, got %s", input.Description, todo.Description)
	}

	if todo.Priority != input.Priority {
		t.Errorf("Expected priority %s, got %s", input.Priority, todo.Priority)
	}

	if todo.Status != domain.TodoStatusPending {
		t.Errorf("Expected status %s, got %s", domain.TodoStatusPending, todo.Status)
	}

	if todo.Version != 1 {
		t.Errorf("Expected version 1, got %d", todo.Version)
	}

	if todo.CreatedBy != user.ID {
		t.Errorf("Expected created_by %s, got %s", user.ID, todo.CreatedBy)
	}
}

func TestTodoService_Create_DefaultPriority(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}

	todo, err := service.Create(user.ID, input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if todo.Priority != domain.TodoPriorityMedium {
		t.Errorf("Expected default priority %s, got %s", domain.TodoPriorityMedium, todo.Priority)
	}
}

func TestTodoService_Create_InvalidTitle(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title: "",
	}

	_, err := service.Create(user.ID, input)

	if err == nil {
		t.Fatal("Expected error for empty title")
	}

	if err != domain.ErrInvalidTodoTitle {
		t.Errorf("Expected ErrInvalidTodoTitle, got %v", err)
	}
}

func TestTodoService_Create_TitleTooLong(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	longTitle := ""
	for i := 0; i < 201; i++ {
		longTitle += "a"
	}

	input := domain.CreateTodoInput{
		Title: longTitle,
	}

	_, err := service.Create(user.ID, input)

	if err == nil {
		t.Fatal("Expected error for long title")
	}

	if err != domain.ErrInvalidTodoTitle {
		t.Errorf("Expected ErrInvalidTodoTitle, got %v", err)
	}
}

func TestTodoService_Get_Success(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}
	createdTodo, _ := service.Create(user.ID, input)

	retrievedTodo, err := service.Get(user.ID, createdTodo.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrievedTodo.ID != createdTodo.ID {
		t.Errorf("Expected ID %s, got %s", createdTodo.ID, retrievedTodo.ID)
	}
}

func TestTodoService_Get_NotFound(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	_, err := service.Get(user.ID, uuid.New())

	if err == nil {
		t.Fatal("Expected error for non-existent todo")
	}

	if err != domain.ErrTodoNotFound {
		t.Errorf("Expected ErrTodoNotFound, got %v", err)
	}
}

func TestTodoService_Get_Unauthorized(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	owner := &domain.User{
		Email:       "owner@example.com",
		DisplayName: "Owner",
		IsActive:    true,
	}
	userRepo.Create(owner)

	otherUser := &domain.User{
		Email:       "other@example.com",
		DisplayName: "Other",
		IsActive:    true,
	}
	userRepo.Create(otherUser)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}
	createdTodo, _ := service.Create(owner.ID, input)

	_, err := service.Get(otherUser.ID, createdTodo.ID)

	if err == nil {
		t.Fatal("Expected error for unauthorized access")
	}

	if err != domain.ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized, got %v", err)
	}
}

func TestTodoService_List_Success(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	service.Create(user.ID, domain.CreateTodoInput{Title: "Todo 1"})
	service.Create(user.ID, domain.CreateTodoInput{Title: "Todo 2"})
	service.Create(user.ID, domain.CreateTodoInput{Title: "Todo 3"})

	filters := domain.TodoFilters{
		Page:     1,
		PageSize: 10,
	}

	todos, total, err := service.List(user.ID, filters)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if total != 3 {
		t.Errorf("Expected 3 todos, got %d", total)
	}

	if len(todos) != 3 {
		t.Errorf("Expected 3 todos in list, got %d", len(todos))
	}
}

func TestTodoService_Update_Success(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title:       "Original Title",
		Description: "Original Description",
		Priority:    domain.TodoPriorityLow,
	}
	createdTodo, _ := service.Create(user.ID, input)

	updateInput := domain.UpdateTodoInput{
		Title:       "Updated Title",
		Description: "Updated Description",
		Priority:    domain.TodoPriorityHigh,
	}

	updatedTodo, err := service.Update(user.ID, createdTodo.ID, updateInput, createdTodo.Version)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedTodo.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", updatedTodo.Title)
	}

	if updatedTodo.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got %s", updatedTodo.Description)
	}

	if updatedTodo.Priority != domain.TodoPriorityHigh {
		t.Errorf("Expected priority %s, got %s", domain.TodoPriorityHigh, updatedTodo.Priority)
	}

	if updatedTodo.Version != 2 {
		t.Errorf("Expected version 2, got %d", updatedTodo.Version)
	}
}

func TestTodoService_Update_VersionMismatch(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}
	createdTodo, _ := service.Create(user.ID, input)

	updateInput := domain.UpdateTodoInput{
		Title: "Updated Title",
	}

	_, err := service.Update(user.ID, createdTodo.ID, updateInput, 999)

	if err == nil {
		t.Fatal("Expected error for version mismatch")
	}

	if err != domain.ErrVersionMismatch {
		t.Errorf("Expected ErrVersionMismatch, got %v", err)
	}
}

func TestTodoService_Update_InvalidStatusTransition(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}
	createdTodo, _ := service.Create(user.ID, input)

	updateInput := domain.UpdateTodoInput{
		Status: domain.TodoStatusCompleted,
	}

	updatedTodo, err := service.Update(user.ID, createdTodo.ID, updateInput, createdTodo.Version)

	if err != nil {
		t.Fatalf("Expected no error for valid transition, got %v", err)
	}

	if updatedTodo.Status != domain.TodoStatusCompleted {
		t.Errorf("Expected status %s, got %s", domain.TodoStatusCompleted, updatedTodo.Status)
	}
}

func TestTodoService_Update_Unauthorized(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	owner := &domain.User{
		Email:       "owner@example.com",
		DisplayName: "Owner",
		IsActive:    true,
	}
	userRepo.Create(owner)

	otherUser := &domain.User{
		Email:       "other@example.com",
		DisplayName: "Other",
		IsActive:    true,
	}
	userRepo.Create(otherUser)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}
	createdTodo, _ := service.Create(owner.ID, input)

	updateInput := domain.UpdateTodoInput{
		Title: "Updated Title",
	}

	_, err := service.Update(otherUser.ID, createdTodo.ID, updateInput, createdTodo.Version)

	if err == nil {
		t.Fatal("Expected error for unauthorized update")
	}

	if err != domain.ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized, got %v", err)
	}
}

func TestTodoService_Delete_Success(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}
	createdTodo, _ := service.Create(user.ID, input)

	err := service.Delete(user.ID, createdTodo.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, err = service.Get(user.ID, createdTodo.ID)
	if err == nil {
		t.Fatal("Expected error for deleted todo")
	}
}

func TestTodoService_Delete_NotFound(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	err := service.Delete(user.ID, uuid.New())

	if err == nil {
		t.Fatal("Expected error for non-existent todo")
	}

	if err != domain.ErrTodoNotFound {
		t.Errorf("Expected ErrTodoNotFound, got %v", err)
	}
}

func TestTodoService_Complete_Success(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}
	createdTodo, _ := service.Create(user.ID, input)

	updatedTodo, err := service.Complete(user.ID, createdTodo.ID, createdTodo.Version)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedTodo.Status != domain.TodoStatusCompleted {
		t.Errorf("Expected status %s, got %s", domain.TodoStatusCompleted, updatedTodo.Status)
	}
}

func TestTodoService_Complete_AlreadyCompleted(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}
	todo, _ := service.Create(user.ID, input)

	updateInput := domain.UpdateTodoInput{Status: domain.TodoStatusCompleted}
	service.Update(user.ID, todo.ID, updateInput, todo.Version)
	todo, _ = service.Get(user.ID, todo.ID)

	_, err := service.Complete(user.ID, todo.ID, todo.Version)

	if err == nil {
		t.Fatal("Expected error for already completed todo")
	}

	if err != domain.ErrInvalidStatusTransition {
		t.Errorf("Expected ErrInvalidStatusTransition, got %v", err)
	}
}

func TestTodoService_Complete_VersionMismatch(t *testing.T) {
	service, _, _, userRepo := setupTestTodoService(t)

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsActive:    true,
	}
	userRepo.Create(user)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}
	createdTodo, _ := service.Create(user.ID, input)

	_, err := service.Complete(user.ID, createdTodo.ID, 999)

	if err == nil {
		t.Fatal("Expected error for version mismatch")
	}

	if err != domain.ErrVersionMismatch {
		t.Errorf("Expected ErrVersionMismatch, got %v", err)
	}
}

func TestTodoService_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from     domain.TodoStatus
		to       domain.TodoStatus
		expected bool
	}{
		{domain.TodoStatusPending, domain.TodoStatusInProgress, true},
		{domain.TodoStatusPending, domain.TodoStatusCompleted, true},
		{domain.TodoStatusInProgress, domain.TodoStatusCompleted, true},
		{domain.TodoStatusInProgress, domain.TodoStatusPending, true},
		{domain.TodoStatusCompleted, domain.TodoStatusInProgress, true},
		{domain.TodoStatusCompleted, domain.TodoStatusPending, true},
	}

	for _, tt := range tests {
		todo := &domain.Todo{Status: tt.from}
		if got := todo.CanTransitionTo(tt.to); got != tt.expected {
			t.Errorf("CanTransitionTo(%s -> %s) = %v, want %v", tt.from, tt.to, got, tt.expected)
		}
	}
}

func TestTodoService_Assign_Success(t *testing.T) {
	service, _, connRepo, userRepo := setupTestTodoService(t)

	owner := &domain.User{
		Email:       "owner@example.com",
		DisplayName: "Owner",
		IsActive:    true,
	}
	userRepo.Create(owner)

	assignee := &domain.User{
		Email:       "assignee@example.com",
		DisplayName: "Assignee",
		IsActive:    true,
	}
	userRepo.Create(assignee)

	conn := &domain.Connection{
		UserAID:     owner.ID,
		UserBID:     assignee.ID,
		Status:      domain.ConnectionStatusAccepted,
		RequestedBy: owner.ID,
	}
	connRepo.Create(conn)

	input := domain.CreateTodoInput{
		Title: "Test Todo",
	}
	createdTodo, _ := service.Create(owner.ID, input)

	updatedTodo, err := service.Assign(createdTodo.ID, owner.ID, assignee.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedTodo.AssignedTo == nil || *updatedTodo.AssignedTo != assignee.ID {
		t.Errorf("Expected assigned_to %s, got %v", assignee.ID, updatedTodo.AssignedTo)
	}
}
