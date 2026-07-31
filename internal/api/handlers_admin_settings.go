package api

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	proxycfg "gplaydl-dispenser/internal/proxy"
	"gplaydl-dispenser/internal/store"
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

type proxyBackfillResult struct {
	Targeted int64 `json:"targeted"`
	Updated  int64 `json:"updated"`
	Passed   int64 `json:"passed"`
	Failed   int64 `json:"failed"`
	Errors   int64 `json:"errors"`
}

func (s *Server) handleBackfillAccountProxies(w http.ResponseWriter, r *http.Request) {
	templateEnc, err := s.store.ProxyTemplate(r.Context())
	if err != nil {
		s.log.Error("load proxy template for backfill", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load proxy settings")
		return
	}
	if len(templateEnc) == 0 {
		writeError(w, http.StatusConflict, "configure a proxy template before backfilling accounts")
		return
	}
	template, err := s.box.Decrypt(templateEnc)
	if err != nil {
		s.log.Error("decrypt proxy template for backfill", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load proxy settings")
		return
	}

	accounts, err := s.store.AccountsNeedingProxyBackfill(r.Context())
	if err != nil {
		s.log.Error("list accounts for proxy backfill", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load accounts")
		return
	}
	result := proxyBackfillResult{Targeted: int64(len(accounts))}
	if len(accounts) == 0 {
		writeJSON(w, http.StatusOK, result)
		return
	}

	jobs := make(chan *store.Account)
	var updated, passed, failed, failedUpdates atomic.Int64
	workerCount := min(8, len(accounts))
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for account := range jobs {
				concreteURL, err := proxycfg.Expand(template)
				if err != nil {
					failedUpdates.Add(1)
					s.log.Error("expand proxy template during backfill", "account", account.ID, "err", err)
					continue
				}

				testedAt := time.Now().UTC()
				testStatus := "passed"
				if err := s.probeProxy(r.Context(), concreteURL); err != nil {
					testStatus = "failed"
				}
				encrypted, err := s.box.Encrypt(concreteURL)
				if err != nil {
					failedUpdates.Add(1)
					s.log.Error("encrypt proxy during backfill", "account", account.ID, "err", err)
					continue
				}
				if err := s.store.ReplaceAccountProxy(
					r.Context(), account.ID, encrypted, testStatus, testedAt,
				); err != nil {
					failedUpdates.Add(1)
					s.log.Error("save proxy during backfill", "account", account.ID, "err", err)
					continue
				}
				updated.Add(1)
				if testStatus == "passed" {
					passed.Add(1)
				} else {
					failed.Add(1)
					s.log.Warn("backfilled proxy test failed", "account", account.ID)
				}
			}
		}()
	}
	for _, account := range accounts {
		jobs <- account
	}
	close(jobs)
	workers.Wait()

	result.Updated = updated.Load()
	result.Passed = passed.Load()
	result.Failed = failed.Load()
	result.Errors = failedUpdates.Load()
	writeJSON(w, http.StatusOK, result)
}
