package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/service"
)

type authResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func Auth(jwtService *service.JWTService, apiKeyService domain.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string
			var scopes []string
			var authType string
			
			authHeader := r.Header.Get("Authorization")
			apiKeyHeader := r.Header.Get("X-API-Key")
			
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := jwtService.ValidateAccessToken(r.Context(), token)
				if err != nil {
					writeAuthError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired JWT token")
					return
				}
				userID = claims.UserID
				scopes = []string{"read:todos", "write:todos", "admin"}
				authType = "jwt"
			} else if apiKeyHeader != "" {
				apiKey, err := apiKeyService.ValidateAPIKey(apiKeyHeader)
				if err != nil {
					switch err {
					case domain.ErrAPIKeyExpired:
						writeAuthError(w, http.StatusUnauthorized, "api_key_expired", "API key has expired")
					case domain.ErrAPIKeyRevoked:
						writeAuthError(w, http.StatusUnauthorized, "api_key_revoked", "API key has been revoked")
					default:
						writeAuthError(w, http.StatusUnauthorized, "invalid_api_key", "Invalid API key")
					}
					return
				}
				userID = apiKey.UserID.String()
				scopes = apiKey.Scopes
				authType = "api_key"
			} else {
				writeAuthError(w, http.StatusUnauthorized, "missing_auth", "Authentication required. Provide JWT Bearer token or X-API-Key header")
				return
			}
			
			ctx := r.Context()
			ctx = SetUserID(ctx, userID)
			ctx = SetUserScopes(ctx, scopes)
			ctx = SetAuthType(ctx, authType)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(jwtService *service.JWTService, apiKeyService domain.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string
			var scopes []string
			var authType string
			
			authHeader := r.Header.Get("Authorization")
			apiKeyHeader := r.Header.Get("X-API-Key")
			
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := jwtService.ValidateAccessToken(r.Context(), token)
				if err == nil {
					userID = claims.UserID
					scopes = []string{"read:todos", "write:todos", "admin"}
					authType = "jwt"
				}
			} else if apiKeyHeader != "" {
				apiKey, err := apiKeyService.ValidateAPIKey(apiKeyHeader)
				if err == nil {
					userID = apiKey.UserID.String()
					scopes = apiKey.Scopes
					authType = "api_key"
				}
			}
			
			if userID != "" {
				ctx := r.Context()
				ctx = SetUserID(ctx, userID)
				ctx = SetUserScopes(ctx, scopes)
				ctx = SetAuthType(ctx, authType)
				r = r.WithContext(ctx)
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scopes := GetUserScopes(r.Context())
			if scopes == nil {
				writeAuthError(w, http.StatusForbidden, "forbidden", "Authentication required")
				return
			}
			
			hasScope := false
			for _, s := range scopes {
				if s == scope {
					hasScope = true
					break
				}
			}
			
			if !hasScope {
				writeAuthError(w, http.StatusForbidden, "insufficient_scope", "Required scope: "+scope)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

func writeAuthError(w http.ResponseWriter, status int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(authResponse{
		Error:   errorCode,
		Message: message,
	})
}

func GetAuthenticatedUserID(ctx context.Context) string {
	return GetUserID(ctx)
}
