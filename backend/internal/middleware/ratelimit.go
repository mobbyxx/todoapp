package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultRateLimit    = 100
	authRateLimit       = 5
	generalRateLimit    = 100
	burstLimit          = 20
	authBurstLimit      = 10
	rateLimitWindow     = time.Minute
	rateLimitKeyPrefix  = "ratelimit:"
	blockedKeyPrefix    = "ratelimit:blocked:"
	blockDuration       = 15 * time.Minute
	maxFailures         = 10
	failureWindow       = 5 * time.Minute
	failureKeyPrefix    = "ratelimit:failures:"
)

type RateLimiter struct {
	redis      *redis.Client
	defaultLimit int
	authLimit    int
}

type RateLimitConfig struct {
	Limit     int
	Burst     int
	Window    time.Duration
	BlockAfter int
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis:        redis,
		defaultLimit: generalRateLimit,
		authLimit:    authRateLimit,
	}
}

func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := rl.getConfigForPath(r.URL.Path)
		
		clientID := rl.getClientID(r)
		key := rateLimitKeyPrefix + clientID + ":" + r.URL.Path
		blockedKey := blockedKeyPrefix + clientID
		
		ctx := r.Context()
		
		blocked, err := rl.redis.Exists(ctx, blockedKey).Result()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		if blocked > 0 {
			ttl, _ := rl.redis.TTL(ctx, blockedKey).Result()
			rl.writeRateLimitError(w, http.StatusTooManyRequests, 
				"rate_limit_blocked", 
				fmt.Sprintf("Too many requests. Try again in %v", ttl),
				0)
			return
		}
		
		pipe := rl.redis.Pipeline()
		getCmd := pipe.Get(ctx, key)
		pipe.Expire(ctx, key, config.Window)
		_, err = pipe.Exec(ctx)
		
		count := int64(0)
		if err == nil || err == redis.Nil {
			if err != redis.Nil {
				count, _ = strconv.ParseInt(getCmd.Val(), 10, 64)
			}
		}
		
		newCount, err := rl.redis.Incr(ctx, key).Result()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		
		if newCount == 1 {
			rl.redis.Expire(ctx, key, config.Window)
		}
		
		remaining := config.Limit - int(newCount)
		if remaining < 0 {
			remaining = 0
		}
		
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.Limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(config.Window).Unix(), 10))
		
		if int(newCount) > config.Limit+config.Burst {
			failureKey := failureKeyPrefix + clientID
			failures, _ := rl.redis.Incr(ctx, failureKey).Result()
			if failures == 1 {
				rl.redis.Expire(ctx, failureKey, failureWindow)
			}
			
			if failures >= int64(config.BlockAfter) {
				rl.redis.Set(ctx, blockedKey, "1", blockDuration)
				rl.redis.Del(ctx, failureKey)
			}
			
			rl.writeRateLimitError(w, http.StatusTooManyRequests,
				"rate_limit_exceeded",
				"Rate limit exceeded. Please slow down.",
				remaining)
			return
		}
		
		rl.redis.Del(ctx, failureKeyPrefix+clientID)
		
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) AuthRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := RateLimitConfig{
			Limit:      authRateLimit,
			Burst:      authBurstLimit,
			Window:     rateLimitWindow,
			BlockAfter: maxFailures,
		}
		
		clientID := rl.getClientID(r)
		key := rateLimitKeyPrefix + "auth:" + clientID
		blockedKey := blockedKeyPrefix + "auth:" + clientID
		
		ctx := r.Context()
		
		blocked, err := rl.redis.Exists(ctx, blockedKey).Result()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		if blocked > 0 {
			ttl, _ := rl.redis.TTL(ctx, blockedKey).Result()
			rl.writeRateLimitError(w, http.StatusTooManyRequests,
				"auth_rate_limit_blocked",
				fmt.Sprintf("Too many authentication attempts. Try again in %v", ttl),
				0)
			return
		}
		
		newCount, err := rl.redis.Incr(ctx, key).Result()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		
		if newCount == 1 {
			rl.redis.Expire(ctx, key, config.Window)
		}
		
		remaining := config.Limit - int(newCount)
		if remaining < 0 {
			remaining = 0
		}
		
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.Limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(config.Window).Unix(), 10))
		
		if int(newCount) > config.Limit {
			failureKey := failureKeyPrefix + "auth:" + clientID
			failures, _ := rl.redis.Incr(ctx, failureKey).Result()
			if failures == 1 {
				rl.redis.Expire(ctx, failureKey, failureWindow)
			}
			
			if failures >= int64(config.BlockAfter) {
				rl.redis.Set(ctx, blockedKey, "1", blockDuration)
				rl.redis.Del(ctx, failureKey)
			}
			
			rl.writeRateLimitError(w, http.StatusTooManyRequests,
				"auth_rate_limit_exceeded",
				"Too many authentication attempts. Please try again later.",
				remaining)
			return
		}
		
		rl.redis.Del(ctx, failureKeyPrefix+"auth:"+clientID)
		
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) StrictRateLimit(limit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := rl.getClientID(r)
			key := rateLimitKeyPrefix + "strict:" + clientID + ":" + r.URL.Path
			
			ctx := r.Context()
			
			newCount, err := rl.redis.Incr(ctx, key).Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			
			if newCount == 1 {
				rl.redis.Expire(ctx, key, rateLimitWindow)
			}
			
			remaining := limit - int(newCount)
			if remaining < 0 {
				remaining = 0
			}
			
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rateLimitWindow).Unix(), 10))
			
			if int(newCount) > limit {
				rl.writeRateLimitError(w, http.StatusTooManyRequests,
					"rate_limit_exceeded",
					"Rate limit exceeded. Please slow down.",
					remaining)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) getConfigForPath(path string) RateLimitConfig {
	if strings.HasPrefix(path, "/auth/") || 
	   strings.HasPrefix(path, "/api/v1/auth/") ||
	   strings.Contains(path, "/login") ||
	   strings.Contains(path, "/register") ||
	   strings.Contains(path, "/refresh") {
		return RateLimitConfig{
			Limit:      authRateLimit,
			Burst:      authBurstLimit,
			Window:     rateLimitWindow,
			BlockAfter: maxFailures,
		}
	}
	
	if strings.HasPrefix(path, "/api/") {
		return RateLimitConfig{
			Limit:      generalRateLimit,
			Burst:      burstLimit,
			Window:     rateLimitWindow,
			BlockAfter: maxFailures,
		}
	}
	
	return RateLimitConfig{
		Limit:      rl.defaultLimit,
		Burst:      burstLimit,
		Window:     rateLimitWindow,
		BlockAfter: maxFailures,
	}
}

func (rl *RateLimiter) getClientID(r *http.Request) string {
	userID := GetUserID(r.Context())
	if userID != "" {
		return "user:" + userID
	}
	
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return "apikey:" + hashString(apiKey)
	}
	
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	
	return "ip:" + ip
}

type rateLimitError struct {
	Error       string `json:"error"`
	Message     string `json:"message"`
	RetryAfter  int    `json:"retry_after,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Remaining   int    `json:"remaining,omitempty"`
}

func (rl *RateLimiter) writeRateLimitError(w http.ResponseWriter, status int, code, message string, remaining int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	response := rateLimitError{
		Error:     code,
		Message:   message,
		Remaining: remaining,
	}
	
	if status == http.StatusTooManyRequests {
		w.Header().Set("Retry-After", strconv.Itoa(int(blockDuration.Seconds())))
		response.RetryAfter = int(blockDuration.Seconds())
	}
	
	json.NewEncoder(w).Encode(response)
}

func hashString(s string) string {
	hash := uint64(14695981039346656037)
	for _, c := range s {
		hash ^= uint64(c)
		hash *= 1099511628211
	}
	return fmt.Sprintf("%x", hash)
}

func (rl *RateLimiter) GetRateLimitStatus(ctx context.Context, clientID string) (map[string]interface{}, error) {
	key := rateLimitKeyPrefix + clientID
	
	count, err := rl.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		count = 0
	} else if err != nil {
		return nil, err
	}
	
	ttl, err := rl.redis.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"limit":     generalRateLimit,
		"remaining": generalRateLimit - int(count),
		"reset_at":  time.Now().Add(ttl).Unix(),
	}, nil
}

func (rl *RateLimiter) ResetRateLimit(ctx context.Context, clientID string) error {
	key := rateLimitKeyPrefix + clientID
	blockedKey := blockedKeyPrefix + clientID
	failureKey := failureKeyPrefix + clientID
	
	pipe := rl.redis.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, blockedKey)
	pipe.Del(ctx, failureKey)
	
	_, err := pipe.Exec(ctx)
	return err
}

func (rl *RateLimiter) BlockClient(ctx context.Context, clientID string, duration time.Duration) error {
	blockedKey := blockedKeyPrefix + clientID
	return rl.redis.Set(ctx, blockedKey, "1", duration).Err()
}
