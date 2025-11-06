package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Based on:
// www.alexedwards.net/blog/how-to-rate-limit-http-requests/

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Change the map to hold values of the type visitor.
var visitors = make(map[string]*visitor)
var mu sync.Mutex

// Run a background goroutine to remove old entries from the visitors map.
func init() {
	go cleanupVisitors()
}

func (m *Middleware) LimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var clientIP string
		// Use proxy headers to get real IP
		if m.Config.TrustProxyHeaders {
			xff := r.Header.Get("X-Forwarded-For")
			if xff != "" {
				// The header may contain multiple IPs like "1.2.3.4, 5.6.7.8"
				parts := strings.Split(xff, ",")
				clientIP = strings.TrimSpace(parts[0])
			} else if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
				clientIP = xrip
			}
		}

		// User the regular source IP, even if attempt to use proxy headers results in empty string
		if clientIP == "" {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				m.Logger.Error("Limit middleware error", "error", err)
				http.Error(w, "Limit middleware error", http.StatusInternalServerError)
				return

			}
			clientIP = host
		}

		limiter := getVisitor(clientIP, m.Config.RateLimit)
		if !limiter.Allow() {
			m.Logger.Warn("Too many requests", "source ip", clientIP)
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getVisitor(ip string, rateLimit int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(1, rateLimit) // eg (1, 3) 3 requests per 1 second
		// Include the current time when creating a new visitor.
		visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}
	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	return v.limiter
}

// Every minute check the map for visitors that haven't been seen for
// more than 3 minutes and delete the entries.
func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}
