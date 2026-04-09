package domain

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// APIKeyScope defines the available permission scopes for API keys
type APIKeyScope string

const (
	ScopeReadTodos  APIKeyScope = "read:todos"
	ScopeWriteTodos APIKeyScope = "write:todos"
	ScopeAdmin      APIKeyScope = "admin"
)

// ValidScopes returns all valid API key scopes
var ValidScopes = []APIKeyScope{
	ScopeReadTodos,
	ScopeWriteTodos,
	ScopeAdmin,
}

// IsValidScope checks if a scope string is valid
func IsValidScope(scope string) bool {
	for _, s := range ValidScopes {
		if string(s) == scope {
			return true
		}
	}
	return false
}

// APIKey represents an API key entity in the system
type APIKey struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	KeyHash     string      `json:"-"` // Never expose hash
	KeyPrefix   string      `json:"key_prefix"` // First 8 chars of full key for identification
	Name        string      `json:"name"`
	Scopes      []string    `json:"scopes"`
	RateLimit   int         `json:"rate_limit"`
	LastUsedAt  *time.Time  `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time  `json:"expires_at,omitempty"`
	IsActive    bool        `json:"is_active"`
	RevokedAt   *time.Time  `json:"revoked_at,omitempty"`
	RevokedReason string   `json:"revoked_reason,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// APIKeyErrors
type APIKeyError string

func (e APIKeyError) Error() string {
	return string(e)
}

const (
	ErrAPIKeyNotFound    APIKeyError = "api key not found"
	ErrAPIKeyExpired     APIKeyError = "api key has expired"
	ErrAPIKeyRevoked     APIKeyError = "api key has been revoked"
	ErrAPIKeyInvalid     APIKeyError = "api key is invalid"
	ErrInvalidScope      APIKeyError = "invalid scope provided"
	ErrKeyGeneration     APIKeyError = "failed to generate api key"
)

// APIKeyRepository defines the interface for API key persistence
type APIKeyRepository interface {
	Create(apiKey *APIKey) error
	GetByPrefix(prefix string) (*APIKey, error)
	GetByID(id uuid.UUID) (*APIKey, error)
	GetByUserID(userID uuid.UUID) ([]*APIKey, error)
	UpdateLastUsed(id uuid.UUID) error
	Revoke(id uuid.UUID, reason string) error
	Delete(id uuid.UUID) error
}

// APIKeyService defines the interface for API key business logic
type APIKeyService interface {
	GenerateAPIKey(userID uuid.UUID, name string, scopes []string, expiresAt *time.Time) (string, *APIKey, error)
	ValidateAPIKey(key string) (*APIKey, error)
	RevokeAPIKey(keyID uuid.UUID, reason string) error
	ListAPIKeys(userID uuid.UUID) ([]*APIKey, error)
	HasScope(apiKey *APIKey, scope string) bool
}

// KeyPrefix is the prefix for all API keys
const KeyPrefix = "ouk_v1_"

// KeyPrefixLength is the length of the prefix (ouk_v1_ = 7 chars)
const KeyPrefixLength = 7

// GenerateAPIKeyRandom generates a random API key string
// Format: ouk_v1_<base64url(random32bytes)>
// Returns the full key (which should be shown ONCE to the user)
func GenerateAPIKeyRandom() (string, error) {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Encode as base64url (URL-safe base64 without padding)
	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)
	
	// Construct full key
	fullKey := KeyPrefix + encoded
	
	return fullKey, nil
}

// ExtractPrefix extracts the prefix (first 8 chars after prefix) from a full API key
// This is used for database lookups
func ExtractPrefix(fullKey string) (string, error) {
	if !strings.HasPrefix(fullKey, KeyPrefix) {
		return "", errors.New("invalid api key format: missing prefix")
	}
	
	// Remove the prefix
	keyWithoutPrefix := strings.TrimPrefix(fullKey, KeyPrefix)
	
	if len(keyWithoutPrefix) < 8 {
		return "", errors.New("invalid api key format: key too short")
	}
	
	// Return first 8 characters
	return keyWithoutPrefix[:8], nil
}

// ValidateScopes validates a list of scope strings
func ValidateScopes(scopes []string) error {
	if len(scopes) == 0 {
		return errors.New("at least one scope is required")
	}
	
	for _, scope := range scopes {
		if !IsValidScope(scope) {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}
	
	return nil
}

// IsExpired checks if the API key has expired
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsRevoked checks if the API key has been revoked
func (k *APIKey) IsRevoked() bool {
	return k.RevokedAt != nil || !k.IsActive
}

// HasScope checks if the API key has a specific scope
func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}
