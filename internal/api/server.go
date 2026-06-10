package api

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"gplaydl-dispenser/internal/config"
	"gplaydl-dispenser/internal/crypto"
	"gplaydl-dispenser/internal/gplay"
	"gplaydl-dispenser/internal/mail"
	"gplaydl-dispenser/internal/store"
)

type Server struct {
	cfg    *config.Config
	store  *store.Store
	box    *crypto.Box
	gplay  *gplay.Client
	mailer *mail.Mailer
	log    *slog.Logger
	static fs.FS
}

func NewServer(cfg *config.Config, st *store.Store, box *crypto.Box, gp *gplay.Client, mailer *mail.Mailer, static fs.FS, log *slog.Logger) *Server {
	return &Server{cfg: cfg, store: st, box: box, gplay: gp, mailer: mailer, static: static, log: log}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(requestLogger(s.log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(2 * time.Minute))
	r.Use(securityHeaders)

	// Aurora Store compatible dispense endpoints (anonymous or API key).
	r.Group(func(r chi.Router) {
		r.Use(s.maybeAPIKey)
		r.Use(s.dispenseRateLimit())
		r.Get("/api/auth", s.handleDispenseAnonymous)
		r.Post("/api/auth", s.handleDispenseWithConfig)
	})

	r.Get("/api/health", s.handleHealth)

	// Web app API.
	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(s.authRateLimit())
			r.Post("/register", s.handleRegister)
			r.Post("/login", s.handleLogin)
			r.Post("/verify-email", s.handleVerifyEmail)
			r.Post("/forgot-password", s.handleForgotPassword)
			r.Post("/reset-password", s.handleResetPassword)
		})

		r.Group(func(r chi.Router) {
			r.Use(s.requireSession)
			r.Post("/logout", s.handleLogout)
			r.Get("/me", s.handleMe)
			r.Post("/me/api-key", s.handleRotateAPIKey)
			r.Post("/resend-verification", s.handleResendVerification)

			r.Get("/accounts", s.handleListAccounts)
			r.Post("/accounts", s.handleCreateAccount)
			r.Patch("/accounts/{id}", s.handleUpdateAccount)
			r.Delete("/accounts/{id}", s.handleDeleteAccount)
			r.Post("/accounts/{id}/test", s.handleTestAccount)

			r.Get("/stats", s.handleStats)
			r.Get("/timeline", s.handleTimeline)
			r.Get("/devices", s.handleDevices)
		})

		// Public landing-page stats (no auth, modest rate limit).
		r.Group(func(r chi.Router) {
			r.Use(s.authRateLimit())
			r.Get("/public-stats", s.handlePublicStats)
		})
	})

	// SPA frontend.
	r.NotFound(s.serveStatic)

	return r
}

// serveStatic serves the embedded frontend with an SPA index.html fallback.
func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	if f, err := s.static.Open(path); err == nil {
		f.Close()
		if path != "index.html" && !strings.HasPrefix(path, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		} else if strings.HasPrefix(path, "assets/") {
			// Vite emits content-hashed asset names.
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		http.FileServerFS(s.static).ServeHTTP(w, r)
		return
	}

	// SPA fallback
	index, err := fs.ReadFile(s.static, "index.html")
	if err != nil {
		http.Error(w, "frontend not built", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(index)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		next.ServeHTTP(w, r)
	})
}

func requestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			log.Info("http",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).Round(time.Millisecond).String(),
				"ip", r.RemoteAddr,
			)
		})
	}
}

// --- JSON helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}
