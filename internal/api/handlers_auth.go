package api

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"gplaydl-dispenser/internal/crypto"
	"gplaydl-dispenser/internal/store"
)

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if !readJSON(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if _, err := mail.ParseAddress(req.Email); err != nil {
		writeError(w, http.StatusBadRequest, "invalid email address")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	apiKey, err := crypto.RandomToken(32)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate API key")
		return
	}

	user, err := s.store.CreateUser(r.Context(), req.Email, passwordHash, crypto.HashToken(apiKey))
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "an account with this email already exists")
			return
		}
		s.log.Error("create user", "err", err)
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	if err := s.startSession(w, r, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not start session")
		return
	}

	// The plaintext API key is only shown once, at registration.
	writeJSON(w, http.StatusCreated, map[string]any{
		"user":   user,
		"apiKey": apiKey,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if !readJSON(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, passwordHash, err := s.store.UserByEmail(r.Context(), req.Email)
	if err != nil || !crypto.CheckPassword(passwordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := s.startSession(w, r, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not start session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		_ = s.store.DeleteSession(r.Context(), crypto.HashToken(cookie.Value))
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   !s.cfg.Dev,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"user": userFrom(r.Context())})
}

// handleRotateAPIKey issues a fresh API key, invalidating the previous one.
func (s *Server) handleRotateAPIKey(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	apiKey, err := crypto.RandomToken(32)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate API key")
		return
	}
	if err := s.store.RotateAPIKey(r.Context(), user.ID, crypto.HashToken(apiKey)); err != nil {
		writeError(w, http.StatusInternalServerError, "could not rotate API key")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"apiKey": apiKey})
}

func (s *Server) startSession(w http.ResponseWriter, r *http.Request, userID string) error {
	token, err := crypto.RandomToken(32)
	if err != nil {
		return err
	}
	if err := s.store.CreateSession(r.Context(), crypto.HashToken(token), userID, s.cfg.SessionTTL); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   !s.cfg.Dev,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}
