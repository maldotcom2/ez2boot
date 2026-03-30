package middleware

import (
	"encoding/json"
	"ez2boot/internal/shared"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Based on:
// www.alexedwards.net/blog/how-to-rate-limit-http-requests/
// Modified to include constructor pattern

type RateLimitConfig struct {
	Rate         rate.Limit    // token refill rate per sec e.g. 5
	Burst        int           // max burst/bucket size e.g. 10
	CleanupEvery time.Duration // how often to run cleanup e.g. time.Minute
	ExpiresAfter time.Duration // how long before a visitor is forgotten e.g. 10 * time.Minute
}

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	rconfig  RateLimitConfig
	visitors map[string]*visitor
	mu       sync.Mutex
}

func NewRateLimiter(rconfig RateLimitConfig) *RateLimiter {
	rlimit := &RateLimiter{
		rconfig:  rconfig,
		visitors: make(map[string]*visitor),
	}
	go rlimit.cleanupVisitors()
	return rlimit
}

func (m *Middleware) PublicLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := m.resolveClientIP(r)

		if !m.PublicRateLimiter.getVisitor(clientIP).Allow() {
			m.Logger.Warn("Too many requests", "domain", "middleware", "source ip", clientIP)
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Too many requests"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) PrivateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := m.resolveClientIP(r)

		if !m.PrivateRateLimiter.getVisitor(clientIP).Allow() {
			m.Logger.Warn("Too many requests", "domain", "middleware", "source ip", clientIP)
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Too many requests"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) resolveClientIP(r *http.Request) string {
	if m.Config.TrustProxyHeaders {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			return strings.TrimSpace(strings.Split(xff, ",")[0])
		}
		if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
			return xrip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		m.Logger.Error("Limit middleware error", "domain", "middleware", "error", err)
		return ""
	}
	return host
}

func (rlimit *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rlimit.mu.Lock()
	defer rlimit.mu.Unlock()

	v, exists := rlimit.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rlimit.rconfig.Rate, rlimit.rconfig.Burst)
		// Include the current time when creating a new visitor.
		rlimit.visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}
	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	return v.limiter
}

// Peridically check map for stale users and remove
// Check constructor for intervals
func (rlimit *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rlimit.rconfig.CleanupEvery)

		rlimit.mu.Lock()
		for ip, v := range rlimit.visitors {
			if time.Since(v.lastSeen) > rlimit.rconfig.ExpiresAfter {
				delete(rlimit.visitors, ip)
			}
		}
		rlimit.mu.Unlock()
	}
}
