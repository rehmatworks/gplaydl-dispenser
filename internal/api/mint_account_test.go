package api

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"gplaydl-dispenser/internal/config"
	"gplaydl-dispenser/internal/crypto"
	"gplaydl-dispenser/internal/gplay"
	"gplaydl-dispenser/internal/store"
)

type recordedMint struct {
	routes  []string
	results []error
}

func (m *recordedMint) Mint(_ context.Context, _ gplay.Account, _ gplay.DeviceConfig, _, proxyURL string) (*gplay.AuthBundle, error) {
	m.routes = append(m.routes, proxyURL)
	err := m.results[len(m.routes)-1]
	if err != nil {
		return nil, err
	}
	return &gplay.AuthBundle{AuthToken: "ok"}, nil
}

func TestMintStoredAccountProxyFallback(t *testing.T) {
	box, err := crypto.NewBox(bytes.Repeat([]byte{0x42}, 32))
	if err != nil {
		t.Fatal(err)
	}
	encryptedProxy, err := box.Encrypt("http://user:password@proxy.example:1234")
	if err != nil {
		t.Fatal(err)
	}
	account := &store.Account{
		ID:          "account-id",
		Email:       "person@example.com",
		ProxyURLEnc: encryptedProxy,
	}

	t.Run("production retries proxy twice then direct", func(t *testing.T) {
		minter := &recordedMint{results: []error{errors.New("proxy one"), errors.New("proxy two"), nil}}
		server := &Server{
			cfg: &config.Config{MintTimeout: time.Second},
			box: box, gplay: minter,
		}
		bundle, err := server.mintStoredAccount(context.Background(), account, "aas_et/token", gplay.DeviceConfig{}, "en")
		if err != nil || bundle.AuthToken != "ok" {
			t.Fatalf("mint: bundle=%v err=%v", bundle, err)
		}
		want := []string{
			"http://user:password@proxy.example:1234",
			"http://user:password@proxy.example:1234",
			"",
		}
		if len(minter.routes) != len(want) {
			t.Fatalf("routes = %#v, want %#v", minter.routes, want)
		}
		for i := range want {
			if minter.routes[i] != want[i] {
				t.Fatalf("routes = %#v, want %#v", minter.routes, want)
			}
		}
	})

	t.Run("development never falls back directly", func(t *testing.T) {
		minter := &recordedMint{results: []error{errors.New("proxy failed")}}
		server := &Server{
			cfg: &config.Config{Dev: true, MintTimeout: time.Second},
			box: box, gplay: minter,
		}
		if _, err := server.mintStoredAccount(context.Background(), account, "aas_et/token", gplay.DeviceConfig{}, "en"); err == nil {
			t.Fatal("mint unexpectedly succeeded")
		}
		if len(minter.routes) != 1 || minter.routes[0] == "" {
			t.Fatalf("routes = %#v, want one proxy attempt", minter.routes)
		}
	})

	t.Run("credential rejection is not retried", func(t *testing.T) {
		minter := &recordedMint{results: []error{&gplay.CredentialError{Code: "BadAuthentication"}}}
		server := &Server{
			cfg: &config.Config{MintTimeout: time.Second},
			box: box, gplay: minter,
		}
		if _, err := server.mintStoredAccount(context.Background(), account, "aas_et/token", gplay.DeviceConfig{}, "en"); !gplay.IsCredentialError(err) {
			t.Fatalf("error = %v, want credential error", err)
		}
		if len(minter.routes) != 1 {
			t.Fatalf("routes = %#v, want one attempt", minter.routes)
		}
	})
}
