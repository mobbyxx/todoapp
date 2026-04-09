package middleware

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/user/todo-api/internal/observability"
)

var sensitiveHeaders = []string{
	"authorization",
	"x-api-key",
	"cookie",
	"x-csrf-token",
	"x-xsrf-token",
}

var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`"password"\s*:\s*"[^"]*"`),
	regexp.MustCompile(`"token"\s*:\s*"[^"]*"`),
	regexp.MustCompile(`"refresh_token"\s*:\s*"[^"]*"`),
	regexp.MustCompile(`"api_key"\s*:\s*"[^"]*"`),
	regexp.MustCompile(`"secret"\s*:\s*"[^"]*"`),
	regexp.MustCompile(`"credit_card"\s*:\s*"[^"]*"`),
	regexp.MustCompile(`"cvv"\s*:\s*"[^"]*"`),
	regexp.MustCompile(`"ssn"\s*:\s*"[^"]*"`),
}

func SecureLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := NewResponseWriter(w)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		requestID := GetRequestID(r.Context())
		userID := GetUserID(r.Context())

		observability.RecordHTTPRequest(r.Method, r.URL.Path, rw.StatusCode(), duration.Seconds())

		sanitizedHeaders := sanitizeHeaders(r.Header)
		sanitizedQuery := sanitizeQueryParams(r.URL.RawQuery)

		event := log.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("query", sanitizedQuery).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Int("status", rw.StatusCode()).
			Dur("duration", duration).
			Interface("headers", sanitizedHeaders)

		if userID != "" {
			event = event.Str("user_id", userID)
		}

		authType := GetAuthType(r.Context())
		if authType != "" {
			event = event.Str("auth_type", authType)
		}

		if rw.StatusCode() >= 500 {
			event.Msg("server error")
		} else if rw.StatusCode() >= 400 {
			event.Msg("client error")
		} else {
			event.Msg("request completed")
		}
	})
}

func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)
	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if isSensitiveHeader(lowerKey) {
			sanitized[key] = "[REDACTED]"
		} else {
			sanitized[key] = strings.Join(values, ", ")
		}
	}
	return sanitized
}

func isSensitiveHeader(key string) bool {
	for _, sensitive := range sensitiveHeaders {
		if key == sensitive {
			return true
		}
	}
	return false
}

func sanitizeQueryParams(query string) string {
	if query == "" {
		return ""
	}

	parts := strings.Split(query, "&")
	for i, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			key := strings.ToLower(kv[0])
			if strings.Contains(key, "token") ||
				strings.Contains(key, "key") ||
				strings.Contains(key, "secret") ||
				strings.Contains(key, "password") {
				parts[i] = kv[0] + "=[REDACTED]"
			}
		}
	}
	return strings.Join(parts, "&")
}

func SanitizeBody(body []byte) []byte {
	bodyStr := string(body)
	for _, pattern := range sensitivePatterns {
		bodyStr = pattern.ReplaceAllString(bodyStr, `"$1":"[REDACTED]"`)
	}
	return []byte(bodyStr)
}

func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func MaskJWT(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "[INVALID_JWT]"
	}
	return parts[0] + ".****." + parts[2][:4] + "****"
}

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return "***@" + domain
	}

	maskedLocal := local[:1] + strings.Repeat("*", len(local)-2) + local[len(local)-1:]
	return maskedLocal + "@" + domain
}
