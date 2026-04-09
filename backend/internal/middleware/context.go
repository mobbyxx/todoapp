package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Context keys for storing values in request context
type contextKey string

const (
	// RequestIDKey is the key for storing request ID in context
	RequestIDKey contextKey = "request_id"
	// UserIDKey is the key for storing user ID in context
	UserIDKey contextKey = "user_id"
	// UserScopesKey is the key for storing user scopes in context
	UserScopesKey contextKey = "user_scopes"
	// AuthTypeKey is the key for storing authentication type in context
	AuthTypeKey contextKey = "auth_type"
)

// GetRequestID retrieves the request ID from the context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserScopes retrieves the user scopes from the context
func GetUserScopes(ctx context.Context) []string {
	if scopes, ok := ctx.Value(UserScopesKey).([]string); ok {
		return scopes
	}
	return nil
}

// GetAuthType retrieves the authentication type from the context
func GetAuthType(ctx context.Context) string {
	if authType, ok := ctx.Value(AuthTypeKey).(string); ok {
		return authType
	}
	return ""
}

// SetRequestID sets the request ID in the context
func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey, id)
}

// SetUserID sets the user ID in the context
func SetUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

// SetUserScopes sets the user scopes in the context
func SetUserScopes(ctx context.Context, scopes []string) context.Context {
	return context.WithValue(ctx, UserScopesKey, scopes)
}

// SetAuthType sets the authentication type in the context
func SetAuthType(ctx context.Context, authType string) context.Context {
	return context.WithValue(ctx, AuthTypeKey, authType)
}

// GenerateRequestID generates a new UUID-based request ID
func GenerateRequestID() string {
	return uuid.New().String()
}

// IsAuthenticated checks if the request has a valid authentication
func IsAuthenticated(ctx context.Context) bool {
	return GetUserID(ctx) != ""
}

// HasScope checks if the authenticated user has a specific scope
func HasScope(ctx context.Context, scope string) bool {
	scopes := GetUserScopes(ctx)
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// ResponseWriter is a custom response writer that captures status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

// WriteHeader captures the status code and writes it
func (rw *ResponseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write writes the response body
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// StatusCode returns the status code
func (rw *ResponseWriter) StatusCode() int {
	return rw.statusCode
}
