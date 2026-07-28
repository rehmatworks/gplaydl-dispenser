package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gplaydl-dispenser/internal/config"
	"gplaydl-dispenser/internal/crypto"
	"gplaydl-dispenser/internal/store"
)

// The dispenser no longer serves strangers: every dispense needs a linked API
// key, and drawing from the rotation needs at least one shared account behind
// that key. These tests drive the whole journey over HTTP, from claiming a
// pairing code as gplaydl would through each layer of refusal.
//
// The gplay client is nil here, so a request that clears the gates stops at
// the Play handshake. Corrupting the stored token makes that failure
// deterministic (a 502 from the decrypt step), which is exactly the evidence
// needed: 401 and 403 come from the gates, 502 means the rotation was reached.
func TestDispenseEnforcement(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	st, err := store.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	_, err = admin.Exec(ctx, `
		TRUNCATE api_keys, pairing_codes, mint_events, mint_stats_hourly,
		         mint_totals, accounts, sessions, email_tokens, users
		RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatal(err)
	}

	box, err := crypto.NewBox(bytes.Repeat([]byte{0x42}, 32))
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		SessionTTL:  15 * time.Minute,
		MintTimeout: 5 * time.Second,
		PublicURL:   "http://example.test",
		Dev:         true,
	}
	static := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("test")}}
	server := NewServer(
		cfg,
		st,
		box,
		nil,
		fs.FS(static),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	router := server.Router()

	deviceKey := "device-api-key"
	user, err := st.CreateDeviceUser(
		ctx,
		crypto.HashToken("device-secret"),
		"Test phone",
		crypto.HashToken(deviceKey),
		currentConsentVersion,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("no key means no dispense", func(t *testing.T) {
		get := performJSONRequest(router, http.MethodGet, "/api/auth", "", nil)
		if get.Code != http.StatusUnauthorized {
			t.Fatalf("GET without key: got %d, want 401", get.Code)
		}
		post := performJSONRequest(router, http.MethodPost, "/api/auth",
			`{"device":"test"}`, nil)
		if post.Code != http.StatusUnauthorized {
			t.Fatalf("POST without key: got %d, want 401", post.Code)
		}
		if !strings.Contains(post.Body.String(), "gplaydl link") {
			t.Fatalf("401 should explain how to link: %s", post.Body.String())
		}
	})

	t.Run("a made-up key is rejected outright", func(t *testing.T) {
		res := performJSONRequest(router, http.MethodPost, "/api/auth",
			`{"device":"test"}`, map[string]string{"X-Api-Key": "not-a-key"})
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("bogus key: got %d, want 401", res.Code)
		}
	})

	// The CLI's side of onboarding: turn a pairing code into a key of its own.
	var cliKey string
	t.Run("a pairing code buys the CLI an API key", func(t *testing.T) {
		if err := st.CreatePairingCode(ctx, "ABCDEFGH", user.ID, time.Minute); err != nil {
			t.Fatal(err)
		}
		claim := performJSONRequest(router, http.MethodPost, "/api/v1/pair/claim-cli",
			`{"code":"abcd-efgh","label":"test laptop"}`, nil)
		if claim.Code != http.StatusOK {
			t.Fatalf("claim-cli: got %d: %s", claim.Code, claim.Body.String())
		}
		var body struct {
			APIKey string `json:"apiKey"`
		}
		if err := json.Unmarshal(claim.Body.Bytes(), &body); err != nil || body.APIKey == "" {
			t.Fatalf("claim-cli body: %s", claim.Body.String())
		}
		cliKey = body.APIKey

		me := performJSONRequest(router, http.MethodGet, "/api/v1/me", "",
			map[string]string{"X-Api-Key": cliKey})
		if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), user.ID) {
			t.Fatalf("minted key on /me: got %d: %s", me.Code, me.Body.String())
		}

		again := performJSONRequest(router, http.MethodPost, "/api/v1/pair/claim-cli",
			`{"code":"ABCDEFGH"}`, nil)
		if again.Code != http.StatusBadRequest {
			t.Fatalf("reused code: got %d, want 400", again.Code)
		}
	})

	cliHeaders := map[string]string{"X-Api-Key": cliKey}

	t.Run("a key with nothing behind it cannot draw from the pool", func(t *testing.T) {
		res := performJSONRequest(router, http.MethodPost, "/api/auth",
			`{"device":"test"}`, cliHeaders)
		if res.Code != http.StatusForbidden {
			t.Fatalf("no accounts: got %d, want 403: %s", res.Code, res.Body.String())
		}
	})

	// Add a private account through the app's own API, then break its stored
	// token so a dispense that reaches it fails at decrypt instead of hitting
	// the nil Play client.
	aas := "aas_et/" + strings.Repeat("x", 40)
	private := performJSONRequest(router, http.MethodPost, "/api/v1/accounts",
		`{"email":"own@example.test","aasToken":"`+aas+`","visibility":"private"}`,
		map[string]string{"X-Api-Key": deviceKey})
	if private.Code != http.StatusCreated {
		t.Fatalf("create private account: got %d: %s", private.Code, private.Body.String())
	}
	if _, err := admin.Exec(ctx,
		`UPDATE accounts SET aas_token_enc = 'garbage'::bytea`); err != nil {
		t.Fatal(err)
	}

	t.Run("a private account alone does not open the pool", func(t *testing.T) {
		res := performJSONRequest(router, http.MethodPost, "/api/auth",
			`{"device":"test"}`, cliHeaders)
		if res.Code != http.StatusForbidden {
			t.Fatalf("private only: got %d, want 403: %s", res.Code, res.Body.String())
		}
	})

	t.Run("pinning your own private account is allowed", func(t *testing.T) {
		res := performJSONRequest(router, http.MethodPost,
			"/api/auth?email=own@example.test", `{"device":"test"}`, cliHeaders)
		if res.Code != http.StatusBadGateway {
			t.Fatalf("pinned draw should reach the rotation (502), got %d: %s",
				res.Code, res.Body.String())
		}
	})

	t.Run("pinning somebody else's address is not", func(t *testing.T) {
		res := performJSONRequest(router, http.MethodPost,
			"/api/auth?email=stranger@example.test", `{"device":"test"}`, cliHeaders)
		if res.Code != http.StatusNotFound {
			t.Fatalf("foreign pin: got %d, want 404: %s", res.Code, res.Body.String())
		}
	})

	t.Run("sharing an account opens the pool", func(t *testing.T) {
		if _, err := admin.Exec(ctx, `
			UPDATE accounts SET visibility = 'public', shared_at = now()
			WHERE email = 'own@example.test'`); err != nil {
			t.Fatal(err)
		}
		res := performJSONRequest(router, http.MethodPost, "/api/auth",
			`{"device":"test"}`, cliHeaders)
		if res.Code != http.StatusBadGateway {
			t.Fatalf("contributed draw should reach the rotation (502), got %d: %s",
				res.Code, res.Body.String())
		}

		get := performJSONRequest(router, http.MethodGet, "/api/auth", "", cliHeaders)
		if get.Code != http.StatusBadGateway && get.Code != http.StatusBadRequest {
			t.Fatalf("GET with key: got %d: %s", get.Code, get.Body.String())
		}
	})
}
