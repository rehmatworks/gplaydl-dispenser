package api

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"gplaydl-dispenser/internal/crypto"
	"gplaydl-dispenser/internal/store"
)

type ctxKey int

const userKey ctxKey = 1

const sessionCookie = "dispenser_session"

func userFrom(ctx context.Context) *store.User {
	u, _ := ctx.Value(userKey).(*store.User)
	return u
}

// requireSession authenticates browser requests via the session cookie.
func (s *Server) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		user, err := s.store.UserBySession(r.Context(), crypto.HashToken(cookie.Value))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "session expired")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	})
}

// maybeAPIKey attaches a user when a valid X-Api-Key header (or api_key query
// param) is present; anonymous requests pass through untouched.
func (s *Server) maybeAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Api-Key")
		if key == "" {
			key = r.URL.Query().Get("api_key")
		}
		if key != "" {
			user, err := s.store.UserByAPIKeyHash(r.Context(), crypto.HashToken(key))
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid API key")
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), userKey, user))
		}
		next.ServeHTTP(w, r)
	})
}

// ipLimiter is a self-pruning per-IP token bucket map.
type ipLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPLimiter(r rate.Limit, burst int) *ipLimiter {
	l := &ipLimiter{
		visitors: map[string]*visitor{},
		rate:     r,
		burst:    burst,
	}
	go func() {
		for range time.Tick(5 * time.Minute) {
			l.mu.Lock()
			for ip, v := range l.visitors {
				if time.Since(v.lastSeen) > 30*time.Minute {
					delete(l.visitors, ip)
				}
			}
			l.mu.Unlock()
		}
	}()
	return l
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	l.mu.Unlock()
	return v.limiter.Allow()
}

func limitMiddleware(l *ipLimiter, exemptAuthenticated bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exemptAuthenticated && userFrom(r.Context()) != nil {
				next.ServeHTTP(w, r)
				return
			}
			if !l.allow(clientIP(r)) {
				writeError(w, http.StatusTooManyRequests, "too many requests, try again later")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// dispenseRateLimit mirrors the original dispenser: ~5 anonymous mints per
// 10 minutes per IP. Users presenting a valid API key are exempt.
func (s *Server) dispenseRateLimit() func(http.Handler) http.Handler {
	return limitMiddleware(newIPLimiter(rate.Every(2*time.Minute), 5), true)
}

// authRateLimit guards credential endpoints against brute force.
func (s *Server) authRateLimit() func(http.Handler) http.Handler {
	return limitMiddleware(newIPLimiter(rate.Every(2*time.Second), 20), false)
}
