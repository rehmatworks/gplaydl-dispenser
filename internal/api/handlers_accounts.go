package api

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"gplaydl-dispenser/internal/gplay"
	"gplaydl-dispenser/internal/store"
)

// consentVersion is recorded when a device enrols, so the wording someone
// agreed to when they first stored an account is kept on file.
const currentConsentVersion = "2026-07-27"

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	accounts, err := s.store.AccountsByOwner(r.Context(), user.ID)
	if err != nil {
		s.log.Error("list accounts", "err", err)
		writeError(w, http.StatusInternalServerError, "could not list accounts")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accounts": accounts})
}

type createAccountRequest struct {
	Email    string `json:"email"`
	AASToken string `json:"aasToken"`
}

// handleCreateAccount registers or refreshes a Google account, always private
// to its owner. The Android app re-syncs the same address whenever it mints a
// new token, so this is an upsert rather than a plain insert.
func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	var req createAccountRequest
	if !readJSON(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.AASToken = strings.TrimSpace(req.AASToken)

	if _, err := mail.ParseAddress(req.Email); err != nil {
		writeError(w, http.StatusBadRequest, "invalid Google account email")
		return
	}
	if !strings.HasPrefix(req.AASToken, "aas_et/") || len(req.AASToken) < 32 {
		writeError(w, http.StatusBadRequest, "AAS token looks invalid (should start with aas_et/)")
		return
	}

	enc, err := s.box.Encrypt(req.AASToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not encrypt token")
		return
	}

	account, err := s.store.UpsertAccount(r.Context(), user.ID, req.Email, enc, "app")
	if err != nil {
		s.log.Error("save account", "err", err)
		writeError(w, http.StatusInternalServerError, "could not save account")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"account": account})
}

func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	err := s.store.DeleteAccount(r.Context(), chi.URLParam(r, "id"), user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete account")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// handleTestAccount runs a real mint against a specific account so users can
// verify their credentials work.
func (s *Server) handleTestAccount(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	account, err := s.store.AccountByID(r.Context(), chi.URLParam(r, "id"), user.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}

	aasToken, err := s.box.Decrypt(account.AASTokenEnc)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not decrypt token")
		return
	}

	dc, err := gplay.LoadDeviceConfig(s.cfg.ResourcesDir, s.cfg.DefaultDevice)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "default device profile missing")
		return
	}

	start := time.Now()
	_, mintErr := s.gplay.Mint(r.Context(), gplay.Account{
		Email:    account.Email,
		AASToken: aasToken,
	}, dc, "en")
	duration := time.Since(start)

	success := mintErr == nil
	errMsg := ""
	if mintErr != nil {
		errMsg = mintErr.Error()
	}

	_ = s.store.RecordMintResult(r.Context(), account.ID, success)
	_ = s.store.RecordMintOutcome(r.Context(), success)

	writeJSON(w, http.StatusOK, map[string]any{
		"success":    success,
		"error":      errMsg,
		"durationMs": duration.Milliseconds(),
	})
}
