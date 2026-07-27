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

	t.Run("public visibility requires current consent", func(t *testing.T) {
		headers := map[string]string{"X-Api-Key": apiKey}
		token := "aas_et/" + strings.Repeat("x", 40)

		publicWithoutConsent := performJSONRequest(
			router,
			http.MethodPost,
			"/api/v1/accounts",
			`{"email":"spare@example.test","aasToken":"`+token+`","visibility":"public"}`,
			headers,
		)
		if publicWithoutConsent.Code != http.StatusBadRequest {
			t.Fatalf("public without consent: got %d", publicWithoutConsent.Code)
		}

		private := performJSONRequest(
			router,
			http.MethodPost,
			"/api/v1/accounts",
			`{"email":"spare@example.test","aasToken":"`+token+`","visibility":"private"}`,
			headers,
		)
		if private.Code != http.StatusCreated {
			t.Fatalf("create private: got %d: %s", private.Code, private.Body.String())
		}
		account, err := st.AccountsByOwner(ctx, user.ID)
		if err != nil || len(account) != 1 {
			t.Fatalf("load private account: %v, count %d", err, len(account))
		}

		publicWithoutConsent = performJSONRequest(
			router,
			http.MethodPatch,
			"/api/v1/accounts/"+account[0].ID,
			`{"visibility":"public"}`,
			headers,
		)
		if publicWithoutConsent.Code != http.StatusBadRequest {
			t.Fatalf("publish without consent: got %d", publicWithoutConsent.Code)
		}

		publish := performJSONRequest(
			router,
			http.MethodPatch,
			"/api/v1/accounts/"+account[0].ID,
			`{"visibility":"public","consentVersion":"`+currentConsentVersion+`"}`,
			headers,
		)
		if publish.Code != http.StatusOK ||
			!strings.Contains(publish.Body.String(), `"visibility":"public"`) {
			t.Fatalf("publish with consent: got %d: %s", publish.Code, publish.Body.String())
		}

		makePrivate := performJSONRequest(
			router,
			http.MethodPatch,
			"/api/v1/accounts/"+account[0].ID,
			`{"visibility":"private"}`,
			headers,
		)
		if makePrivate.Code != http.StatusOK ||
			!strings.Contains(makePrivate.Body.String(), `"visibility":"private"`) {
			t.Fatalf("make private: got %d: %s", makePrivate.Code, makePrivate.Body.String())
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
