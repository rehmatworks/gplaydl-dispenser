package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
// returns {email, auth} minted with the default device profile. It still
// requires a linked API key and serves only the caller's own accounts; the
// name is kept for wire compatibility with the original endpoint.
func (s *Server) handleDispenseAnonymous(w http.ResponseWriter, r *http.Request) {
	// Whether the caller may dispense at all comes before anything about how.
	if userFrom(r.Context()) == nil {
		s.dispenseError(w, errNeedsKey)
		return
	}
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
	// Whether the caller may dispense at all comes before anything about how.
	if userFrom(r.Context()) == nil {
		s.dispenseError(w, errNeedsKey)
		return
	}
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

var (
	errNoAccounts   = fmt.Errorf("no accounts available")
	errUnknownEmail = fmt.Errorf("unknown account email")
	errNeedsKey     = fmt.Errorf("api key required")
)

// dispense claims one of the caller's own accounts and attempts the handshake,
// failing over to the next (up to 3) on credential errors.
//
// Every dispense requires a linked API key, and only ever draws from accounts
// that key's owner added. With no --email it rotates over them least-recently
// used; with --email it pins the one whose address matches.
func (s *Server) dispense(r *http.Request, dc gplay.DeviceConfig, locale string) (*gplay.AuthBundle, error) {
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.MintTimeout)
	defer cancel()

	user := userFrom(r.Context())
	if user == nil {
		return nil, errNeedsKey
	}
	email := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("email")))
	ownerID := user.ID

	attempts := 3
	if email != "" {
		// A pinned account has no alternative to fail over to.
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		var account *store.Account
		var err error
		if email != "" {
			account, err = s.store.NextAccountForEmail(ctx, ownerID, email)
			if err != nil {
				return nil, errUnknownEmail
			}
		} else {
			account, err = s.store.NextAccount(ctx, ownerID)
		}
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

		bundle, mintErr := s.gplay.Mint(ctx, gplay.Account{
			Email:    account.Email,
			AASToken: aasToken,
		}, dc, locale)

		success := mintErr == nil

		// Bookkeeping must not be cancelled along with the request.
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = s.store.RecordMintResult(bgCtx, account.ID, success)
		_ = s.store.RecordMintOutcome(bgCtx, success)
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
	switch err {
	case errNoAccounts:
		writeError(w, http.StatusServiceUnavailable,
			"no accounts on this device yet: add a Google account in the gplaydl Authenticator app")
	case errUnknownEmail:
		writeError(w, http.StatusNotFound,
			"that email is not one of your registered accounts")
	case errNeedsKey:
		writeError(w, http.StatusUnauthorized,
			"this dispenser requires a linked device: install the Authenticator app from "+
				s.cfg.PublicURL+", add your Google account, then run `gplaydl link` and enter the pairing code")
	default:
		writeError(w, http.StatusBadGateway, err.Error())
	}
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
