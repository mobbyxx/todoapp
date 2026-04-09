package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

// apiKeyService implements the domain.APIKeyService interface
type apiKeyService struct {
	repo domain.APIKeyRepository
}

// NewAPIKeyService creates a new API key service instance
func NewAPIKeyService(repo domain.APIKeyRepository) domain.APIKeyService {
	return &apiKeyService{
		repo: repo,
	}
}

// GenerateAPIKey creates a new API key for a user
// Returns the full key (which must be shown to the user immediately) and the stored APIKey record
func (s *apiKeyService) GenerateAPIKey(userID uuid.UUID, name string, scopes []string, expiresAt *time.Time) (string, *domain.APIKey, error) {
	// Validate scopes
	if err := domain.ValidateScopes(scopes); err != nil {
		return "", nil, fmt.Errorf("%w: %v", domain.ErrInvalidScope, err)
	}

	// Validate name
	if strings.TrimSpace(name) == "" {
		return "", nil, errors.New("name is required")
	}

	// Generate random API key
	fullKey, err := domain.GenerateAPIKeyRandom()
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", domain.ErrKeyGeneration, err)
	}

	// Extract prefix for database lookup
	prefix, err := domain.ExtractPrefix(fullKey)
	if err != nil {
		return "", nil, err
	}

	// Hash the key using Argon2id
	keyHash, err := hashAPIKey(fullKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash api key: %w", err)
	}

	// Create API key record
	apiKey := &domain.APIKey{
		UserID:    userID,
		KeyHash:   keyHash,
		KeyPrefix: prefix,
		Name:      name,
		Scopes:    scopes,
		RateLimit: 100, // Default rate limit
		IsActive:  true,
		ExpiresAt: expiresAt,
	}

	// Store in database
	if err := s.repo.Create(apiKey); err != nil {
		return "", nil, fmt.Errorf("failed to store api key: %w", err)
	}

	// Return the full key (this is the ONLY time the user will see it)
	return fullKey, apiKey, nil
}

// ValidateAPIKey validates an API key and returns the associated record
// Updates last_used_at on successful validation
func (s *apiKeyService) ValidateAPIKey(key string) (*domain.APIKey, error) {
	// Extract prefix for database lookup
	prefix, err := domain.ExtractPrefix(key)
	if err != nil {
		return nil, domain.ErrAPIKeyInvalid
	}

	// Retrieve API key record by prefix
	apiKey, err := s.repo.GetByPrefix(prefix)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrAPIKeyInvalid
		}
		return nil, fmt.Errorf("failed to retrieve api key: %w", err)
	}

	// Check if key is active
	if !apiKey.IsActive {
		return nil, domain.ErrAPIKeyRevoked
	}

	// Check if key is revoked
	if apiKey.IsRevoked() {
		return nil, domain.ErrAPIKeyRevoked
	}

	// Check if key has expired
	if apiKey.IsExpired() {
		return nil, domain.ErrAPIKeyExpired
	}

	// Verify key hash using constant-time comparison
	if !verifyAPIKey(key, apiKey.KeyHash) {
		return nil, domain.ErrAPIKeyInvalid
	}

	// Update last used timestamp
	if err := s.repo.UpdateLastUsed(apiKey.ID); err != nil {
		// Log error but don't fail validation
		// In production, this should be logged
		_ = err
	}
	now := time.Now()
	apiKey.LastUsedAt = &now

	return apiKey, nil
}

// RevokeAPIKey deactivates an API key
func (s *apiKeyService) RevokeAPIKey(keyID uuid.UUID, reason string) error {
	if err := s.repo.Revoke(keyID, reason); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrAPIKeyNotFound
		}
		return fmt.Errorf("failed to revoke api key: %w", err)
	}
	return nil
}

// ListAPIKeys returns all API keys for a user
func (s *apiKeyService) ListAPIKeys(userID uuid.UUID) ([]*domain.APIKey, error) {
	apiKeys, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}
	return apiKeys, nil
}

// HasScope checks if an API key has a specific scope
func (s *apiKeyService) HasScope(apiKey *domain.APIKey, scope string) bool {
	if apiKey == nil {
		return false
	}
	return apiKey.HasScope(scope)
}

// hashAPIKey hashes an API key using Argon2id
// Format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
func hashAPIKey(key string) (string, error) {
	// Generate random salt
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the key using Argon2id
	hash := argon2.IDKey([]byte(key), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Encode as base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads, b64Salt, b64Hash)

	return encodedHash, nil
}

// verifyAPIKey verifies an API key against a stored hash
func verifyAPIKey(key, encodedHash string) bool {
	// Parse the encoded hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	// Check algorithm
	if parts[1] != "argon2id" {
		return false
	}

	// Parse parameters
	var version, memory, time, threads int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false
	}

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	// Compute hash with same parameters
	computedHash := argon2.IDKey([]byte(key), salt, uint32(time), uint32(memory), uint8(threads), uint32(len(storedHash)))

	// Constant-time comparison
	return subtle.ConstantTimeCompare(computedHash, storedHash) == 1
}
