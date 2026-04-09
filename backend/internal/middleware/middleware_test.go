package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "request ID exists",
			ctx:      context.WithValue(context.Background(), RequestIDKey, "test-request-id"),
			expected: "test-request-id",
		},
		{
			name:     "request ID does not exist",
			ctx:      context.Background(),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRequestID(tt.ctx)
			if result != tt.expected {
				t.Errorf("GetRequestID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "user ID exists",
			ctx:      context.WithValue(context.Background(), UserIDKey, "user-123"),
			expected: "user-123",
		},
		{
			name:     "user ID does not exist",
			ctx:      context.Background(),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserID(tt.ctx)
			if result != tt.expected {
				t.Errorf("GetUserID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetUserScopes(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected []string
	}{
		{
			name:     "scopes exist",
			ctx:      context.WithValue(context.Background(), UserScopesKey, []string{"read:todos", "write:todos"}),
			expected: []string{"read:todos", "write:todos"},
		},
		{
			name:     "scopes do not exist",
			ctx:      context.Background(),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserScopes(tt.ctx)
			if len(result) != len(tt.expected) {
				t.Errorf("GetUserScopes() length = %v, want %v", len(result), len(tt.expected))
			}
		})
	}
}

func TestIsAuthenticated(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected bool
	}{
		{
			name:     "authenticated",
			ctx:      context.WithValue(context.Background(), UserIDKey, "user-123"),
			expected: true,
		},
		{
			name:     "not authenticated",
			ctx:      context.Background(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthenticated(tt.ctx)
			if result != tt.expected {
				t.Errorf("IsAuthenticated() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasScope(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		scope    string
		expected bool
	}{
		{
			name:     "has scope",
			ctx:      context.WithValue(context.Background(), UserScopesKey, []string{"read:todos", "write:todos"}),
			scope:    "read:todos",
			expected: true,
		},
		{
			name:     "does not have scope",
			ctx:      context.WithValue(context.Background(), UserScopesKey, []string{"read:todos"}),
			scope:    "admin",
			expected: false,
		},
		{
			name:     "no scopes",
			ctx:      context.Background(),
			scope:    "read:todos",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasScope(tt.ctx, tt.scope)
			if result != tt.expected {
				t.Errorf("HasScope() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateRequestID(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()
	
	if id1 == "" {
		t.Error("GenerateRequestID() returned empty string")
	}
	
	if id1 == id2 {
		t.Error("GenerateRequestID() returned duplicate IDs")
	}
	
	if len(id1) != 36 {
		t.Errorf("GenerateRequestID() returned ID of length %d, want 36", len(id1))
	}
}

func TestResponseWriter(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		rr := httptest.NewRecorder()
		rw := NewResponseWriter(rr)
		
		rw.WriteHeader(http.StatusCreated)
		
		if rw.StatusCode() != http.StatusCreated {
			t.Errorf("StatusCode() = %d, want %d", rw.StatusCode(), http.StatusCreated)
		}
	})
	
	t.Run("default status code", func(t *testing.T) {
		rr := httptest.NewRecorder()
		rw := NewResponseWriter(rr)
		
		if rw.StatusCode() != http.StatusOK {
			t.Errorf("StatusCode() = %d, want %d", rw.StatusCode(), http.StatusOK)
		}
	})
	
	t.Run("write triggers default status", func(t *testing.T) {
		rr := httptest.NewRecorder()
		rw := NewResponseWriter(rr)
		
		rw.Write([]byte("test"))
		
		if rw.StatusCode() != http.StatusOK {
			t.Errorf("StatusCode() = %d, want %d", rw.StatusCode(), http.StatusOK)
		}
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			t.Error("RequestID not found in context")
		}
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := RequestID(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	
	middleware.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	
	header := rr.Header().Get("X-Request-ID")
	if header == "" {
		t.Error("X-Request-ID header not set")
	}
	
	if len(header) != 36 {
		t.Errorf("X-Request-ID header length = %d, want 36", len(header))
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Run("recovers from panic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})
		
		middleware := Recovery(handler)
		
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		
		middleware.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusInternalServerError)
		}
		
		body := rr.Body.String()
		if !strings.Contains(body, "internal_server_error") {
			t.Errorf("response body does not contain error code: got %v", body)
		}
	})
	
	t.Run("normal request passes through", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})
		
		middleware := Recovery(handler)
		
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		
		middleware.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
		}
		
		if rr.Body.String() != "success" {
			t.Errorf("handler returned wrong body: got %v want %v", rr.Body.String(), "success")
		}
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := SecurityHeaders(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	
	middleware.ServeHTTP(rr, req)
	
	tests := []struct {
		header   string
		expected string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}
	
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			value := rr.Header().Get(tt.header)
			if value != tt.expected {
				t.Errorf("%s header = %v, want %v", tt.header, value, tt.expected)
			}
		})
	}
	
	csp := rr.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("Content-Security-Policy header not set")
	}
	
	hsts := rr.Header().Get("Strict-Transport-Security")
	if hsts == "" {
		t.Error("Strict-Transport-Security header not set")
	}
}

func TestCORSMiddleware(t *testing.T) {
	t.Run("allows preflight request", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler should not be called for OPTIONS request")
		})
		
		config := &CORSConfig{
			AllowedOrigins:   []string{"http://localhost:3000"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: true,
			MaxAge:           86400,
		}
		
		middleware := CORS(config)(handler)
		
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rr := httptest.NewRecorder()
		
		middleware.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusNoContent {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusNoContent)
		}
		
		if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Error("Access-Control-Allow-Origin header not set correctly")
		}
	})
	
	t.Run("blocks unauthorized origin", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		
		config := &CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
			AllowedMethods: []string{"GET"},
		}
		
		middleware := CORS(config)(handler)
		
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://evil.com")
		rr := httptest.NewRecorder()
		
		middleware.ServeHTTP(rr, req)
		
		if rr.Header().Get("Access-Control-Allow-Origin") == "http://evil.com" {
			t.Error("Access-Control-Allow-Origin should not be set for unauthorized origin")
		}
	})
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		allowed  []string
		expected bool
	}{
		{
			name:     "origin is allowed",
			origin:   "http://localhost:3000",
			allowed:  []string{"http://localhost:3000", "http://localhost:8080"},
			expected: true,
		},
		{
			name:     "origin is not allowed",
			origin:   "http://evil.com",
			allowed:  []string{"http://localhost:3000"},
			expected: false,
		},
		{
			name:     "empty origin",
			origin:   "",
			allowed:  []string{"http://localhost:3000"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowed)
			if result != tt.expected {
				t.Errorf("isOriginAllowed() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetAllowedOrigins(t *testing.T) {
	t.Run("returns default origins when env not set", func(t *testing.T) {
		origins := getAllowedOrigins()
		if len(origins) == 0 {
			t.Error("getAllowedOrigins() returned empty slice")
		}
	})
}

type MockAPIKeyService struct {
	validateFunc func(key string) (*domain.APIKey, error)
}

func (m *MockAPIKeyService) GenerateAPIKey(userID uuid.UUID, name string, scopes []string, expiresAt *time.Time) (string, *domain.APIKey, error) {
	return "", nil, nil
}

func (m *MockAPIKeyService) ValidateAPIKey(key string) (*domain.APIKey, error) {
	if m.validateFunc != nil {
		return m.validateFunc(key)
	}
	return nil, domain.ErrAPIKeyInvalid
}

func (m *MockAPIKeyService) RevokeAPIKey(keyID uuid.UUID, reason string) error {
	return nil
}

func (m *MockAPIKeyService) ListAPIKeys(userID uuid.UUID) ([]*domain.APIKey, error) {
	return nil, nil
}

func (m *MockAPIKeyService) HasScope(apiKey *domain.APIKey, scope string) bool {
	return false
}
