package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/user/todo-api/internal/domain"
)

func setupTestUserService(t *testing.T) (domain.UserService, *JWTService, *mockUserRepository, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockRepo := newMockUserRepository()
	jwtService := NewJWTService("test-secret-key-for-jwt-signing", rdb)
	userService := NewUserService(mockRepo, jwtService)

	return userService, jwtService, mockRepo, mr
}

type mockUserRepository struct {
	users   map[uuid.UUID]*domain.User
	byEmail map[string]*domain.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:   make(map[uuid.UUID]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

func (m *mockUserRepository) Create(user *domain.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	if user, ok := m.users[id]; ok && user.IsActive {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) GetByEmail(email string) (*domain.User, error) {
	if user, ok := m.byEmail[email]; ok && user.IsActive {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) Update(user *domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return domain.ErrUserNotFound
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) Delete(id uuid.UUID) error {
	if user, ok := m.users[id]; ok {
		user.IsActive = false
		user.UpdatedAt = time.Now()
		return nil
	}
	return domain.ErrUserNotFound
}

func (m *mockUserRepository) UpdateLastSeen(id uuid.UUID) error {
	if user, ok := m.users[id]; ok {
		now := time.Now()
		user.LastSeenAt = &now
		user.UpdatedAt = now
		return nil
	}
	return domain.ErrUserNotFound
}

func TestUserService_Register_Success(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	user, err := service.Register(email, password, displayName)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("Expected user, got nil")
	}

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}

	if user.DisplayName != displayName {
		t.Errorf("Expected display name %s, got %s", displayName, user.DisplayName)
	}

	if user.PasswordHash == "" {
		t.Error("Expected password hash to be set")
	}

	if !user.IsActive {
		t.Error("Expected user to be active")
	}
}

func TestUserService_Register_DuplicateEmail(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	_, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("First registration should succeed: %v", err)
	}

	_, err = service.Register(email, "D1fferent!Pass#2024", "Different Name")
	if err == nil {
		t.Fatal("Expected error for duplicate email, got nil")
	}

	if err != domain.ErrEmailTaken {
		t.Errorf("Expected ErrEmailTaken, got %v", err)
	}
}

func TestUserService_Register_InvalidEmail(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	_, err := service.Register("invalid-email", "password123", "Test User")
	if err == nil {
		t.Fatal("Expected error for invalid email")
	}

	if err != domain.ErrValidation {
		t.Errorf("Expected ErrValidation, got %v", err)
	}
}

func TestUserService_Register_PasswordTooShort(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	_, err := service.Register("test@example.com", "short", "Test User")
	if err == nil {
		t.Fatal("Expected error for short password")
	}

	if err != domain.ErrValidation {
		t.Errorf("Expected ErrValidation, got %v", err)
	}
}

func TestUserService_Register_DisplayNameTooShort(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	_, err := service.Register("test@example.com", "password123", "A")
	if err == nil {
		t.Fatal("Expected error for short display name")
	}

	if err != domain.ErrValidation {
		t.Errorf("Expected ErrValidation, got %v", err)
	}
}

func TestUserService_Register_DisplayNameTooLong(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	longName := ""
	for i := 0; i < 51; i++ {
		longName += "a"
	}

	_, err := service.Register("test@example.com", "password123", longName)
	if err == nil {
		t.Fatal("Expected error for long display name")
	}

	if err != domain.ErrValidation {
		t.Errorf("Expected ErrValidation, got %v", err)
	}
}

func TestUserService_Login_Success(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	_, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	user, err := service.Login(email, password)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("Expected user, got nil")
	}

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}

	if user.LastSeenAt == nil {
		t.Error("Expected LastSeenAt to be updated")
	}
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	_, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	_, err = service.Login(email, "wrongpassword")
	if err == nil {
		t.Fatal("Expected error for invalid password")
	}

	if err != domain.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	_, err := service.Login("nonexistent@example.com", "password123")
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}

	if err != domain.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_GetUser_Success(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	createdUser, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	user, err := service.GetUser(createdUser.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("Expected user, got nil")
	}

	if user.ID != createdUser.ID {
		t.Errorf("Expected user ID %s, got %s", createdUser.ID, user.ID)
	}
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	_, err := service.GetUser(uuid.New())
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}

	if err != domain.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_UpdateProfile_Success(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	createdUser, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	newDisplayName := "Updated Name"
	user, err := service.UpdateProfile(createdUser.ID, newDisplayName)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("Expected user, got nil")
	}

	if user.DisplayName != newDisplayName {
		t.Errorf("Expected display name %s, got %s", newDisplayName, user.DisplayName)
	}
}

func TestUserService_UpdateProfile_DisplayNameTooShort(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	createdUser, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	_, err = service.UpdateProfile(createdUser.ID, "A")
	if err == nil {
		t.Fatal("Expected error for short display name")
	}

	if err != domain.ErrInvalidDisplayName {
		t.Errorf("Expected ErrInvalidDisplayName, got %v", err)
	}
}

func TestUserService_UpdateProfile_DisplayNameTooLong(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	createdUser, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	longName := ""
	for i := 0; i < 51; i++ {
		longName += "a"
	}

	_, err = service.UpdateProfile(createdUser.ID, longName)
	if err == nil {
		t.Fatal("Expected error for long display name")
	}

	if err != domain.ErrInvalidDisplayName {
		t.Errorf("Expected ErrInvalidDisplayName, got %v", err)
	}
}

func TestUserService_UpdateProfile_UserNotFound(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	_, err := service.UpdateProfile(uuid.New(), "New Name")
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}

	if err != domain.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_SoftDelete_Success(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	createdUser, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	err = service.SoftDelete(createdUser.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, err = service.GetUser(createdUser.ID)
	if err == nil {
		t.Fatal("Expected error for deleted user")
	}

	if err != domain.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_SoftDelete_UserNotFound(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	err := service.SoftDelete(uuid.New())
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}

	if err != domain.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_UpdateLastSeen_Success(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	createdUser, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	err = service.UpdateLastSeen(createdUser.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	user, err := service.GetUser(createdUser.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user.LastSeenAt == nil {
		t.Error("Expected LastSeenAt to be updated")
	}
}

func TestUserService_GetUserByEmail_Success(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	_, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	user, err := service.GetUserByEmail(email)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("Expected user, got nil")
	}

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
}

func TestUserService_GetUserByEmail_NotFound(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	_, err := service.GetUserByEmail("nonexistent@example.com")
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}

	if err != domain.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_PasswordHashing(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	user, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	if user.PasswordHash == password {
		t.Error("Password should be hashed, not stored in plain text")
	}

	if len(user.PasswordHash) < 50 {
		t.Error("Password hash seems too short for bcrypt")
	}
}

func TestUserService_EmailNormalization(t *testing.T) {
	service, _, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "Test.User@Example.COM"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	user, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	if user.Email != "test.user@example.com" {
		t.Errorf("Expected normalized email test.user@example.com, got %s", user.Email)
	}

	_, err = service.Login("TEST.USER@EXAMPLE.COM", password)
	if err != nil {
		t.Errorf("Login with different case email should work: %v", err)
	}
}

func TestUserService_TokenGenerationOnRegister(t *testing.T) {
	service, jwtSvc, _, mr := setupTestUserService(t)
	defer mr.Close()

	email := "test@example.com"
	password := "S3cure!Pass#2024"
	displayName := "Test User"

	user, err := service.Register(email, password, displayName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	tokenPair, err := jwtSvc.GenerateTokenPair(context.Background(), user.ID.String())
	if err != nil {
		t.Fatalf("Token generation failed: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("Expected access token to be generated")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("Expected refresh token to be generated")
	}

	if tokenPair.AccessToken == tokenPair.RefreshToken {
		t.Error("Access and refresh tokens should be different")
	}
}
