package api

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
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

func TestPairingOnlyAPI(t *testing.T) {
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
		TRUNCATE pairing_codes, mint_events, mint_stats_hourly, mint_totals,
		         accounts, sessions, email_tokens, users
		RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatal(err)
	}

	box, err := crypto.NewBox(bytes.Repeat([]byte{0x42}, 32))
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		SessionTTL: 15 * time.Minute,
		PublicURL:  "http://example.test",
		Dev:        true,
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

	t.Run("password routes are unreachable", func(t *testing.T) {
		for _, path := range []string{
			"/api/v1/register",
			"/api/v1/login",
			"/api/v1/forgot-password",
			"/api/v1/reset-password",
			"/api/v1/verify-email",
			"/api/v1/resend-verification",
			"/api/v1/me/api-key",
		} {
			response := performRequest(router, http.MethodPost, path, `{}`, nil)
			if response.Code != http.StatusNotFound {
				t.Errorf("%s: got %d, want 404", path, response.Code)
			}
		}
	})

	apiKey := "device-api-key"
	user, err := st.CreateDeviceUser(
		ctx,
		crypto.HashToken("device-secret"),
		"Test phone",
		crypto.HashToken(apiKey),
		currentConsentVersion,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("pairing creates and revokes a browser session", func(t *testing.T) {
		if err := st.CreatePairingCode(ctx, "ABCDEFGH", user.ID, time.Minute); err != nil {
			t.Fatal(err)
		}
		claim := performRequest(
			router,
			http.MethodPost,
			"/api/v1/pair/claim",
			`{"code":"abcd-efgh"}`,
			nil,
		)
		if claim.Code != http.StatusOK {
			t.Fatalf("claim: got %d: %s", claim.Code, claim.Body.String())
		}
		cookies := claim.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != sessionCookie || !cookies[0].HttpOnly {
			t.Fatalf("unexpected session cookie: %#v", cookies)
		}

		me := performRequest(router, http.MethodGet, "/api/v1/me", "", cookies[0])
		if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), `"kind":"device"`) {
			t.Fatalf("paired session: got %d: %s", me.Code, me.Body.String())
		}

		logout := performRequest(router, http.MethodPost, "/api/v1/logout", `{}`, cookies[0])
		if logout.Code != http.StatusOK {
			t.Fatalf("logout: got %d: %s", logout.Code, logout.Body.String())
		}
		me = performRequest(router, http.MethodGet, "/api/v1/me", "", cookies[0])
		if me.Code != http.StatusUnauthorized {
			t.Fatalf("revoked session: got %d, want 401", me.Code)
		}
	})

	t.Run("legacy web sessions cannot open the dashboard", func(t *testing.T) {
		var legacyID string
		err := admin.QueryRow(ctx, `
			INSERT INTO users (email, password_hash, api_key_hash, kind)
			VALUES ('legacy@example.test', 'dormant', $1, 'web')
			RETURNING id`,
			crypto.HashToken("legacy-key"),
		).Scan(&legacyID)
		if err != nil {
			t.Fatal(err)
		}
		token := "legacy-session"
		if err := st.CreateSession(ctx, crypto.HashToken(token), legacyID, time.Minute); err != nil {
			t.Fatal(err)
		}
		cookie := &http.Cookie{Name: sessionCookie, Value: token}
		me := performRequest(router, http.MethodGet, "/api/v1/me", "", cookie)
		if me.Code != http.StatusUnauthorized {
			t.Fatalf("legacy session: got %d, want 401", me.Code)
		}
	})

	t.Run("accounts are always private and cannot be made public", func(t *testing.T) {
		headers := map[string]string{"X-Api-Key": apiKey}
		token := "aas_et/" + strings.Repeat("x", 40)

		created := performJSONRequest(
			router,
			http.MethodPost,
			"/api/v1/accounts",
			`{"email":"spare@example.test","aasToken":"`+token+`"}`,
			headers,
		)
		if created.Code != http.StatusCreated {
			t.Fatalf("create account: got %d: %s", created.Code, created.Body.String())
		}
		if !strings.Contains(created.Body.String(), `"visibility":"private"`) {
			t.Fatalf("new account should be private: %s", created.Body.String())
		}
		account, err := st.AccountsByOwner(ctx, user.ID)
		if err != nil || len(account) != 1 {
			t.Fatalf("load account: %v, count %d", err, len(account))
		}

		// There is no longer any endpoint that could turn an account public.
		patch := performJSONRequest(
			router,
			http.MethodPatch,
			"/api/v1/accounts/"+account[0].ID,
			`{"visibility":"public"}`,
			headers,
		)
		if patch.Code != http.StatusMethodNotAllowed && patch.Code != http.StatusNotFound {
			t.Fatalf("PATCH should be gone: got %d", patch.Code)
		}
	})

	// Runs after the subtest above, which leaves spare@example.test owned by
	// user. Signing the same address in on a second device proves control of
	// it, so the account moves across instead of being duplicated.
	t.Run("re-authenticating an address moves it to the newest owner", func(t *testing.T) {
		secondKey := "second-device-key"
		second, err := st.CreateDeviceUser(
			ctx,
			crypto.HashToken("second-device-secret"),
			"Second phone",
			crypto.HashToken(secondKey),
			currentConsentVersion,
		)
		if err != nil {
			t.Fatal(err)
		}

		token := "aas_et/" + strings.Repeat("y", 40)
		claim := performJSONRequest(
			router,
			http.MethodPost,
			"/api/v1/accounts",
			`{"email":"spare@example.test","aasToken":"`+token+`","visibility":"private"}`,
			map[string]string{"X-Api-Key": secondKey},
		)
		if claim.Code != http.StatusCreated {
			t.Fatalf("claim from second device: got %d: %s", claim.Code, claim.Body.String())
		}

		var rows int
		if err := admin.QueryRow(ctx,
			`SELECT count(*) FROM accounts WHERE email = 'spare@example.test'`,
		).Scan(&rows); err != nil {
			t.Fatal(err)
		}
		if rows != 1 {
			t.Fatalf("rows for the address: got %d, want 1", rows)
		}

		gained, err := st.AccountsByOwner(ctx, second.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(gained) != 1 || gained[0].Email != "spare@example.test" {
			t.Fatalf("second device owns: %#v", gained)
		}

		lost, err := st.AccountsByOwner(ctx, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(lost) != 0 {
			t.Fatalf("first device should have lost the account, still has %d", len(lost))
		}

		// The new owner's token has to be the one the pool dispenses.
		stored, err := st.NextAccountForEmail(ctx, second.ID, "spare@example.test")
		if err != nil {
			t.Fatal(err)
		}
		plain, err := box.Decrypt(stored.AASTokenEnc)
		if err != nil {
			t.Fatal(err)
		}
		if plain != token {
			t.Fatal("dispensed token is not the one the new owner synced")
		}
	})

	// A case-different spelling is the same Google account, so it must not slip
	// past the constraint as a second row.
	t.Run("addresses are matched without regard to case", func(t *testing.T) {
		token := "aas_et/" + strings.Repeat("z", 40)
		mixed := performJSONRequest(
			router,
			http.MethodPost,
			"/api/v1/accounts",
			`{"email":"SPARE@Example.Test","aasToken":"`+token+`","visibility":"private"}`,
			map[string]string{"X-Api-Key": "second-device-key"},
		)
		if mixed.Code != http.StatusCreated {
			t.Fatalf("mixed case sync: got %d: %s", mixed.Code, mixed.Body.String())
		}

		var rows int
		if err := admin.QueryRow(ctx,
			`SELECT count(*) FROM accounts WHERE email = 'spare@example.test'`,
		).Scan(&rows); err != nil {
			t.Fatal(err)
		}
		if rows != 1 {
			t.Fatalf("rows after mixed case sync: got %d, want 1", rows)
		}
	})
}

func performRequest(
	handler http.Handler,
	method string,
	path string,
	body string,
	cookie *http.Cookie,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func performJSONRequest(
	handler http.Handler,
	method string,
	path string,
	body string,
	headers map[string]string,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
