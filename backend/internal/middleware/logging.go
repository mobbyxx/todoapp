package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/user/todo-api/internal/observability"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := NewResponseWriter(w)
		
		next.ServeHTTP(rw, r)
		
		duration := time.Since(start)
		requestID := GetRequestID(r.Context())
		userID := GetUserID(r.Context())

		observability.RecordHTTPRequest(r.Method, r.URL.Path, rw.StatusCode(), duration.Seconds())

		event := log.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Int("status", rw.StatusCode()).
			Dur("duration", duration)
		
		if userID != "" {
			event = event.Str("user_id", userID)
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
