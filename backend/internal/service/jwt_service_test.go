package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupTestService(t *testing.T) (*JWTService, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	
	secret := "test-secret-key-for-jwt-signing"
	service := NewJWTService(secret, rdb)
	
	return service, mr
}

func TestNewJWTService(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	
	secret := "test-secret"
	service := NewJWTService(secret, rdb)
	
	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}
	
	if service.secret != secret {
		t.Errorf("Expected secret %s, got %s", secret, service.secret)
	}
	
	if service.redis != rdb {
		t.Error("Expected redis client to be set")
	}
}

func TestGenerateTokenPair(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	if tokenPair == nil {
		t.Fatal("Expected token pair, got nil")
	}
	
	if tokenPair.AccessToken == "" {
		t.Error("Expected access token to be set")
	}
	
	if tokenPair.RefreshToken == "" {
		t.Error("Expected refresh token to be set")
	}
	
	if tokenPair.AccessToken == tokenPair.RefreshToken {
		t.Error("Access and refresh tokens should be different")
	}
	
	expectedExpiry := time.Now().Add(accessTokenTTL)
	if tokenPair.ExpiresAt.Before(expectedExpiry.Add(-time.Second)) || tokenPair.ExpiresAt.After(expectedExpiry.Add(time.Second)) {
		t.Errorf("Expected expiry around %v, got %v", expectedExpiry, tokenPair.ExpiresAt)
	}
}

func TestGenerateTokenPair_InitializesTokenVersion(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	_, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	version, err := service.GetTokenVersion(ctx, userID)
	if err != nil {
		t.Fatalf("GetTokenVersion failed: %v", err)
	}
	
	if version != 1 {
		t.Errorf("Expected initial version 1, got %d", version)
	}
}

func TestGenerateTokenPair_UsesExistingVersion(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	mr.Set("token_version:user-123", "5")
	
	_, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	version, err := service.GetTokenVersion(ctx, userID)
	if err != nil {
		t.Fatalf("GetTokenVersion failed: %v", err)
	}
	
	if version != 5 {
		t.Errorf("Expected version 5, got %d", version)
	}
}

func TestValidateAccessToken_Success(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	claims, err := service.ValidateAccessToken(ctx, tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}
	
	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}
	
	if claims.TokenType != accessTokenType {
		t.Errorf("Expected token type %s, got %s", accessTokenType, claims.TokenType)
	}
	
	if claims.TokenVersion != 1 {
		t.Errorf("Expected token version 1, got %d", claims.TokenVersion)
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	
	_, err := service.ValidateAccessToken(ctx, "invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateAccessToken_WrongTokenType(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	_, err = service.ValidateAccessToken(ctx, tokenPair.RefreshToken)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for refresh token used as access, got %v", err)
	}
}

func TestValidateAccessToken_VersionMismatch(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	mr.Set("token_version:user-123", "2")
	
	_, err = service.ValidateAccessToken(ctx, tokenPair.AccessToken)
	if err != ErrTokenVersionMismatch {
		t.Errorf("Expected ErrTokenVersionMismatch, got %v", err)
	}
}

func TestValidateAccessToken_RevokedToken(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	err = service.RevokeToken(ctx, tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("RevokeToken failed: %v", err)
	}
	
	_, err = service.ValidateAccessToken(ctx, tokenPair.AccessToken)
	if err != ErrTokenRevoked {
		t.Errorf("Expected ErrTokenRevoked, got %v", err)
	}
}

func TestValidateAccessToken_ExpiredToken(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMiLCJ0b2tlbl92ZXJzaW9uIjoxLCJ0b2tlbl90eXBlIjoiYWNjZXNzIiwiZXhwIjoxNjAwMDAwMDAwLCJpYXQiOjE2MDAwMDAwMDB9.invalid"
	
	_, err := service.ValidateAccessToken(ctx, expiredToken)
	if err != ErrTokenExpired && !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected ErrTokenExpired, got %v", err)
	}
}

func TestValidateRefreshToken_Success(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	newTokenPair, err := service.ValidateRefreshToken(ctx, tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("ValidateRefreshToken failed: %v", err)
	}
	
	if newTokenPair == nil {
		t.Fatal("Expected new token pair, got nil")
	}
	
	if newTokenPair.AccessToken == tokenPair.AccessToken {
		t.Error("New access token should be different from old")
	}
	
	if newTokenPair.RefreshToken == tokenPair.RefreshToken {
		t.Error("New refresh token should be different from old")
	}
}

func TestValidateRefreshToken_InvalidToken(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	
	_, err := service.ValidateRefreshToken(ctx, "invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateRefreshToken_WrongTokenType(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	_, err = service.ValidateRefreshToken(ctx, tokenPair.AccessToken)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for access token used as refresh, got %v", err)
	}
}

func TestValidateRefreshToken_VersionMismatch(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	mr.Set("token_version:user-123", "2")
	
	_, err = service.ValidateRefreshToken(ctx, tokenPair.RefreshToken)
	if err != ErrTokenVersionMismatch {
		t.Errorf("Expected ErrTokenVersionMismatch, got %v", err)
	}
}

func TestValidateRefreshToken_RevokedToken(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	err = service.RevokeToken(ctx, tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("RevokeToken failed: %v", err)
	}
	
	_, err = service.ValidateRefreshToken(ctx, tokenPair.RefreshToken)
	if err != ErrTokenRevoked {
		t.Errorf("Expected ErrTokenRevoked, got %v", err)
	}
}

func TestRevokeToken_Success(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	err = service.RevokeToken(ctx, tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("RevokeToken failed: %v", err)
	}
	
	blacklistKey := tokenBlacklistPrefix + tokenPair.AccessToken
	exists := mr.Exists(blacklistKey)
	if !exists {
		t.Error("Expected token to be blacklisted")
	}
	
	ttl := mr.TTL(blacklistKey)
	if ttl <= 0 || ttl > accessTokenTTL {
		t.Errorf("Expected TTL between 0 and %v, got %v", accessTokenTTL, ttl)
	}
}

func TestRevokeToken_InvalidToken(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	
	err := service.RevokeToken(ctx, "invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestRevokeToken_ExpiredToken(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMiLCJ0b2tlbl92ZXJzaW9uIjoxLCJ0b2tlbl90eXBlIjoiYWNjZXNzIiwiZXhwIjoxNjAwMDAwMDAwLCJpYXQiOjE2MDAwMDAwMDB9.invalid"
	
	err := service.RevokeToken(ctx, expiredToken)
	if err != nil {
		t.Errorf("Expected no error for expired token, got %v", err)
	}
}

func TestIncrementTokenVersion_Success(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	err := service.IncrementTokenVersion(ctx, userID)
	if err != nil {
		t.Fatalf("IncrementTokenVersion failed: %v", err)
	}
	
	version, err := service.GetTokenVersion(ctx, userID)
	if err != nil {
		t.Fatalf("GetTokenVersion failed: %v", err)
	}
	
	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}
	
	err = service.IncrementTokenVersion(ctx, userID)
	if err != nil {
		t.Fatalf("IncrementTokenVersion failed: %v", err)
	}
	
	version, err = service.GetTokenVersion(ctx, userID)
	if err != nil {
		t.Fatalf("GetTokenVersion failed: %v", err)
	}
	
	if version != 2 {
		t.Errorf("Expected version 2, got %d", version)
	}
}

func TestIncrementTokenVersion_MultipleIncrements(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	for i := 0; i < 5; i++ {
		err := service.IncrementTokenVersion(ctx, userID)
		if err != nil {
			t.Fatalf("IncrementTokenVersion failed at iteration %d: %v", i, err)
		}
	}
	
	version, err := service.GetTokenVersion(ctx, userID)
	if err != nil {
		t.Fatalf("GetTokenVersion failed: %v", err)
	}
	
	if version != 5 {
		t.Errorf("Expected version 5, got %d", version)
	}
}

func TestGetTokenVersion_InitializesOnFirstCall(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "new-user"
	
	version, err := service.GetTokenVersion(ctx, userID)
	if err != nil {
		t.Fatalf("GetTokenVersion failed: %v", err)
	}
	
	if version != 1 {
		t.Errorf("Expected initial version 1, got %d", version)
	}
}

func TestGetTokenVersion_ReturnsExistingVersion(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	mr.Set("token_version:user-123", "10")
	
	version, err := service.GetTokenVersion(ctx, userID)
	if err != nil {
		t.Fatalf("GetTokenVersion failed: %v", err)
	}
	
	if version != 10 {
		t.Errorf("Expected version 10, got %d", version)
	}
}

func TestParseToken_WrongSigningMethod(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	tokenWithRS256 := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMiLCJ0b2tlbl92ZXJzaW9uIjoxLCJ0b2tlbl90eXBlIjoiYWNjZXNzIn0.signature"
	
	_, err := service.parseToken(tokenWithRS256)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for wrong signing method, got %v", err)
	}
}

func TestParseToken_MalformedToken(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	_, err := service.parseToken("not.a.valid.jwt")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for malformed token, got %v", err)
	}
}

func TestTokenPairExpirationTimes(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	beforeGen := time.Now()
	tokenPair, err := service.GenerateTokenPair(ctx, userID)
	afterGen := time.Now()
	
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	
	expectedExpiry := beforeGen.Add(accessTokenTTL)
	if tokenPair.ExpiresAt.Before(expectedExpiry.Add(-time.Second)) || tokenPair.ExpiresAt.After(afterGen.Add(accessTokenTTL).Add(time.Second)) {
		t.Errorf("Expected expiry within 1 second of %v, got %v", expectedExpiry, tokenPair.ExpiresAt)
	}
}

func TestTokenVersionKeyHasTTL(t *testing.T) {
	service, mr := setupTestService(t)
	defer mr.Close()
	
	ctx := context.Background()
	userID := "user-123"
	
	err := service.IncrementTokenVersion(ctx, userID)
	if err != nil {
		t.Fatalf("IncrementTokenVersion failed: %v", err)
	}
	
	ttl := mr.TTL("token_version:" + userID)
	expectedMinTTL := refreshTokenTTL + 6*24*time.Hour
	expectedMaxTTL := refreshTokenTTL + 8*24*time.Hour
	
	if ttl < expectedMinTTL || ttl > expectedMaxTTL {
		t.Errorf("Expected TTL between %v and %v, got %v", expectedMinTTL, expectedMaxTTL, ttl)
	}
}
