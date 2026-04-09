package service

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

// mockAPIKeyRepository is an in-memory implementation for testing
type mockAPIKeyRepository struct {
	mu       sync.RWMutex
	keys     map[uuid.UUID]*domain.APIKey
	byPrefix map[string]*domain.APIKey
	byUser   map[uuid.UUID][]*domain.APIKey
}

func newMockAPIKeyRepository() *mockAPIKeyRepository {
	return &mockAPIKeyRepository{
		keys:     make(map[uuid.UUID]*domain.APIKey),
		byPrefix: make(map[string]*domain.APIKey),
		byUser:   make(map[uuid.UUID][]*domain.APIKey),
	}
}

func (m *mockAPIKeyRepository) Create(apiKey *domain.APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if apiKey.ID == uuid.Nil {
		apiKey.ID = uuid.New()
	}
	apiKey.CreatedAt = time.Now()
	apiKey.UpdatedAt = time.Now()

	// Deep copy to avoid external modification
	keyCopy := *apiKey
	m.keys[apiKey.ID] = &keyCopy
	m.byPrefix[apiKey.KeyPrefix] = &keyCopy

	if _, exists := m.byUser[apiKey.UserID]; !exists {
		m.byUser[apiKey.UserID] = []*domain.APIKey{}
	}
	userKeys := m.byUser[apiKey.UserID]
	userKeys = append(userKeys, &keyCopy)
	m.byUser[apiKey.UserID] = userKeys

	return nil
}

func (m *mockAPIKeyRepository) GetByPrefix(prefix string) (*domain.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if key, exists := m.byPrefix[prefix]; exists {
		keyCopy := *key
		return &keyCopy, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockAPIKeyRepository) GetByID(id uuid.UUID) (*domain.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if key, exists := m.keys[id]; exists {
		keyCopy := *key
		return &keyCopy, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockAPIKeyRepository) GetByUserID(userID uuid.UUID) ([]*domain.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := m.byUser[userID]
	result := make([]*domain.APIKey, len(keys))
	for i, key := range keys {
		keyCopy := *key
		result[i] = &keyCopy
	}
	return result, nil
}

func (m *mockAPIKeyRepository) UpdateLastUsed(id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key, exists := m.keys[id]; exists {
		now := time.Now()
		key.LastUsedAt = &now
		return nil
	}
	return domain.ErrNotFound
}

func (m *mockAPIKeyRepository) Revoke(id uuid.UUID, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key, exists := m.keys[id]; exists {
		now := time.Now()
		key.IsActive = false
		key.RevokedAt = &now
		key.RevokedReason = reason
		return nil
	}
	return domain.ErrNotFound
}

func (m *mockAPIKeyRepository) Delete(id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key, exists := m.keys[id]; exists {
		delete(m.keys, id)
		delete(m.byPrefix, key.KeyPrefix)

		// Remove from user's key list
		if userKeys, exists := m.byUser[key.UserID]; exists {
			filtered := make([]*domain.APIKey, 0, len(userKeys))
			for _, k := range userKeys {
				if k.ID != id {
					filtered = append(filtered, k)
				}
			}
			m.byUser[key.UserID] = filtered
		}
	}
	return nil
}

// Test helpers
func setupTestService() (domain.APIKeyService, *mockAPIKeyRepository) {
	repo := newMockAPIKeyRepository()
	service := NewAPIKeyService(repo)
	return service, repo
}

func TestGenerateAPIKey(t *testing.T) {
	service, _ := setupTestService()
	userID := uuid.New()

	t.Run("successful generation with valid scopes", func(t *testing.T) {
		fullKey, apiKey, err := service.GenerateAPIKey(userID, "Test Key", []string{"read:todos"}, nil)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if fullKey == "" {
			t.Error("expected full key to not be empty")
		}

		if apiKey == nil {
			t.Fatal("expected apiKey to not be nil")
		}

		if apiKey.UserID != userID {
			t.Errorf("expected userID %s, got %s", userID, apiKey.UserID)
		}

		if apiKey.Name != "Test Key" {
			t.Errorf("expected name 'Test Key', got '%s'", apiKey.Name)
		}

		if len(apiKey.Scopes) != 1 || apiKey.Scopes[0] != "read:todos" {
			t.Errorf("expected scopes ['read:todos'], got %v", apiKey.Scopes)
		}

		if !apiKey.IsActive {
			t.Error("expected key to be active")
		}

		if apiKey.KeyHash == "" {
			t.Error("expected key hash to not be empty")
		}

		if apiKey.KeyPrefix == "" {
			t.Error("expected key prefix to not be empty")
		}

		// Verify key format
		if len(fullKey) < len(domain.KeyPrefix) {
			t.Error("key is too short")
		}
	})

	t.Run("successful generation with all scopes", func(t *testing.T) {
		fullKey, apiKey, err := service.GenerateAPIKey(userID, "Admin Key", []string{"read:todos", "write:todos", "admin"}, nil)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if fullKey == "" {
			t.Error("expected full key to not be empty")
		}

		if len(apiKey.Scopes) != 3 {
			t.Errorf("expected 3 scopes, got %d", len(apiKey.Scopes))
		}
	})

	t.Run("successful generation with expiration", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		_, apiKey, err := service.GenerateAPIKey(userID, "Expiring Key", []string{"read:todos"}, &expiresAt)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if apiKey.ExpiresAt == nil {
			t.Fatal("expected expires_at to be set")
		}

		if !apiKey.ExpiresAt.Equal(expiresAt) {
			t.Errorf("expected expires_at %v, got %v", expiresAt, *apiKey.ExpiresAt)
		}
	})

	t.Run("fails with invalid scope", func(t *testing.T) {
		_, _, err := service.GenerateAPIKey(userID, "Test Key", []string{"invalid:scope"}, nil)

		if !errors.Is(err, domain.ErrInvalidScope) {
			t.Errorf("expected ErrInvalidScope, got: %v", err)
		}
	})

	t.Run("fails with empty scopes", func(t *testing.T) {
		_, _, err := service.GenerateAPIKey(userID, "Test Key", []string{}, nil)

		if !errors.Is(err, domain.ErrInvalidScope) {
			t.Errorf("expected ErrInvalidScope, got: %v", err)
		}
	})

	t.Run("fails with empty name", func(t *testing.T) {
		_, _, err := service.GenerateAPIKey(userID, "", []string{"read:todos"}, nil)

		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("fails with whitespace-only name", func(t *testing.T) {
		_, _, err := service.GenerateAPIKey(userID, "   ", []string{"read:todos"}, nil)

		if err == nil {
			t.Error("expected error for whitespace-only name, got nil")
		}
	})

	t.Run("generates unique keys", func(t *testing.T) {
		fullKey1, _, _ := service.GenerateAPIKey(userID, "Key 1", []string{"read:todos"}, nil)
		fullKey2, _, _ := service.GenerateAPIKey(userID, "Key 2", []string{"read:todos"}, nil)

		if fullKey1 == fullKey2 {
			t.Error("expected different keys, got same key")
		}
	})

	t.Run("key has correct prefix format", func(t *testing.T) {
		fullKey, apiKey, _ := service.GenerateAPIKey(userID, "Test", []string{"read:todos"}, nil)

		// Check full key starts with prefix
		if len(fullKey) < len(domain.KeyPrefix) || fullKey[:len(domain.KeyPrefix)] != domain.KeyPrefix {
			t.Errorf("expected key to start with '%s', got '%s'", domain.KeyPrefix, fullKey)
		}

		// Check prefix extraction
		extractedPrefix, err := domain.ExtractPrefix(fullKey)
		if err != nil {
			t.Fatalf("failed to extract prefix: %v", err)
		}

		if extractedPrefix != apiKey.KeyPrefix {
			t.Errorf("extracted prefix %s doesn't match stored prefix %s", extractedPrefix, apiKey.KeyPrefix)
		}

		if len(extractedPrefix) != 8 {
			t.Errorf("expected prefix length 8, got %d", len(extractedPrefix))
		}
	})
}

func TestValidateAPIKey(t *testing.T) {
	service, _ := setupTestService()
	userID := uuid.New()

	t.Run("validates active key successfully", func(t *testing.T) {
		fullKey, apiKey, _ := service.GenerateAPIKey(userID, "Test Key", []string{"read:todos"}, nil)

		validated, err := service.ValidateAPIKey(fullKey)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if validated.ID != apiKey.ID {
			t.Errorf("expected ID %s, got %s", apiKey.ID, validated.ID)
		}
	})

	t.Run("updates last_used_at on validation", func(t *testing.T) {
		fullKey, _, _ := service.GenerateAPIKey(userID, "Test Key", []string{"read:todos"}, nil)

		// Initial validation should set last_used_at
		validated, err := service.ValidateAPIKey(fullKey)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if validated.LastUsedAt == nil {
			t.Error("expected LastUsedAt to be set after validation")
		}

		// Second validation should update last_used_at
		time.Sleep(10 * time.Millisecond)
		validated2, _ := service.ValidateAPIKey(fullKey)

		if validated2.LastUsedAt == nil {
			t.Error("expected LastUsedAt to be set after second validation")
		} else if validated.LastUsedAt != nil && !validated2.LastUsedAt.After(*validated.LastUsedAt) {
			t.Error("expected LastUsedAt to be updated on second validation")
		}
	})

	t.Run("fails with invalid key format", func(t *testing.T) {
		_, err := service.ValidateAPIKey("invalid-key-format")

		if !errors.Is(err, domain.ErrAPIKeyInvalid) {
			t.Errorf("expected ErrAPIKeyInvalid, got: %v", err)
		}
	})

	t.Run("fails with wrong prefix", func(t *testing.T) {
		_, err := service.ValidateAPIKey("wrongprefix_abc123")

		if !errors.Is(err, domain.ErrAPIKeyInvalid) {
			t.Errorf("expected ErrAPIKeyInvalid, got: %v", err)
		}
	})

	t.Run("fails with key that's too short", func(t *testing.T) {
		_, err := service.ValidateAPIKey("ouk_v1_123")

		if !errors.Is(err, domain.ErrAPIKeyInvalid) {
			t.Errorf("expected ErrAPIKeyInvalid, got: %v", err)
		}
	})

	t.Run("fails with non-existent key", func(t *testing.T) {
		// Generate a valid key format but don't store it
		validKey, _, _ := service.GenerateAPIKey(userID, "Test", []string{"read:todos"}, nil)

		// Modify the key slightly so it won't match
		modifiedKey := validKey[:len(validKey)-1] + "X"

		_, err := service.ValidateAPIKey(modifiedKey)

		if !errors.Is(err, domain.ErrAPIKeyInvalid) {
			t.Errorf("expected ErrAPIKeyInvalid for modified key, got: %v", err)
		}
	})

	t.Run("fails with expired key", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
		fullKey, _, _ := service.GenerateAPIKey(userID, "Expired Key", []string{"read:todos"}, &expiresAt)

		_, err := service.ValidateAPIKey(fullKey)

		if !errors.Is(err, domain.ErrAPIKeyExpired) {
			t.Errorf("expected ErrAPIKeyExpired, got: %v", err)
		}
	})

	t.Run("succeeds with key expiring in future", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		fullKey, _, _ := service.GenerateAPIKey(userID, "Future Expiring Key", []string{"read:todos"}, &expiresAt)

		validated, err := service.ValidateAPIKey(fullKey)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if validated == nil {
			t.Error("expected validated key, got nil")
		}
	})
}

func TestRevokeAPIKey(t *testing.T) {
	service, _ := setupTestService()
	userID := uuid.New()

	t.Run("successfully revokes active key", func(t *testing.T) {
		_, apiKey, _ := service.GenerateAPIKey(userID, "Test Key", []string{"read:todos"}, nil)

		err := service.RevokeAPIKey(apiKey.ID, "User requested")

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("fails to validate revoked key", func(t *testing.T) {
		fullKey, apiKey, _ := service.GenerateAPIKey(userID, "Test Key", []string{"read:todos"}, nil)

		// Revoke the key
		err := service.RevokeAPIKey(apiKey.ID, "Security breach")
		if err != nil {
			t.Fatalf("failed to revoke key: %v", err)
		}

		// Try to validate revoked key
		_, err = service.ValidateAPIKey(fullKey)

		if !errors.Is(err, domain.ErrAPIKeyRevoked) {
			t.Errorf("expected ErrAPIKeyRevoked, got: %v", err)
		}
	})

	t.Run("fails with non-existent key ID", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := service.RevokeAPIKey(nonExistentID, "Test")

		if !errors.Is(err, domain.ErrAPIKeyNotFound) {
			t.Errorf("expected ErrAPIKeyNotFound, got: %v", err)
		}
	})
}

func TestListAPIKeys(t *testing.T) {
	service, _ := setupTestService()
	user1 := uuid.New()
	user2 := uuid.New()

	t.Run("returns empty list for user with no keys", func(t *testing.T) {
		keys, err := service.ListAPIKeys(user1)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(keys) != 0 {
			t.Errorf("expected 0 keys, got %d", len(keys))
		}
	})

	t.Run("returns keys for user", func(t *testing.T) {
		service.GenerateAPIKey(user1, "Key 1", []string{"read:todos"}, nil)
		service.GenerateAPIKey(user1, "Key 2", []string{"write:todos"}, nil)

		keys, err := service.ListAPIKeys(user1)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(keys) != 2 {
			t.Errorf("expected 2 keys, got %d", len(keys))
		}
	})

	t.Run("does not return other user's keys", func(t *testing.T) {
		service.GenerateAPIKey(user1, "User1 Key", []string{"read:todos"}, nil)
		service.GenerateAPIKey(user2, "User2 Key", []string{"read:todos"}, nil)

		user1Keys, _ := service.ListAPIKeys(user1)
		user2Keys, _ := service.ListAPIKeys(user2)

		if len(user1Keys) != 1 {
			t.Errorf("expected 1 key for user1, got %d", len(user1Keys))
		}

		if len(user2Keys) != 1 {
			t.Errorf("expected 1 key for user2, got %d", len(user2Keys))
		}

		if user1Keys[0].Name != "User1 Key" {
			t.Errorf("expected 'User1 Key', got '%s'", user1Keys[0].Name)
		}

		if user2Keys[0].Name != "User2 Key" {
			t.Errorf("expected 'User2 Key', got '%s'", user2Keys[0].Name)
		}
	})

	t.Run("includes revoked keys in list", func(t *testing.T) {
		_, apiKey, _ := service.GenerateAPIKey(user1, "To Be Revoked", []string{"read:todos"}, nil)
		service.RevokeAPIKey(apiKey.ID, "Testing")

		keys, _ := service.ListAPIKeys(user1)

		found := false
		for _, key := range keys {
			if key.ID == apiKey.ID {
				found = true
				if key.IsActive {
					t.Error("expected revoked key to be inactive")
				}
				break
			}
		}

		if !found {
			t.Error("expected revoked key to be in list")
		}
	})
}

func TestHasScope(t *testing.T) {
	service, _ := setupTestService()
	userID := uuid.New()

	t.Run("returns true for existing scope", func(t *testing.T) {
		_, apiKey, _ := service.GenerateAPIKey(userID, "Test", []string{"read:todos", "write:todos"}, nil)

		if !service.HasScope(apiKey, "read:todos") {
			t.Error("expected HasScope to return true for read:todos")
		}

		if !service.HasScope(apiKey, "write:todos") {
			t.Error("expected HasScope to return true for write:todos")
		}
	})

	t.Run("returns false for missing scope", func(t *testing.T) {
		_, apiKey, _ := service.GenerateAPIKey(userID, "Test", []string{"read:todos"}, nil)

		if service.HasScope(apiKey, "write:todos") {
			t.Error("expected HasScope to return false for write:todos")
		}

		if service.HasScope(apiKey, "admin") {
			t.Error("expected HasScope to return false for admin")
		}
	})

	t.Run("returns false for nil apiKey", func(t *testing.T) {
		if service.HasScope(nil, "read:todos") {
			t.Error("expected HasScope to return false for nil apiKey")
		}
	})
}

func TestAPIKeyFormat(t *testing.T) {
	t.Run("GenerateAPIKeyRandom produces correct format", func(t *testing.T) {
		key, err := domain.GenerateAPIKeyRandom()

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if !strings.HasPrefix(key, domain.KeyPrefix) {
			t.Errorf("expected key to start with '%s', got '%s'", domain.KeyPrefix, key)
		}

		// Minimum length check: prefix + 8 chars for prefix extraction + base64 encoding
		if len(key) < len(domain.KeyPrefix)+10 {
			t.Errorf("key seems too short: %d chars", len(key))
		}
	})

	t.Run("ExtractPrefix validates correctly", func(t *testing.T) {
		tests := []struct {
			name    string
			key     string
			wantErr bool
		}{
			{
				name:    "valid key",
				key:     "ouk_v1_abc123def456",
				wantErr: false,
			},
			{
				name:    "missing prefix",
				key:     "invalid_abc123",
				wantErr: true,
			},
			{
				name:    "too short after prefix",
				key:     "ouk_v1_abc",
				wantErr: true,
			},
			{
				name:    "exactly 8 chars after prefix",
				key:     "ouk_v1_abcdefgh",
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				prefix, err := domain.ExtractPrefix(tt.key)

				if tt.wantErr {
					if err == nil {
						t.Error("expected error, got nil")
					}
				} else {
					if err != nil {
						t.Errorf("expected no error, got: %v", err)
					}
					if len(prefix) != 8 {
						t.Errorf("expected prefix length 8, got %d", len(prefix))
					}
				}
			})
		}
	})
}

func TestValidateScopes(t *testing.T) {
	t.Run("accepts valid scopes", func(t *testing.T) {
		err := domain.ValidateScopes([]string{"read:todos"})
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		err = domain.ValidateScopes([]string{"read:todos", "write:todos", "admin"})
		if err != nil {
			t.Errorf("expected no error for all scopes, got: %v", err)
		}
	})

	t.Run("rejects invalid scopes", func(t *testing.T) {
		err := domain.ValidateScopes([]string{"invalid:scope"})
		if err == nil {
			t.Error("expected error for invalid scope, got nil")
		}

		err = domain.ValidateScopes([]string{"read:todos", "invalid:scope"})
		if err == nil {
			t.Error("expected error when one scope is invalid, got nil")
		}
	})

	t.Run("rejects empty scopes", func(t *testing.T) {
		err := domain.ValidateScopes([]string{})
		if err == nil {
			t.Error("expected error for empty scopes, got nil")
		}
	})
}

func TestAPIKeyStateChecks(t *testing.T) {
	t.Run("IsExpired with no expiration", func(t *testing.T) {
		key := &domain.APIKey{ExpiresAt: nil}
		if key.IsExpired() {
			t.Error("expected key without expiration to not be expired")
		}
	})

	t.Run("IsExpired with past expiration", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		key := &domain.APIKey{ExpiresAt: &past}
		if !key.IsExpired() {
			t.Error("expected key with past expiration to be expired")
		}
	})

	t.Run("IsExpired with future expiration", func(t *testing.T) {
		future := time.Now().Add(1 * time.Hour)
		key := &domain.APIKey{ExpiresAt: &future}
		if key.IsExpired() {
			t.Error("expected key with future expiration to not be expired")
		}
	})

	t.Run("IsRevoked with revoked_at set", func(t *testing.T) {
		now := time.Now()
		key := &domain.APIKey{RevokedAt: &now, IsActive: true}
		if !key.IsRevoked() {
			t.Error("expected key with revoked_at to be revoked")
		}
	})

	t.Run("IsRevoked with is_active false", func(t *testing.T) {
		key := &domain.APIKey{RevokedAt: nil, IsActive: false}
		if !key.IsRevoked() {
			t.Error("expected key with is_active=false to be revoked")
		}
	})

	t.Run("IsRevoked with active key", func(t *testing.T) {
		key := &domain.APIKey{RevokedAt: nil, IsActive: true}
		if key.IsRevoked() {
			t.Error("expected active key to not be revoked")
		}
	})

	t.Run("HasScope finds existing scope", func(t *testing.T) {
		key := &domain.APIKey{Scopes: []string{"read:todos", "write:todos"}}
		if !key.HasScope("read:todos") {
			t.Error("expected HasScope to find read:todos")
		}
	})

	t.Run("HasScope returns false for missing scope", func(t *testing.T) {
		key := &domain.APIKey{Scopes: []string{"read:todos"}}
		if key.HasScope("admin") {
			t.Error("expected HasScope to return false for missing scope")
		}
	})
}

func TestHashVerification(t *testing.T) {
	t.Run("hash and verify same key", func(t *testing.T) {
		key, _ := domain.GenerateAPIKeyRandom()
		hash, err := hashAPIKey(key)

		if err != nil {
			t.Fatalf("failed to hash key: %v", err)
		}

		if !verifyAPIKey(key, hash) {
			t.Error("expected verification to succeed for same key")
		}
	})

	t.Run("verify fails for different key", func(t *testing.T) {
		key1, _ := domain.GenerateAPIKeyRandom()
		key2, _ := domain.GenerateAPIKeyRandom()
		hash, _ := hashAPIKey(key1)

		if verifyAPIKey(key2, hash) {
			t.Error("expected verification to fail for different key")
		}
	})

	t.Run("verify fails for invalid hash format", func(t *testing.T) {
		key, _ := domain.GenerateAPIKeyRandom()

		if verifyAPIKey(key, "invalid-hash") {
			t.Error("expected verification to fail for invalid hash format")
		}

		if verifyAPIKey(key, "$wrong$format") {
			t.Error("expected verification to fail for wrong format")
		}
	})

	t.Run("verify fails for wrong algorithm", func(t *testing.T) {
		key, _ := domain.GenerateAPIKeyRandom()
		wrongHash := "$argon2i$v=19$m=65536,t=1,p=4$salt$hash"

		if verifyAPIKey(key, wrongHash) {
			t.Error("expected verification to fail for wrong algorithm")
		}
	})

	t.Run("verify fails for invalid base64", func(t *testing.T) {
		key, _ := domain.GenerateAPIKeyRandom()
		invalidB64Hash := "$argon2id$v=19$m=65536,t=1,p=4$!!!invalid!!!$hash"

		if verifyAPIKey(key, invalidB64Hash) {
			t.Error("expected verification to fail for invalid base64")
		}
	})
}

func TestValidScopes(t *testing.T) {
	t.Run("IsValidScope returns true for valid scopes", func(t *testing.T) {
		validScopes := []string{"read:todos", "write:todos", "admin"}
		for _, scope := range validScopes {
			if !domain.IsValidScope(scope) {
				t.Errorf("expected IsValidScope to return true for %s", scope)
			}
		}
	})

	t.Run("IsValidScope returns false for invalid scope", func(t *testing.T) {
		invalidScopes := []string{"invalid", "read", "write", "delete:todos", ""}
		for _, scope := range invalidScopes {
			if domain.IsValidScope(scope) {
				t.Errorf("expected IsValidScope to return false for %s", scope)
			}
		}
	})
}
