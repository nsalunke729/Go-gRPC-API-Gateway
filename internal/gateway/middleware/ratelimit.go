package middleware

import (
	"net/http"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimiter is a per-client token-bucket rate limiter.
// Each unique RemoteAddr gets its own bucket; buckets are created on first use.
type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*rate.Limiter
	r       rate.Limit // tokens per second
	b       int        // burst size
	log     *zap.Logger
}

// NewRateLimiter creates a RateLimiter with the given steady-state rate and burst.
func NewRateLimiter(r rate.Limit, b int, log *zap.Logger) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*rate.Limiter),
		r:       r,
		b:       b,
		log:     log,
	}
}

func (rl *RateLimiter) limiterFor(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	l, ok := rl.clients[key]
	if !ok {
		l = rate.NewLimiter(rl.r, rl.b)
		rl.clients[key] = l
	}
	return l
}

// Middleware returns an http.Handler middleware that enforces the rate limit.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		if !rl.limiterFor(key).Allow() {
			rl.log.Warn("rate limit exceeded", zap.String("client", key))
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
