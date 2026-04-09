package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())
				
				log.Error().
					Str("request_id", requestID).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Interface("panic", err).
					Str("stack", string(debug.Stack())).
					Msg("panic recovered")
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(errorResponse{
					Error:   "internal_server_error",
					Message: "An internal error occurred",
				})
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}
