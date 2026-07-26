package api

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"time"

	"gplaydl-dispenser/internal/crypto"
	"gplaydl-dispenser/internal/store"
)

const (
	pairingCodeTTL = 10 * time.Minute
	pairingCodeLen = 8
	// Ambiguous characters are left out so a code can be read off a phone
	// screen and typed without guesswork.
	pairingAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

type enrollRequest struct {
	// DeviceSecret is a high-entropy value the app generates once and keeps in
	// its private storage. It doubles as the recovery credential: presenting it
	// again re-issues the API key instead of orphaning shared accounts.
	DeviceSecret   string `json:"deviceSecret"`
	Label          string `json:"label"`
	ConsentVersion string `json:"consentVersion"`
}

// handleEnrollDevice gives an app install its own identity, so contributing to
// the pool never requires signing up for a dispenser account.
func (s *Server) handleEnrollDevice(w http.ResponseWriter, r *http.Request) {
	var req enrollRequest
	if !readJSON(w, r, &req) {
		return
	}
	req.DeviceSecret = strings.TrimSpace(req.DeviceSecret)
	req.Label = strings.TrimSpace(req.Label)

	if len(req.DeviceSecret) < 32 || len(req.DeviceSecret) > 128 {
		writeError(w, http.StatusBadRequest, "deviceSecret must be 32-128 characters of high-entropy randomness")
		return
	}
	if len(req.Label) > 64 {
		req.Label = req.Label[:64]
	}
	if req.Label == "" {
		req.Label = "Android device"
	}
	if req.ConsentVersion == "" {
		writeError(w, http.StatusBadRequest, "consentVersion is required")
		return
	}

	apiKey, err := crypto.RandomToken(32)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate API key")
		return
	}
	// The secret identifies the device; the database only ever sees its digest.
	deviceID := crypto.HashToken(req.DeviceSecret)

	user, err := s.store.CreateDeviceUser(r.Context(), deviceID, req.Label, crypto.HashToken(apiKey), req.ConsentVersion)
	if errors.Is(err, store.ErrDuplicate) {
		// Re-install or a reset app: recover the same identity.
		user, err = s.store.ReissueDeviceKey(r.Context(), deviceID, req.Label, crypto.HashToken(apiKey), req.ConsentVersion)
	}
	if err != nil {
		s.log.Error("enroll device", "err", err)
		writeError(w, http.StatusInternalServerError, "could not enroll this device")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":   user,
		"apiKey": apiKey,
	})
}

// handlePairingCode issues a short code the app can show, so a phone-enrolled
// contributor can open the web dashboard without ever setting a password.
func (s *Server) handlePairingCode(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	code, err := randomPairingCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate a pairing code")
		return
	}
	if err := s.store.CreatePairingCode(r.Context(), code, user.ID, pairingCodeTTL); err != nil {
		s.log.Error("create pairing code", "err", err)
		writeError(w, http.StatusInternalServerError, "could not create a pairing code")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"code":      code,
		"expiresAt": time.Now().UTC().Add(pairingCodeTTL).Format(time.RFC3339),
		"url":       s.cfg.PublicURL + "/pair",
	})
}

func (s *Server) handleClaimPairingCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	// Codes are shown in uppercase but nobody should have to care.
	code := strings.ToUpper(strings.NewReplacer(" ", "", "-", "").Replace(req.Code))
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing pairing code")
		return
	}

	userID, err := s.store.ConsumePairingCode(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusBadRequest, "that pairing code is invalid or has expired")
		return
	}
	user, err := s.store.UserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load the paired device")
		return
	}
	if err := s.startSession(w, r, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not start session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

// handleAppLatest advertises the current APK for the in-app update check and
// the website download button.
func (s *Server) handleAppLatest(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version":     s.cfg.AppVersion,
		"versionCode": s.cfg.AppVersionCode,
		"url":         s.cfg.AppDownloadURL,
		"sha256":      s.cfg.AppSHA256,
	})
}

func randomPairingCode() (string, error) {
	limit := big.NewInt(int64(len(pairingAlphabet)))
	buf := make([]byte, pairingCodeLen)
	for i := range buf {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		buf[i] = pairingAlphabet[n.Int64()]
	}
	return string(buf), nil
}
