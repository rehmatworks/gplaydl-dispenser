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
		if user.Kind != "device" {
			writeError(w, http.StatusUnauthorized, "pair the Android app to continue")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	})
}

// requireAdmin is deliberately session-only. Device and CLI API keys may
// manage their own accounts but cannot change server-wide settings.
func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return s.requireSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userFrom(r.Context())
		if user == nil || !user.IsAdmin {
			writeError(w, http.StatusForbidden, "administrator access required")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

// maybeAPIKey attaches a user when a valid X-Api-Key header (or api_key query
// param) is present; anonymous requests pass through untouched.
func (s *Server) maybeAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := apiKeyFrom(r)
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

// requireUser accepts either a browser session or an API key, so the Android
// app and the dashboard drive the same account endpoints.
func (s *Server) requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if key := apiKeyFrom(r); key != "" {
			user, err := s.store.UserByAPIKeyHash(r.Context(), crypto.HashToken(key))
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid API key")
				return
			}
			// Apps poll these endpoints, which doubles as a liveness signal.
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = s.store.TouchUser(ctx, user.ID)
			}()
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
			return
		}
		s.requireSession(next).ServeHTTP(w, r)
	})
}

func apiKeyFrom(r *http.Request) string {
	if key := r.Header.Get("X-Api-Key"); key != "" {
		return key
	}
	return r.URL.Query().Get("api_key")
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

// dispenseRateLimit allows a burst of 15 anonymous mints and then one per
// minute per IP. The burst matters because gplaydl rotates through several
// device profiles in a single run, and a shared NAT can put a whole household
// behind one address. Users presenting a valid API key are exempt.
func (s *Server) dispenseRateLimit() func(http.Handler) http.Handler {
	return limitMiddleware(newIPLimiter(rate.Every(time.Minute), 15), true)
}

// enrollRateLimit keeps device enrolment cheap for real installs but useless
// for someone trying to farm identities.
func (s *Server) enrollRateLimit() func(http.Handler) http.Handler {
	return limitMiddleware(newIPLimiter(rate.Every(time.Minute), 10), false)
}

// authRateLimit guards credential endpoints against brute force.
func (s *Server) authRateLimit() func(http.Handler) http.Handler {
	return limitMiddleware(newIPLimiter(rate.Every(2*time.Second), 20), false)
}
