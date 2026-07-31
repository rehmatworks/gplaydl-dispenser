package api

import (
	"net/http"
	"strings"

	proxycfg "gplaydl-dispenser/internal/proxy"
)

type updateProxySettingsRequest struct {
	ProxyTemplate string `json:"proxyTemplate"`
}

func (s *Server) handleAdminSettings(w http.ResponseWriter, r *http.Request) {
	encrypted, err := s.store.ProxyTemplate(r.Context())
	if err != nil {
		s.log.Error("load admin settings", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load settings")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"proxyConfigured": len(encrypted) > 0})
}

func (s *Server) handleSetProxyTemplate(w http.ResponseWriter, r *http.Request) {
	var req updateProxySettingsRequest
	if !readJSON(w, r, &req) {
		return
	}
	req.ProxyTemplate = strings.TrimSpace(req.ProxyTemplate)
	if err := proxycfg.ValidateTemplate(req.ProxyTemplate); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	encrypted, err := s.box.Encrypt(req.ProxyTemplate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not encrypt proxy setting")
		return
	}
	if err := s.store.SetProxyTemplate(r.Context(), encrypted); err != nil {
		s.log.Error("save proxy setting", "err", err)
		writeError(w, http.StatusInternalServerError, "could not save settings")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"proxyConfigured": true})
}

func (s *Server) handleClearProxyTemplate(w http.ResponseWriter, r *http.Request) {
	if err := s.store.ClearProxyTemplate(r.Context()); err != nil {
		s.log.Error("clear proxy setting", "err", err)
		writeError(w, http.StatusInternalServerError, "could not save settings")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"proxyConfigured": false})
}
