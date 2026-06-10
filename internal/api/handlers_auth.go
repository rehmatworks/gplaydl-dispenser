package api

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"gplaydl-dispenser/internal/crypto"
	"gplaydl-dispenser/internal/store"
)

const (
	verifyTokenTTL = 24 * time.Hour
	resetTokenTTL  = time.Hour
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

	// Without a mailer there is no way to verify, so trust the email as-is.
	user, err := s.store.CreateUser(r.Context(), req.Email, passwordHash, crypto.HashToken(apiKey), !s.mailer.Enabled())
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "an account with this email already exists")
			return
		}
		s.log.Error("create user", "err", err)
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	if s.mailer.Enabled() {
		s.sendVerificationEmail(r.Context(), user.ID, user.Email)
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

// sendVerificationEmail creates a single-use token and emails it in the
// background. Errors are logged; registration never fails because of email.
func (s *Server) sendVerificationEmail(ctx context.Context, userID, email string) {
	token, err := crypto.RandomToken(32)
	if err != nil {
		s.log.Error("verification token", "err", err)
		return
	}
	if err := s.store.CreateEmailToken(ctx, crypto.HashToken(token), userID, "verify", verifyTokenTTL); err != nil {
		s.log.Error("store verification token", "err", err)
		return
	}
	s.mailer.SendAsync("verification", func(ctx context.Context) error {
		return s.mailer.SendVerification(ctx, email, token)
	})
}

func (s *Server) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "missing token")
		return
	}

	userID, err := s.store.ConsumeEmailToken(r.Context(), crypto.HashToken(req.Token), "verify")
	if err != nil {
		writeError(w, http.StatusBadRequest, "this verification link is invalid or has expired")
		return
	}
	if err := s.store.MarkEmailVerified(r.Context(), userID); err != nil {
		s.log.Error("mark verified", "err", err)
		writeError(w, http.StatusInternalServerError, "could not verify email")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "verified"})
}

func (s *Server) handleResendVerification(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	if user.EmailVerified {
		writeError(w, http.StatusBadRequest, "email is already verified")
		return
	}
	if !s.mailer.Enabled() {
		writeError(w, http.StatusServiceUnavailable, "email delivery is not configured on this server")
		return
	}
	s.sendVerificationEmail(r.Context(), user.ID, user.Email)
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (s *Server) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Always respond with the same message so account existence isn't leaked.
	respond := func() {
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "if that email is registered, a reset link is on its way",
		})
	}

	if !s.mailer.Enabled() {
		respond()
		return
	}

	user, _, err := s.store.UserByEmail(r.Context(), req.Email)
	if err != nil {
		respond()
		return
	}

	token, err := crypto.RandomToken(32)
	if err != nil {
		s.log.Error("reset token", "err", err)
		respond()
		return
	}
	if err := s.store.CreateEmailToken(r.Context(), crypto.HashToken(token), user.ID, "reset", resetTokenTTL); err != nil {
		s.log.Error("store reset token", "err", err)
		respond()
		return
	}
	s.mailer.SendAsync("password-reset", func(ctx context.Context) error {
		return s.mailer.SendPasswordReset(ctx, user.Email, token)
	})
	respond()
}

func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "missing token")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	userID, err := s.store.ConsumeEmailToken(r.Context(), crypto.HashToken(req.Token), "reset")
	if err != nil {
		writeError(w, http.StatusBadRequest, "this reset link is invalid or has expired")
		return
	}

	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not hash password")
		return
	}
	if err := s.store.UpdatePassword(r.Context(), userID, passwordHash); err != nil {
		s.log.Error("update password", "err", err)
		writeError(w, http.StatusInternalServerError, "could not update password")
		return
	}
	// A reset also proves control of the inbox.
	_ = s.store.MarkEmailVerified(r.Context(), userID)
	// Revoke every existing session so a stolen cookie dies with the old password.
	_ = s.store.DeleteUserSessions(r.Context(), userID)

	writeJSON(w, http.StatusOK, map[string]string{"status": "password updated"})
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
