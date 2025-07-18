package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"discord-user-api/config"
	"discord-user-api/models"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) IsAllowed(clientID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	if requests, exists := rl.requests[clientID]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[clientID] = validRequests
	}

	if len(rl.requests[clientID]) >= rl.limit {
		return false
	}

	rl.requests[clientID] = append(rl.requests[clientID], now)
	return true
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func RateLimit(limiter *RateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			clientID := getClientID(r)
			
			if !limiter.IsAllowed(clientID) {
				log.Printf("🚫 Rate limit exceeded for client: %s", clientID)
				
				response := models.APIResponse{
					Success:   false,
					Error:     "Rate limit exceeded",
					Message:   "Too many requests, please try again later",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(response)
				return
			}

			next(w, r)
		}
	}
}

func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		responseWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		log.Printf("🌐 %s %s %s - Başlatıldı", r.Method, r.URL.Path, getClientID(r))
		
		next(responseWriter, r)
		
		duration := time.Since(start)
		log.Printf("✅ %s %s - %d - %v", r.Method, r.URL.Path, responseWriter.statusCode, duration)
	}
}

func RequestID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		r.Header.Set("X-Request-ID", requestID)
		w.Header().Set("X-Request-ID", requestID)
		
		next(w, r)
	}
}

func Security(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		next(w, r)
	}
}

func Recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("💥 Panic recovered: %v", err)
				
				response := models.APIResponse{
					Success:   false,
					Error:     "Internal server error",
					Message:   "An unexpected error occurred",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
			}
		}()

		next(w, r)
	}
}

func APIKey(validKeys []string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if len(validKeys) == 0 {
				next(w, r)
				return
			}

			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = r.URL.Query().Get("api_key")
			}

			valid := false
			for _, key := range validKeys {
				if key == apiKey {
					valid = true
					break
				}
			}

			if !valid {
				log.Printf("🔑 Invalid API key: %s", apiKey)
				
				response := models.APIResponse{
					Success:   false,
					Error:     "Invalid API key",
					Message:   "Please provide a valid API key",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(response)
				return
			}

			next(w, r)
		}
	}
}

func Compose(middlewares ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

func getClientID(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	return strings.Split(r.RemoteAddr, ":")[0]
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func CreateRateLimiter(cfg *config.Config) *RateLimiter {
	if !cfg.RateLimit.Enabled {
		return nil
	}
	
	window := time.Duration(60/cfg.RateLimit.RequestsPerMinute) * time.Second
	return NewRateLimiter(cfg.RateLimit.BurstSize, window)
} 