package benchmark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/handler"
	"github.com/user/todo-api/internal/middleware"
)

func BenchmarkTodoHandler_Create(b *testing.B) {
	mockService := newMockTodoService()
	h := handler.NewTodoHandler(mockService)

	userID := uuid.New().String()
	reqBody := map[string]interface{}{
		"title":       "Benchmark Todo",
		"description": "Benchmark Description",
		"priority":    "high",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/todos", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusCreated {
			b.Fatalf("Expected status %d, got %d", http.StatusCreated, rr.Code)
		}
	}
}

func BenchmarkTodoHandler_List(b *testing.B) {
	mockService := newMockTodoService()
	h := handler.NewTodoHandler(mockService)
	userID := uuid.New()

	for i := 0; i < 100; i++ {
		mockService.Create(userID, domain.CreateTodoInput{
			Title: fmt.Sprintf("Todo %d", i),
		})
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/todos", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
	}
}

func BenchmarkTodoHandler_Get(b *testing.B) {
	mockService := newMockTodoService()
	h := handler.NewTodoHandler(mockService)
	userID := uuid.New()
	todo, _ := mockService.Create(userID, domain.CreateTodoInput{Title: "Test Todo"})

	router := chi.NewRouter()
	router.Get("/api/v1/todos/{id}", h.Get)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/todos/"+todo.ID.String(), nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
	}
}

func BenchmarkTodoHandler_Update(b *testing.B) {
	mockService := newMockTodoService()
	h := handler.NewTodoHandler(mockService)
	userID := uuid.New()
	todo, _ := mockService.Create(userID, domain.CreateTodoInput{Title: "Original Title"})

	router := chi.NewRouter()
	router.Put("/api/v1/todos/{id}", h.Update)

	reqBody := map[string]interface{}{
		"title":   "Updated Title",
		"version": 1,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/api/v1/todos/"+todo.ID.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
	}
}

func BenchmarkUserHandler_Register(b *testing.B) {
	mockUserService := newMockUserService()
	mockJWTService := newMockJWTService()
	h := handler.NewUserHandler(mockUserService, mockJWTService)

	reqBody := map[string]interface{}{
		"email":         "benchmark@example.com",
		"password":      "password123",
		"display_name":  "Benchmark User",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reqBody["email"] = fmt.Sprintf("benchmark_%d@example.com", i)
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusCreated {
			b.Fatalf("Expected status %d, got %d", http.StatusCreated, rr.Code)
		}
	}
}

func BenchmarkUserHandler_Login(b *testing.B) {
	mockUserService := newMockUserService()
	mockJWTService := newMockJWTService()
	h := handler.NewUserHandler(mockUserService, mockJWTService)

	_, _ = mockUserService.Register("login@example.com", "password123", "Login User")

	reqBody := map[string]interface{}{
		"email":    "login@example.com",
		"password": "password123",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
	}
}

func BenchmarkDomain_TodoCreation(b *testing.B) {
	userID := uuid.New()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		todo := &domain.Todo{
			ID:          uuid.New(),
			Title:       fmt.Sprintf("Todo %d", i),
			Description: "Description",
			Status:      domain.TodoStatusPending,
			Priority:    domain.TodoPriorityMedium,
			CreatedBy:   userID,
			Version:     1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		_ = todo.ToResponse()
	}
}

func BenchmarkDomain_StatusValidation(b *testing.B) {
	statuses := []domain.TodoStatus{
		domain.TodoStatusPending,
		domain.TodoStatusInProgress,
		domain.TodoStatusCompleted,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		status := statuses[i%len(statuses)]
		_ = domain.ValidateStatus(status)
	}
}

func BenchmarkDomain_PriorityValidation(b *testing.B) {
	priorities := []domain.TodoPriority{
		domain.TodoPriorityLow,
		domain.TodoPriorityMedium,
		domain.TodoPriorityHigh,
		domain.TodoPriorityUrgent,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		priority := priorities[i%len(priorities)]
		_ = domain.ValidatePriority(priority)
	}
}

func BenchmarkJSON_Serialization(b *testing.B) {
	todo := domain.TodoResponse{
		ID:          uuid.New().String(),
		Title:       "Benchmark Todo",
		Description: "Description",
		Status:      "pending",
		Priority:    "medium",
		CreatedBy:   uuid.New().String(),
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(todo)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSON_Deserialization(b *testing.B) {
	jsonData := `{
		"id": "550e8400-e29b-41d4-a716-446655440000",
		"title": "Benchmark Todo",
		"description": "Description",
		"status": "pending",
		"priority": "medium",
		"created_by": "550e8400-e29b-41d4-a716-446655440001",
		"version": 1,
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z"
	}`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var todo domain.TodoResponse
		err := json.Unmarshal([]byte(jsonData), &todo)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUUID_Generation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = uuid.New()
	}
}

func BenchmarkUUID_Parsing(b *testing.B) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := uuid.Parse(uuidStr)
		if err != nil {
			b.Fatal(err)
		}
	}
}

type mockTodoServiceBench struct {
	todos map[uuid.UUID]*domain.Todo
}

func newMockTodoService() *mockTodoServiceBench {
	return &mockTodoServiceBench{
		todos: make(map[uuid.UUID]*domain.Todo),
	}
}

func (m *mockTodoServiceBench) Create(userID uuid.UUID, input domain.CreateTodoInput) (*domain.Todo, error) {
	todo := &domain.Todo{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		Status:      domain.TodoStatusPending,
		Priority:    input.Priority,
		CreatedBy:   userID,
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if todo.Priority == "" {
		todo.Priority = domain.TodoPriorityMedium
	}
	m.todos[todo.ID] = todo
	return todo, nil
}

func (m *mockTodoServiceBench) Get(userID uuid.UUID, todoID uuid.UUID) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	return todo, nil
}

func (m *mockTodoServiceBench) List(userID uuid.UUID, filters domain.TodoFilters) ([]*domain.Todo, int, error) {
	var result []*domain.Todo
	for _, todo := range m.todos {
		result = append(result, todo)
	}
	return result, len(result), nil
}

func (m *mockTodoServiceBench) Update(userID uuid.UUID, todoID uuid.UUID, input domain.UpdateTodoInput, version int) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	if input.Title != "" {
		todo.Title = input.Title
	}
	todo.Version++
	todo.UpdatedAt = time.Now()
	return todo, nil
}

func (m *mockTodoServiceBench) Delete(userID uuid.UUID, todoID uuid.UUID) error {
	delete(m.todos, todoID)
	return nil
}

func (m *mockTodoServiceBench) Assign(todoID uuid.UUID, userID uuid.UUID, assignToID uuid.UUID) (*domain.Todo, error) {
	return nil, nil
}

func (m *mockTodoServiceBench) Complete(userID uuid.UUID, todoID uuid.UUID, version int) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	todo.Status = domain.TodoStatusCompleted
	todo.Version++
	return todo, nil
}

type mockUserServiceBench struct {
	users map[string]*domain.User
}

func newMockUserService() *mockUserServiceBench {
	return &mockUserServiceBench{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserServiceBench) Register(email, password, displayName string) (*domain.User, error) {
	user := &domain.User{
		ID:          uuid.New(),
		Email:       email,
		DisplayName: displayName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.users[user.ID.String()] = user
	return user, nil
}

func (m *mockUserServiceBench) Login(email, password string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.ErrInvalidCredentials
}

func (m *mockUserServiceBench) GetUser(id uuid.UUID) (*domain.User, error) {
	user, ok := m.users[id.String()]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserServiceBench) UpdateProfile(id uuid.UUID, displayName string) (*domain.User, error) {
	return nil, nil
}

type mockJWTServiceBench struct{}

func newMockJWTService() *mockJWTServiceBench {
	return &mockJWTServiceBench{}
}

func (m *mockJWTServiceBench) GenerateTokenPair(ctx context.Context, userID string) (*TokenPairBench, error) {
	return &TokenPairBench{
		AccessToken:  "mock_token",
		RefreshToken: "mock_refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}

func (m *mockJWTServiceBench) ValidateRefreshToken(ctx context.Context, token string) (*TokenPairBench, error) {
	return m.GenerateTokenPair(ctx, "")
}

func (m *mockJWTServiceBench) RevokeToken(ctx context.Context, token string) error {
	return nil
}

type TokenPairBench struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}
