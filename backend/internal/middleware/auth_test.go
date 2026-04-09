package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/service"
)

func setupAuthTest(t *testing.T) (*service.JWTService, *MockAPIKeyService, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	secret := "test-secret-key-for-jwt-signing"
	jwtService := service.NewJWTService(secret, rdb)
	apiKeyService := &MockAPIKeyService{}

	return jwtService, apiKeyService, mr
}

func TestAuthMiddleware_JWT(t *testing.T) {
	jwtService, _, mr := setupAuthTest(t)
	defer mr.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID == "" {
			t.Error("UserID not found in context")
		}
		authType := GetAuthType(r.Context())
		if authType != "jwt" {
			t.Errorf("Expected auth type 'jwt', got '%s'", authType)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := Auth(jwtService, nil)(handler)

	ctx := context.Background()
	userID := "user-123"
	tokenPair, err := jwtService.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidJWT(t *testing.T) {
	jwtService, _, mr := setupAuthTest(t)
	defer mr.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with invalid JWT")
	})

	middleware := Auth(jwtService, nil)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	var resp authResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != "invalid_token" {
		t.Errorf("Expected error 'invalid_token', got '%s'", resp.Error)
	}
}

func TestAuthMiddleware_APIKey(t *testing.T) {
	_, apiKeyService, mr := setupAuthTest(t)
	defer mr.Close()

	testUserID := uuid.MustParse("00000000-0000-0000-0000-000000000456")
	testKey := "ouk_v1_testkey123456789"

	apiKeyService.validateFunc = func(key string) (*domain.APIKey, error) {
		if key == testKey {
			return &domain.APIKey{
				ID:       uuid.New(),
				UserID:   testUserID,
				Scopes:   []string{"read:todos"},
				IsActive: true,
			}, nil
		}
		return nil, domain.ErrAPIKeyInvalid
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID != testUserID.String() {
			t.Errorf("Expected user ID '%s', got '%s'", testUserID.String(), userID)
		}
		authType := GetAuthType(r.Context())
		if authType != "api_key" {
			t.Errorf("Expected auth type 'api_key', got '%s'", authType)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := Auth(nil, apiKeyService)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", testKey)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidAPIKey(t *testing.T) {
	_, apiKeyService, mr := setupAuthTest(t)
	defer mr.Close()

	apiKeyService.validateFunc = func(key string) (*domain.APIKey, error) {
		return nil, domain.ErrAPIKeyInvalid
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with invalid API key")
	})

	middleware := Auth(nil, apiKeyService)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	var resp authResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != "invalid_api_key" {
		t.Errorf("Expected error 'invalid_api_key', got '%s'", resp.Error)
	}
}

func TestAuthMiddleware_ExpiredAPIKey(t *testing.T) {
	_, apiKeyService, mr := setupAuthTest(t)
	defer mr.Close()

	apiKeyService.validateFunc = func(key string) (*domain.APIKey, error) {
		return nil, domain.ErrAPIKeyExpired
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with expired API key")
	})

	middleware := Auth(nil, apiKeyService)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "expired-key")
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	var resp authResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != "api_key_expired" {
		t.Errorf("Expected error 'api_key_expired', got '%s'", resp.Error)
	}
}

func TestAuthMiddleware_RevokedAPIKey(t *testing.T) {
	_, apiKeyService, mr := setupAuthTest(t)
	defer mr.Close()

	apiKeyService.validateFunc = func(key string) (*domain.APIKey, error) {
		return nil, domain.ErrAPIKeyRevoked
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with revoked API key")
	})

	middleware := Auth(nil, apiKeyService)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "revoked-key")
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	var resp authResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != "api_key_revoked" {
		t.Errorf("Expected error 'api_key_revoked', got '%s'", resp.Error)
	}
}

func TestAuthMiddleware_MissingAuth(t *testing.T) {
	jwtService, apiKeyService, mr := setupAuthTest(t)
	defer mr.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without auth")
	})

	middleware := Auth(jwtService, apiKeyService)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	var resp authResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != "missing_auth" {
		t.Errorf("Expected error 'missing_auth', got '%s'", resp.Error)
	}
}

func TestAuthMiddleware_PrefersJWT(t *testing.T) {
	jwtService, apiKeyService, mr := setupAuthTest(t)
	defer mr.Close()

	ctx := context.Background()
	userID := "jwt-user"
	tokenPair, err := jwtService.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	apiKeyService.validateFunc = func(key string) (*domain.APIKey, error) {
		return &domain.APIKey{
			ID:       uuid.New(),
			UserID:   uuid.MustParse("00000000-0000-0000-0000-00000000a01c"),
			Scopes:   []string{"read:todos"},
			IsActive: true,
		}, nil
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID != "jwt-user" {
			t.Errorf("Expected JWT user ID 'jwt-user', got '%s'", userID)
		}
		authType := GetAuthType(r.Context())
		if authType != "jwt" {
			t.Errorf("Expected auth type 'jwt', got '%s'", authType)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := Auth(jwtService, apiKeyService)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	req.Header.Set("X-API-Key", "valid-api-key")
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestOptionalAuthMiddleware_NoAuth(t *testing.T) {
	jwtService, apiKeyService, mr := setupAuthTest(t)
	defer mr.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID != "" {
			t.Error("UserID should be empty when no auth provided")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := OptionalAuth(jwtService, apiKeyService)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestOptionalAuthMiddleware_WithJWT(t *testing.T) {
	jwtService, _, mr := setupAuthTest(t)
	defer mr.Close()

	ctx := context.Background()
	userID := "user-123"
	tokenPair, err := jwtService.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUserID := GetUserID(r.Context())
		if ctxUserID != userID {
			t.Errorf("Expected user ID '%s', got '%s'", userID, ctxUserID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := OptionalAuth(jwtService, nil)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestRequireScopeMiddleware_HasScope(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireScope("read:todos")(handler)

	ctx := context.WithValue(context.Background(), UserScopesKey, []string{"read:todos", "write:todos"})
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestRequireScopeMiddleware_MissingScope(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without required scope")
	})

	middleware := RequireScope("admin")(handler)

	ctx := context.WithValue(context.Background(), UserScopesKey, []string{"read:todos"})
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}

	var resp authResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != "insufficient_scope" {
		t.Errorf("Expected error 'insufficient_scope', got '%s'", resp.Error)
	}
}

func TestRequireScopeMiddleware_NoAuth(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without auth")
	})

	middleware := RequireScope("read:todos")(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

func TestGetAuthenticatedUserID(t *testing.T) {
	userID := "test-user-123"
	ctx := context.WithValue(context.Background(), UserIDKey, userID)

	result := GetAuthenticatedUserID(ctx)
	if result != userID {
		t.Errorf("GetAuthenticatedUserID() = %v, want %v", result, userID)
	}
}
