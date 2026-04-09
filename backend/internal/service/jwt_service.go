package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
	
	accessTokenType  = "access"
	refreshTokenType = "refresh"
	
	tokenVersionKeyPrefix = "token_version:"
	tokenBlacklistPrefix  = "blacklist:"
)

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrTokenRevoked         = errors.New("token revoked")
	ErrTokenVersionMismatch = errors.New("token version mismatch")
)

type CustomClaims struct {
	UserID       string `json:"user_id"`
	TokenVersion int    `json:"token_version"`
	TokenType    string `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type JWTService struct {
	secret         string
	secretPrevious string
	redis          *redis.Client
}

func NewJWTService(secret string, redisClient *redis.Client) *JWTService {
	return &JWTService{
		secret: secret,
		redis:  redisClient,
	}
}

func NewJWTServiceWithRotation(secret, secretPrevious string, redisClient *redis.Client) *JWTService {
	return &JWTService{
		secret:         secret,
		secretPrevious: secretPrevious,
		redis:          redisClient,
	}
}

func (s *JWTService) GenerateTokenPair(ctx context.Context, userID string) (*TokenPair, error) {
	tokenVersion, err := s.getTokenVersion(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token version: %w", err)
	}

	now := time.Now()

	accessClaims := CustomClaims{
		UserID:       userID,
		TokenVersion: tokenVersion,
		TokenType:    accessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshClaims := CustomClaims{
		UserID:       userID,
		TokenVersion: tokenVersion,
		TokenType:    refreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(accessTokenTTL),
	}, nil
}

func (s *JWTService) ValidateAccessToken(ctx context.Context, tokenString string) (*CustomClaims, error) {
	claims, err := s.validateToken(ctx, tokenString, accessTokenType)
	if err != nil {
		return nil, err
	}

	currentVersion, err := s.getTokenVersion(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token version: %w", err)
	}

	if claims.TokenVersion != currentVersion {
		return nil, ErrTokenVersionMismatch
	}

	return claims, nil
}

func (s *JWTService) ValidateRefreshToken(ctx context.Context, tokenString string) (*TokenPair, error) {
	claims, err := s.validateToken(ctx, tokenString, refreshTokenType)
	if err != nil {
		return nil, err
	}

	currentVersion, err := s.getTokenVersion(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token version: %w", err)
	}

	if claims.TokenVersion != currentVersion {
		return nil, ErrTokenVersionMismatch
	}

	return s.GenerateTokenPair(ctx, claims.UserID)
}

func (s *JWTService) RevokeToken(ctx context.Context, tokenString string) error {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return err
	}

	now := time.Now()
	var ttl time.Duration
	
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.After(now) {
		ttl = claims.ExpiresAt.Time.Sub(now)
	} else {
		return nil
	}

	blacklistKey := fmt.Sprintf("%s%s", tokenBlacklistPrefix, tokenString)
	if err := s.redis.Set(ctx, blacklistKey, "1", ttl).Err(); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

func (s *JWTService) IncrementTokenVersion(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", tokenVersionKeyPrefix, userID)
	
	_, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment token version: %w", err)
	}

	versionTTL := refreshTokenTTL + 7*24*time.Hour
	if err := s.redis.Expire(ctx, key, versionTTL).Err(); err != nil {
		s.redis.Decr(ctx, key)
		return fmt.Errorf("failed to set token version TTL: %w", err)
	}
	
	return nil
}

func (s *JWTService) GetTokenVersion(ctx context.Context, userID string) (int, error) {
	return s.getTokenVersion(ctx, userID)
}

func (s *JWTService) getTokenVersion(ctx context.Context, userID string) (int, error) {
	key := fmt.Sprintf("%s%s", tokenVersionKeyPrefix, userID)
	
	version, err := s.redis.Get(ctx, key).Int()
	if err == redis.Nil {
		version = 1
		if err := s.redis.Set(ctx, key, version, refreshTokenTTL+7*24*time.Hour).Err(); err != nil {
			return 0, fmt.Errorf("failed to initialize token version: %w", err)
		}
		return version, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get token version: %w", err)
	}
	
	return version, nil
}

func (s *JWTService) validateToken(ctx context.Context, tokenString string, expectedType string) (*CustomClaims, error) {
	blacklistKey := fmt.Sprintf("%s%s", tokenBlacklistPrefix, tokenString)
	exists, err := s.redis.Exists(ctx, blacklistKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check blacklist: %w", err)
	}
	if exists > 0 {
		return nil, ErrTokenRevoked
	}

	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != expectedType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *JWTService) parseToken(tokenString string) (*CustomClaims, error) {
	claims, err := s.parseWithSecret(tokenString, s.secret)
	if err == nil {
		return claims, nil
	}

	if s.secretPrevious != "" && (errors.Is(err, ErrInvalidToken) || errors.Is(err, jwt.ErrSignatureInvalid)) {
		claims, err = s.parseWithSecret(tokenString, s.secretPrevious)
		if err == nil {
			return claims, nil
		}
	}

	return nil, err
}

func (s *JWTService) parseWithSecret(tokenString, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
