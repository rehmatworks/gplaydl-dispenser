package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr            string
	DatabaseURL     string
	EncryptionKey   []byte // 32 bytes, AES-256-GCM
	SessionTTL      time.Duration
	MintConcurrency int           // max simultaneous Google handshakes
	MintTimeout     time.Duration // per-mint deadline
	ResourcesDir    string        // device .properties files
	DefaultDevice   string
	PublicURL       string // canonical base URL, embedded in minted auth bundles
	BrevoAPIKey     string // empty = email flows disabled, users auto-verified
	MailFrom        string
	MailFromName    string
	Dev             bool // relaxes cookie security for local development
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func Load() (*Config, error) {
	keyHex := os.Getenv("DISPENSER_ENCRYPTION_KEY")
	if keyHex == "" {
		return nil, fmt.Errorf("DISPENSER_ENCRYPTION_KEY is required (64 hex chars); generate with: openssl rand -hex 32")
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("DISPENSER_ENCRYPTION_KEY must be 64 hex characters (32 bytes)")
	}

	cfg := &Config{
		Addr:            env("DISPENSER_ADDR", ":8080"),
		DatabaseURL:     env("DATABASE_URL", "postgres://dispenser:dispenser@localhost:5466/dispenser?sslmode=disable"),
		EncryptionKey:   key,
		SessionTTL:      time.Duration(envInt("SESSION_TTL_HOURS", 24*14)) * time.Hour,
		MintConcurrency: envInt("MINT_CONCURRENCY", 64),
		MintTimeout:     time.Duration(envInt("MINT_TIMEOUT_SECONDS", 90)) * time.Second,
		ResourcesDir:    env("RESOURCES_DIR", "resources"),
		DefaultDevice:   env("DEFAULT_DEVICE", "arm64_xxhdpi"),
		PublicURL:       strings.TrimRight(env("PUBLIC_URL", "https://dispenser.gplaydl.com"), "/"),
		BrevoAPIKey:     os.Getenv("BREVO_API_KEY"),
		MailFrom:        env("MAIL_FROM", "no-reply@gplaydl.com"),
		MailFromName:    env("MAIL_FROM_NAME", "gplaydl dispenser"),
		Dev:             os.Getenv("DISPENSER_DEV") == "1",
	}
	return cfg, nil
}
