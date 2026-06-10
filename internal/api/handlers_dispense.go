package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gplaydl-dispenser/internal/gplay"
	"gplaydl-dispenser/internal/store"
)

var startedAt = time.Now()

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "gplaydl dispenser is alive!",
		"uptime":   time.Since(startedAt).Seconds(),
		"dateTime": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleDispenseAnonymous mirrors GET /api/auth of the original dispenser:
// returns {email, auth} minted with the default device profile.
func (s *Server) handleDispenseAnonymous(w http.ResponseWriter, r *http.Request) {
	locale := queryDefault(r, "locale", "en")

	dc, err := gplay.LoadDeviceConfig(s.cfg.ResourcesDir, queryDefault(r, "device", s.cfg.DefaultDevice))
	if err != nil {
		writeError(w, http.StatusBadRequest, "unknown device profile")
		return
	}

	bundle, err := s.dispense(r, dc, locale)
	if err != nil {
		s.dispenseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, gplay.AnonymousAuthBundle{
		Email: bundle.Email,
		Auth:  bundle.AuthToken,
	})
}

// handleDispenseWithConfig mirrors POST /api/auth: the caller supplies device
// properties in the body and receives the full AuthBundle.
func (s *Server) handleDispenseWithConfig(w http.ResponseWriter, r *http.Request) {
	locale := queryDefault(r, "locale", "en")

	var raw map[string]any
	if !readJSON(w, r, &raw) {
		return
	}
	if len(raw) == 0 {
		writeError(w, http.StatusBadRequest, "missing device configuration")
		return
	}

	dc := gplay.DeviceConfig{}
	for k, v := range raw {
		dc[k] = fmt.Sprintf("%v", v)
	}

	bundle, err := s.dispense(r, dc, locale)
	if err != nil {
		s.dispenseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bundle)
}

var errNoAccounts = fmt.Errorf("no accounts available")

// dispense claims accounts from the rotation and attempts the handshake,
// failing over to the next account (up to 3) on credential errors.
func (s *Server) dispense(r *http.Request, dc gplay.DeviceConfig, locale string) (*gplay.AuthBundle, error) {
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.MintTimeout)
	defer cancel()

	user := userFrom(r.Context())
	ownerID := ""
	includePublic := true
	if user != nil {
		ownerID = user.ID
		includePublic = queryDefault(r, "pool", "any") != "private"
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		account, err := s.store.NextAccount(ctx, ownerID, includePublic)
		if err != nil {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, errNoAccounts
		}

		aasToken, err := s.box.Decrypt(account.AASTokenEnc)
		if err != nil {
			s.log.Error("decrypt token", "account", account.ID, "err", err)
			lastErr = fmt.Errorf("internal error")
			continue
		}

		start := time.Now()
		bundle, mintErr := s.gplay.Mint(ctx, gplay.Account{
			Email:    account.Email,
			AASToken: aasToken,
		}, dc, locale)
		duration := time.Since(start)

		success := mintErr == nil
		errMsg := ""
		if mintErr != nil {
			errMsg = mintErr.Error()
		}

		// Bookkeeping must not be cancelled along with the request.
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = s.store.RecordMintResult(bgCtx, account.ID, success)
		_ = s.store.RecordMintEvent(bgCtx, store.MintEvent{
			AccountID:  account.ID,
			UserID:     ownerID,
			Anonymous:  user == nil,
			Success:    success,
			Error:      errMsg,
			DurationMS: int(duration.Milliseconds()),
		})
		bgCancel()

		if mintErr == nil {
			return bundle, nil
		}

		s.log.Warn("mint failed", "account", account.Email, "err", mintErr)
		lastErr = mintErr

		if ctx.Err() != nil {
			return nil, lastErr
		}
	}
	return nil, lastErr
}

func (s *Server) dispenseError(w http.ResponseWriter, err error) {
	if err == errNoAccounts {
		writeError(w, http.StatusServiceUnavailable, "no accounts available in the pool")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

// --- Stats ---

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	stats, err := s.store.Stats(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load stats")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handlePublicStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.Stats(r.Context(), "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load stats")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"publicAccounts": stats.PublicAccounts,
		"mints24h":       stats.Mints24h,
		"totalMints":     stats.TotalMints,
	})
}

func (s *Server) handleTimeline(w http.ResponseWriter, r *http.Request) {
	timeline, err := s.store.MintTimeline(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load timeline")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"timeline": timeline})
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := gplay.ListDevices(s.cfg.ResourcesDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list devices")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"devices": devices, "default": s.cfg.DefaultDevice})
}

func queryDefault(r *http.Request, key, fallback string) string {
	if v := r.URL.Query().Get(key); v != "" {
		return v
	}
	return fallback
}
