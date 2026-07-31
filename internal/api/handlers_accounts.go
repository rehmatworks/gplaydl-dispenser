package api

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"gplaydl-dispenser/internal/gplay"
	proxycfg "gplaydl-dispenser/internal/proxy"
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

	response := map[string]any{"account": account}
	if warning := s.assignProxyIfMissing(r.Context(), account); warning != "" {
		response["proxyWarning"] = warning
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s *Server) assignProxyIfMissing(ctx context.Context, account *store.Account) string {
	if account.ProxyConfigured {
		return ""
	}

	templateEnc, err := s.store.ProxyTemplate(ctx)
	if err != nil {
		s.log.Error("load proxy template for account", "account", account.ID, "err", err)
		return "account saved, but proxy assignment could not be completed"
	}
	if len(templateEnc) == 0 {
		return ""
	}

	template, err := s.box.Decrypt(templateEnc)
	if err != nil {
		s.log.Error("decrypt proxy template", "account", account.ID, "err", err)
		return "account saved, but proxy assignment could not be completed"
	}
	concreteURL, err := proxycfg.Expand(template)
	if err != nil {
		s.log.Error("expand proxy template", "account", account.ID, "err", err)
		return "account saved, but proxy assignment could not be completed"
	}

	testedAt := time.Now().UTC()
	testStatus := "passed"
	warning := ""
	if err := s.probeProxy(ctx, concreteURL); err != nil {
		testStatus = "failed"
		warning = "account saved with a proxy that failed its connectivity test"
		s.log.Warn("assigned proxy test failed", "account", account.ID, "err", err)
	}

	proxyEnc, err := s.box.Encrypt(concreteURL)
	if err != nil {
		s.log.Error("encrypt account proxy", "account", account.ID, "err", err)
		return "account saved, but proxy assignment could not be completed"
	}
	assigned, err := s.store.SetAccountProxyIfMissing(ctx, account.ID, proxyEnc, testStatus, testedAt)
	if err != nil {
		s.log.Error("save account proxy", "account", account.ID, "err", err)
		return "account saved, but proxy assignment could not be completed"
	}
	if assigned {
		account.ProxyConfigured = true
		account.ProxyTestStatus = testStatus
		account.ProxyTestedAt = &testedAt
		if testStatus == "failed" {
			account.ProxyFailureCount = 1
			account.LastProxyFailureAt = &testedAt
		}
		return warning
	}

	// Another concurrent sync won the conditional assignment. Return metadata
	// for the proxy that was actually persisted, not the discarded candidate.
	persisted, err := s.store.AccountByID(ctx, account.ID, account.OwnerID)
	if err != nil {
		s.log.Error("reload concurrent proxy assignment", "account", account.ID, "err", err)
		return "account saved, but proxy assignment status could not be loaded"
	}
	account.ProxyURLEnc = persisted.ProxyURLEnc
	account.ProxyConfigured = persisted.ProxyConfigured
	account.ProxyTestStatus = persisted.ProxyTestStatus
	account.ProxyTestedAt = persisted.ProxyTestedAt
	account.ProxyFailureCount = persisted.ProxyFailureCount
	account.LastProxyFailureAt = persisted.LastProxyFailureAt
	if persisted.ProxyTestStatus == "failed" {
		return "account saved with a proxy that failed its connectivity test"
	}
	return ""
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
	_, mintErr := s.mintStoredAccount(r.Context(), account, aasToken, dc, "en")
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
