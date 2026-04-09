package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/middleware"
	"github.com/user/todo-api/internal/service"
)

func setupUserHandlerTest(t *testing.T) (*UserHandler, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockRepo := newMockUserRepo()
	jwtService := service.NewJWTService("test-secret-key-for-jwt-signing", rdb)
	userService := service.NewUserService(mockRepo, jwtService)
	handler := NewUserHandler(userService, jwtService)

	return handler, mr
}

type mockUserRepo struct {
	users   map[uuid.UUID]*domain.User
	byEmail map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:   make(map[uuid.UUID]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) Create(user *domain.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByID(id uuid.UUID) (*domain.User, error) {
	if user, ok := m.users[id]; ok && user.IsActive {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) GetByEmail(email string) (*domain.User, error) {
	if user, ok := m.byEmail[email]; ok && user.IsActive {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) Update(user *domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return domain.ErrUserNotFound
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *mockUserRepo) Delete(id uuid.UUID) error {
	if user, ok := m.users[id]; ok {
		user.IsActive = false
		user.UpdatedAt = time.Now()
		return nil
	}
	return domain.ErrUserNotFound
}

func (m *mockUserRepo) UpdateLastSeen(id uuid.UUID) error {
	if user, ok := m.users[id]; ok {
		now := time.Now()
		user.LastSeenAt = &now
		user.UpdatedAt = now
		return nil
	}
	return domain.ErrUserNotFound
}

func TestUserHandler_Register_Success(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	reqBody := map[string]string{
		"email":        "test@example.com",
		"password":     "securepassword123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var resp authResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.User.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", resp.User.Email)
	}

	if resp.AccessToken == "" {
		t.Error("Expected access token to be present")
	}

	if resp.RefreshToken == "" {
		t.Error("Expected refresh token to be present")
	}
}

func TestUserHandler_Register_DuplicateEmail(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	reqBody := map[string]string{
		"email":        "test@example.com",
		"password":     "securepassword123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req1 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	handler.Register(rr1, req1)

	if rr1.Code != http.StatusCreated {
		t.Fatalf("First registration failed: %d - %s", rr1.Code, rr1.Body.String())
	}

	req2 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	handler.Register(rr2, req2)

	if rr2.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, rr2.Code)
	}
}

func TestUserHandler_Register_InvalidEmail(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	reqBody := map[string]string{
		"email":        "invalid-email",
		"password":     "securepassword123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUserHandler_Register_PasswordTooShort(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	reqBody := map[string]string{
		"email":        "test@example.com",
		"password":     "short",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUserHandler_Login_Success(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	registerBody := map[string]string{
		"email":        "test@example.com",
		"password":     "securepassword123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(registerBody)

	req1 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	handler.Register(rr1, req1)

	if rr1.Code != http.StatusCreated {
		t.Fatalf("Registration failed: %d - %s", rr1.Code, rr1.Body.String())
	}

	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "securepassword123",
	}
	jsonBody, _ = json.Marshal(loginBody)

	req2 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	handler.Login(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, rr2.Code, rr2.Body.String())
	}

	var resp authResponse
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("Expected access token to be present")
	}

	if resp.RefreshToken == "" {
		t.Error("Expected refresh token to be present")
	}
}

func TestUserHandler_Login_InvalidCredentials(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	registerBody := map[string]string{
		"email":        "test@example.com",
		"password":     "securepassword123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(registerBody)

	req1 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	handler.Register(rr1, req1)

	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ = json.Marshal(loginBody)

	req2 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	handler.Login(rr2, req2)

	if rr2.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr2.Code)
	}
}

func TestUserHandler_Login_UserNotFound(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	loginBody := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(loginBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestUserHandler_Refresh_Success(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	registerBody := map[string]string{
		"email":        "test@example.com",
		"password":     "securepassword123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(registerBody)

	req1 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	handler.Register(rr1, req1)

	if rr1.Code != http.StatusCreated {
		t.Fatalf("Registration failed: %d - %s", rr1.Code, rr1.Body.String())
	}

	var registerResp authResponse
	json.Unmarshal(rr1.Body.Bytes(), &registerResp)

	refreshBody := map[string]string{
		"refresh_token": registerResp.RefreshToken,
	}
	jsonBody, _ = json.Marshal(refreshBody)

	req2 := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	handler.Refresh(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, rr2.Code, rr2.Body.String())
	}

	var refreshResp map[string]interface{}
	if err := json.Unmarshal(rr2.Body.Bytes(), &refreshResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if refreshResp["access_token"] == nil || refreshResp["access_token"] == "" {
		t.Error("Expected new access token to be present")
	}

	if refreshResp["refresh_token"] == nil || refreshResp["refresh_token"] == "" {
		t.Error("Expected new refresh token to be present")
	}
}

func TestUserHandler_Refresh_InvalidToken(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	refreshBody := map[string]string{
		"refresh_token": "invalid.token.here",
	}
	jsonBody, _ := json.Marshal(refreshBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestUserHandler_Logout_Success(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	registerBody := map[string]string{
		"email":        "test@example.com",
		"password":     "securepassword123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(registerBody)

	req1 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	handler.Register(rr1, req1)

	if rr1.Code != http.StatusCreated {
		t.Fatalf("Registration failed: %d - %s", rr1.Code, rr1.Body.String())
	}

	var registerResp authResponse
	json.Unmarshal(rr1.Body.Bytes(), &registerResp)

	logoutBody := map[string]string{
		"refresh_token": registerResp.RefreshToken,
	}
	jsonBody, _ = json.Marshal(logoutBody)

	req2 := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	handler.Logout(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, rr2.Code, rr2.Body.String())
	}
}

func TestUserHandler_GetMe_Success(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	registerBody := map[string]string{
		"email":        "test@example.com",
		"password":     "securepassword123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(registerBody)

	req1 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	handler.Register(rr1, req1)

	if rr1.Code != http.StatusCreated {
		t.Fatalf("Registration failed: %d - %s", rr1.Code, rr1.Body.String())
	}

	var registerResp authResponse
	json.Unmarshal(rr1.Body.Bytes(), &registerResp)

	userID := registerResp.User.ID

	req2 := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	ctx := context.WithValue(req2.Context(), middleware.UserIDKey, userID)
	req2 = req2.WithContext(ctx)
	rr2 := httptest.NewRecorder()

	handler.GetMe(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, rr2.Code, rr2.Body.String())
	}

	var resp userResponse
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.User.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", resp.User.Email)
	}
}

func TestUserHandler_GetMe_Unauthorized(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	rr := httptest.NewRecorder()

	handler.GetMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestUserHandler_UpdateMe_Success(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	registerBody := map[string]string{
		"email":        "test@example.com",
		"password":     "securepassword123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(registerBody)

	req1 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	handler.Register(rr1, req1)

	if rr1.Code != http.StatusCreated {
		t.Fatalf("Registration failed: %d - %s", rr1.Code, rr1.Body.String())
	}

	var registerResp authResponse
	json.Unmarshal(rr1.Body.Bytes(), &registerResp)

	userID := registerResp.User.ID

	updateBody := map[string]string{
		"display_name": "Updated Name",
	}
	jsonBody, _ = json.Marshal(updateBody)

	req2 := httptest.NewRequest("PUT", "/api/v1/users/me", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req2.Context(), middleware.UserIDKey, userID)
	req2 = req2.WithContext(ctx)
	rr2 := httptest.NewRecorder()

	handler.UpdateMe(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, rr2.Code, rr2.Body.String())
	}

	var resp userResponse
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.User.DisplayName != "Updated Name" {
		t.Errorf("Expected display name 'Updated Name', got %s", resp.User.DisplayName)
	}
}

func TestUserHandler_UpdateMe_InvalidDisplayName(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	registerBody := map[string]string{
		"email":        "test@example.com",
		"password":     "securepassword123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(registerBody)

	req1 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	handler.Register(rr1, req1)

	var registerResp authResponse
	json.Unmarshal(rr1.Body.Bytes(), &registerResp)

	userID := registerResp.User.ID

	updateBody := map[string]string{
		"display_name": "A",
	}
	jsonBody, _ = json.Marshal(updateBody)

	req2 := httptest.NewRequest("PUT", "/api/v1/users/me", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req2.Context(), middleware.UserIDKey, userID)
	req2 = req2.WithContext(ctx)
	rr2 := httptest.NewRecorder()

	handler.UpdateMe(rr2, req2)

	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr2.Code)
	}
}

func TestUserHandler_Routes(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	router := handler.Routes()

	if router == nil {
		t.Fatal("Expected Routes to return a router")
	}
}

func TestUserHandler_ProtectedRoutes(t *testing.T) {
	handler, mr := setupUserHandlerTest(t)
	defer mr.Close()

	router := handler.ProtectedRoutes()

	if router == nil {
		t.Fatal("Expected ProtectedRoutes to return a router")
	}
}

func TestWriteError(t *testing.T) {
	rr := httptest.NewRecorder()
	writeError(rr, http.StatusBadRequest, "test_error", "Test error message")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type to contain application/json, got %s", contentType)
	}

	var resp errorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != "test_error" {
		t.Errorf("Expected error code 'test_error', got %s", resp.Error)
	}

	if resp.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got %s", resp.Message)
	}
}
