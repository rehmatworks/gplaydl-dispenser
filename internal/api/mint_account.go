package api

import (
	"context"
	"errors"
	"time"

	"gplaydl-dispenser/internal/gplay"
	"gplaydl-dispenser/internal/store"
)

// mintStoredAccount applies the account's network policy. Production retries a
// proxied account once through the same assigned proxy, then makes one direct
// attempt. Development deliberately makes no direct fallback.
func (s *Server) mintStoredAccount(parent context.Context, account *store.Account, aasToken string, dc gplay.DeviceConfig, locale string) (*gplay.AuthBundle, error) {
	proxyURL := ""
	if len(account.ProxyURLEnc) > 0 {
		decrypted, err := s.box.Decrypt(account.ProxyURLEnc)
		if err != nil {
			s.log.Error("decrypt account proxy", "account", account.ID, "err", err)
			return nil, errors.New("could not load account network settings")
		}
		proxyURL = decrypted
	}

	routes := []string{proxyURL}
	if proxyURL != "" && !s.cfg.Dev {
		routes = []string{proxyURL, proxyURL, ""}
	}

	ctx := parent
	cancel := func() {}
	if s.cfg.MintTimeout > 0 {
		ctx, cancel = context.WithTimeout(parent, s.cfg.MintTimeout)
	}
	defer cancel()

	credential := gplay.Account{Email: account.Email, AASToken: aasToken}
	var lastErr error
	for index, route := range routes {
		attemptCtx, attemptCancel := mintAttemptContext(ctx, len(routes)-index)
		bundle, err := s.gplay.Mint(attemptCtx, credential, dc, locale, route)
		attemptCancel()
		if route != "" {
			if err == nil {
				s.recordProxyResult(account.ID, true, index+1)
			} else if gplay.IsProxyConnectionError(err) {
				s.recordProxyResult(account.ID, false, index+1)
			} else if !errors.Is(err, context.Canceled) {
				// Credential/application responses prove the proxy route
				// worked even though the mint itself did not.
				s.recordProxyResult(account.ID, true, index+1)
			}
		}
		if err == nil {
			return bundle, nil
		}
		lastErr = err

		if gplay.IsCredentialError(err) || ctx.Err() != nil {
			return nil, err
		}
	}
	if lastErr == nil {
		lastErr = errors.New("mint failed")
	}
	return nil, lastErr
}

func (s *Server) recordProxyResult(accountID string, success bool, attempt int) {
	if s.store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := s.store.RecordProxyResult(ctx, accountID, success)
	if err != nil {
		s.log.Error("record proxy result", "account", accountID, "success", success, "err", err)
		return
	}
	if success {
		return
	}
	s.log.Warn("proxy connection failed",
		"account", accountID,
		"attempt", attempt,
		"consecutiveFailures", count,
	)
}

// Split the remaining deadline across remaining attempts so a stalled proxy
// cannot consume the direct fallback's entire budget.
func mintAttemptContext(ctx context.Context, attemptsRemaining int) (context.Context, context.CancelFunc) {
	deadline, ok := ctx.Deadline()
	if !ok || attemptsRemaining <= 1 {
		return context.WithCancel(ctx)
	}
	budget := time.Until(deadline) / time.Duration(attemptsRemaining)
	if budget <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, budget)
}
