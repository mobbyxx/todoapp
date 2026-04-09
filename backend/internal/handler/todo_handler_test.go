package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/middleware"
)

type mockTodoService struct {
	todos        map[uuid.UUID]*domain.Todo
	connections  map[uuid.UUID][]uuid.UUID
}

func newMockTodoService() *mockTodoService {
	return &mockTodoService{
		todos:       make(map[uuid.UUID]*domain.Todo),
		connections: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *mockTodoService) Create(userID uuid.UUID, input domain.CreateTodoInput) (*domain.Todo, error) {
	todo := &domain.Todo{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		Status:      domain.TodoStatusPending,
		Priority:    input.Priority,
		CreatedBy:   userID,
		AssignedTo:  input.AssignedTo,
		DueDate:     input.DueDate,
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

func (m *mockTodoService) Get(userID uuid.UUID, todoID uuid.UUID) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}

	if !todo.IsAccessibleBy(userID) {
		return nil, domain.ErrUnauthorized
	}

	return todo, nil
}

func (m *mockTodoService) List(userID uuid.UUID, filters domain.TodoFilters) ([]*domain.Todo, int, error) {
	var result []*domain.Todo
	for _, todo := range m.todos {
		if todo.IsAccessibleBy(userID) {
			if filters.Status != nil && todo.Status != *filters.Status {
				continue
			}
			if filters.Priority != nil && todo.Priority != *filters.Priority {
				continue
			}
			result = append(result, todo)
		}
	}
	return result, len(result), nil
}

func (m *mockTodoService) Update(userID uuid.UUID, todoID uuid.UUID, input domain.UpdateTodoInput, version int) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}

	if !todo.IsOwnedBy(userID) {
		return nil, domain.ErrUnauthorized
	}

	if todo.Version != version {
		return nil, domain.ErrVersionMismatch
	}

	if input.Status != "" && !todo.CanTransitionTo(input.Status) {
		return nil, domain.ErrInvalidStatusTransition
	}

	if input.Title != "" {
		todo.Title = input.Title
	}
	if input.Description != "" {
		todo.Description = input.Description
	}
	if input.Status != "" {
		todo.Status = input.Status
	}
	if input.Priority != "" {
		todo.Priority = input.Priority
	}
	if input.AssignedTo != nil {
		todo.AssignedTo = input.AssignedTo
	}
	if input.DueDate != nil {
		todo.DueDate = input.DueDate
	}

	todo.Version++
	todo.UpdatedAt = time.Now()
	m.todos[todo.ID] = todo

	return todo, nil
}

func (m *mockTodoService) Delete(userID uuid.UUID, todoID uuid.UUID) error {
	todo, ok := m.todos[todoID]
	if !ok {
		return domain.ErrTodoNotFound
	}

	if !todo.IsOwnedBy(userID) {
		return domain.ErrUnauthorized
	}

	delete(m.todos, todoID)
	return nil
}

func (m *mockTodoService) Assign(todoID uuid.UUID, userID uuid.UUID, assignToID uuid.UUID) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}

	if !todo.IsOwnedBy(userID) {
		return nil, domain.ErrUnauthorized
	}

	todo.AssignedTo = &assignToID
	todo.UpdatedAt = time.Now()
	m.todos[todo.ID] = todo

	return todo, nil
}

func (m *mockTodoService) Complete(userID uuid.UUID, todoID uuid.UUID, version int) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}

	if !todo.IsAccessibleBy(userID) {
		return nil, domain.ErrUnauthorized
	}

	if todo.Version != version {
		return nil, domain.ErrVersionMismatch
	}

	if !todo.CanTransitionTo(domain.TodoStatusCompleted) {
		return nil, domain.ErrInvalidStatusTransition
	}

	todo.Status = domain.TodoStatusCompleted
	todo.Version++
	todo.UpdatedAt = time.Now()
	m.todos[todo.ID] = todo

	return todo, nil
}

func setupTodoHandlerTest(t *testing.T) (*TodoHandler, *mockTodoService) {
	mockService := newMockTodoService()
	handler := NewTodoHandler(mockService)
	return handler, mockService
}

func TestTodoHandler_Create_Success(t *testing.T) {
	handler, _ := setupTodoHandlerTest(t)

	userID := uuid.New()
	reqBody := map[string]interface{}{
		"title":       "Test Todo",
		"description": "Test Description",
		"priority":    "high",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var resp todoResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Todo.Title != "Test Todo" {
		t.Errorf("Expected title 'Test Todo', got %s", resp.Todo.Title)
	}

	if resp.Todo.Priority != "high" {
		t.Errorf("Expected priority 'high', got %s", resp.Todo.Priority)
	}
}

func TestTodoHandler_Create_InvalidTitle(t *testing.T) {
	handler, _ := setupTodoHandlerTest(t)

	userID := uuid.New()
	reqBody := map[string]interface{}{
		"title": "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestTodoHandler_Create_Unauthorized(t *testing.T) {
	handler, _ := setupTodoHandlerTest(t)

	reqBody := map[string]interface{}{
		"title": "Test Todo",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestTodoHandler_List_Success(t *testing.T) {
	handler, service := setupTodoHandlerTest(t)

	userID := uuid.New()
	_, _ = service.Create(userID, domain.CreateTodoInput{Title: "Todo 1"})
	_, _ = service.Create(userID, domain.CreateTodoInput{Title: "Todo 2"})

	req := httptest.NewRequest("GET", "/api/v1/todos", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp todoListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.TotalCount != 2 {
		t.Errorf("Expected 2 todos, got %d", resp.TotalCount)
	}

	if len(resp.Todos) != 2 {
		t.Errorf("Expected 2 todos in list, got %d", len(resp.Todos))
	}
}

func TestTodoHandler_Get_Success(t *testing.T) {
	handler, service := setupTodoHandlerTest(t)

	userID := uuid.New()
	todo, _ := service.Create(userID, domain.CreateTodoInput{Title: "Test Todo"})

	req := httptest.NewRequest("GET", "/api/v1/todos/"+todo.ID.String(), nil)
	router := chi.NewRouter()
	router.Get("/api/v1/todos/{id}", handler.Get)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp todoResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Todo.ID != todo.ID.String() {
		t.Errorf("Expected ID %s, got %s", todo.ID.String(), resp.Todo.ID)
	}
}

func TestTodoHandler_Get_NotFound(t *testing.T) {
	handler, _ := setupTodoHandlerTest(t)

	userID := uuid.New()
	req := httptest.NewRequest("GET", "/api/v1/todos/"+uuid.New().String(), nil)
	router := chi.NewRouter()
	router.Get("/api/v1/todos/{id}", handler.Get)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestTodoHandler_Update_Success(t *testing.T) {
	handler, service := setupTodoHandlerTest(t)

	userID := uuid.New()
	todo, _ := service.Create(userID, domain.CreateTodoInput{Title: "Original Title"})

	reqBody := map[string]interface{}{
		"title":   "Updated Title",
		"version": todo.Version,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/v1/todos/"+todo.ID.String(), bytes.NewBuffer(jsonBody))
	router := chi.NewRouter()
	router.Put("/api/v1/todos/{id}", handler.Update)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp todoResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Todo.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", resp.Todo.Title)
	}
}

func TestTodoHandler_Update_VersionMismatch(t *testing.T) {
	handler, service := setupTodoHandlerTest(t)

	userID := uuid.New()
	todo, _ := service.Create(userID, domain.CreateTodoInput{Title: "Original Title"})

	reqBody := map[string]interface{}{
		"title":   "Updated Title",
		"version": 999,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/v1/todos/"+todo.ID.String(), bytes.NewBuffer(jsonBody))
	router := chi.NewRouter()
	router.Put("/api/v1/todos/{id}", handler.Update)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}

func TestTodoHandler_Delete_Success(t *testing.T) {
	handler, service := setupTodoHandlerTest(t)

	userID := uuid.New()
	todo, _ := service.Create(userID, domain.CreateTodoInput{Title: "Test Todo"})

	req := httptest.NewRequest("DELETE", "/api/v1/todos/"+todo.ID.String(), nil)
	router := chi.NewRouter()
	router.Delete("/api/v1/todos/{id}", handler.Delete)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	_, err := service.Get(userID, todo.ID)
	if err == nil {
		t.Error("Expected todo to be deleted")
	}
}

func TestTodoHandler_Delete_NotFound(t *testing.T) {
	handler, _ := setupTodoHandlerTest(t)

	userID := uuid.New()
	req := httptest.NewRequest("DELETE", "/api/v1/todos/"+uuid.New().String(), nil)
	router := chi.NewRouter()
	router.Delete("/api/v1/todos/{id}", handler.Delete)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestTodoHandler_Complete_Success(t *testing.T) {
	handler, service := setupTodoHandlerTest(t)

	userID := uuid.New()
	input := domain.CreateTodoInput{
		Title:  "Test Todo",
		Status: domain.TodoStatusInProgress,
	}
	todo, _ := service.Create(userID, input)

	reqBody := map[string]interface{}{
		"version": todo.Version,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/todos/"+todo.ID.String()+"/complete", bytes.NewBuffer(jsonBody))
	router := chi.NewRouter()
	router.Post("/api/v1/todos/{id}/complete", handler.Complete)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp todoResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Todo.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", resp.Todo.Status)
	}
}

func TestTodoHandler_Complete_VersionMismatch(t *testing.T) {
	handler, service := setupTodoHandlerTest(t)

	userID := uuid.New()
	input := domain.CreateTodoInput{
		Title:  "Test Todo",
		Status: domain.TodoStatusInProgress,
	}
	todo, _ := service.Create(userID, input)

	reqBody := map[string]interface{}{
		"version": 999,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/todos/"+todo.ID.String()+"/complete", bytes.NewBuffer(jsonBody))
	router := chi.NewRouter()
	router.Post("/api/v1/todos/{id}/complete", handler.Complete)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}

func TestTodoHandler_Complete_InvalidStatusTransition(t *testing.T) {
	handler, service := setupTodoHandlerTest(t)

	userID := uuid.New()
	input := domain.CreateTodoInput{
		Title:  "Test Todo",
		Status: domain.TodoStatusCompleted,
	}
	todo, _ := service.Create(userID, input)

	reqBody := map[string]interface{}{
		"version": todo.Version,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/todos/"+todo.ID.String()+"/complete", bytes.NewBuffer(jsonBody))
	router := chi.NewRouter()
	router.Post("/api/v1/todos/{id}/complete", handler.Complete)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID.String())
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestTodoHandler_Routes(t *testing.T) {
	handler, _ := setupTodoHandlerTest(t)

	router := handler.Routes()

	if router == nil {
		t.Fatal("Expected Routes to return a router")
	}
}
