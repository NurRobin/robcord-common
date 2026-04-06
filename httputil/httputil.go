package httputil

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes a structured error response with a stable error code
// and a human-readable message: {"error": "code", "message": "text"}.
// The error code is derived from the HTTP status code.
func WriteError(w http.ResponseWriter, status int, message string) {
	code := httpCodeToErrorCode(status)
	WriteJSON(w, status, map[string]string{"error": code, "message": message})
}

// WriteErrorCode writes a structured error response with an explicit error code.
func WriteErrorCode(w http.ResponseWriter, status int, code string, message string) {
	WriteJSON(w, status, map[string]string{"error": code, "message": message})
}

func httpCodeToErrorCode(status int) string {
	switch status {
	case 400:
		return "bad_request"
	case 401:
		return "unauthorized"
	case 403:
		return "forbidden"
	case 404:
		return "not_found"
	case 409:
		return "conflict"
	case 429:
		return "rate_limited"
	case 500:
		return "internal_error"
	case 502:
		return "bad_gateway"
	case 503:
		return "service_unavailable"
	default:
		return "error"
	}
}

// RateLimiter implements a simple in-memory token bucket rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    int
	window  time.Duration
	done    chan struct{}
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a rate limiter that allows rate requests per window.
// Call Stop to release the background cleanup goroutine.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
		done:    make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// Stop terminates the background cleanup goroutine. Safe to call multiple times.
func (rl *RateLimiter) Stop() {
	select {
	case <-rl.done:
	default:
		close(rl.done)
	}
}

// Configure applies a new rate/window and resets existing buckets so the
// change takes effect immediately.
func (rl *RateLimiter) Configure(rate int, window time.Duration) {
	if rate <= 0 || window <= 0 {
		return
	}
	rl.mu.Lock()
	rl.rate = rate
	rl.window = window
	rl.buckets = make(map[string]*bucket)
	rl.mu.Unlock()
}

// Allow checks whether the given key is within the rate limit.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok || now.Sub(b.lastReset) >= rl.window {
		rl.buckets[key] = &bucket{tokens: rl.rate - 1, lastReset: now}
		return true
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-rl.done:
			return
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-2 * rl.window)
			for key, b := range rl.buckets {
				if b.lastReset.Before(cutoff) {
					delete(rl.buckets, key)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// NormalizeURLScheme converts WebSocket URL schemes to their HTTP equivalents
// (wss:// -> https://, ws:// -> http://) and trims trailing slashes.
// wss:// is replaced before ws:// to avoid "wss://" matching the "ws://" prefix
// and producing a malformed "https://" with a stray "s".
func NormalizeURLScheme(rawURL string) string {
	u := strings.TrimRight(rawURL, "/")
	u = strings.Replace(u, "wss://", "https://", 1)
	u = strings.Replace(u, "ws://", "http://", 1)
	return u
}

// GetClientIP extracts the client IP from the request.
// When trustProxy is true, it uses the first IP from X-Forwarded-For.
func GetClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			parts := strings.SplitN(fwd, ",", 2)
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// GetClientIPHeader extracts the client IP using a configurable header and
// optional CIDR-based proxy trust. When trustedCIDRs is non-empty, the header
// is only trusted if the request's remote address falls within one of the
// listed CIDR ranges. If the remote address is untrusted or the header is
// empty, the raw remote address is returned.
func GetClientIPHeader(r *http.Request, trustProxy bool, header, trustedCIDRs string) string {
	if !trustProxy {
		return remoteIP(r)
	}

	if trustedCIDRs != "" && !isFromTrustedProxy(r, trustedCIDRs) {
		return remoteIP(r)
	}

	if val := r.Header.Get(header); val != "" {
		return extractIPFromHeader(header, val)
	}

	return remoteIP(r)
}

// remoteIP extracts the IP portion from r.RemoteAddr (strips port).
func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// isFromTrustedProxy checks whether the request's remote address falls within
// any of the comma-separated CIDR ranges.
func isFromTrustedProxy(r *http.Request, cidrs string) bool {
	remote := net.ParseIP(remoteIP(r))
	if remote == nil {
		return false
	}
	for _, entry := range strings.Split(cidrs, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		_, network, err := net.ParseCIDR(entry)
		if err != nil {
			continue
		}
		if network.Contains(remote) {
			return true
		}
	}
	return false
}

// extractIPFromHeader reads the client IP from a proxy header value.
// Multi-value headers (X-Forwarded-For, Forwarded) return the leftmost entry.
// Single-value headers (X-Real-IP, CF-Connecting-IP) return the full value.
func extractIPFromHeader(header, value string) string {
	switch header {
	case "X-Forwarded-For":
		parts := strings.SplitN(value, ",", 2)
		return strings.TrimSpace(parts[0])
	case "Forwarded":
		// RFC 7239: Forwarded: for=192.0.2.1;proto=https
		for _, param := range strings.Split(value, ";") {
			param = strings.TrimSpace(param)
			if strings.HasPrefix(strings.ToLower(param), "for=") {
				ip := strings.TrimPrefix(param[4:], "\"")
				ip = strings.TrimSuffix(ip, "\"")
				// Handle IPv6 in brackets: [::1]
				ip = strings.TrimPrefix(ip, "[")
				ip = strings.TrimSuffix(ip, "]")
				// Strip port if present
				if host, _, err := net.SplitHostPort(ip); err == nil {
					return host
				}
				return ip
			}
		}
		return strings.TrimSpace(value)
	default:
		return strings.TrimSpace(value)
	}
}

// RateLimit wraps a handler with IP-based rate limiting.
func RateLimit(rl *RateLimiter, trustProxy bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := GetClientIP(r, trustProxy)
		if !rl.Allow(key) {
			WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next(w, r)
	}
}
