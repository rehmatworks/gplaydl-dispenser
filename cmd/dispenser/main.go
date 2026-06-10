package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gplaydl-dispenser/internal/api"
	"gplaydl-dispenser/internal/config"
	"gplaydl-dispenser/internal/crypto"
	"gplaydl-dispenser/internal/gplay"
	"gplaydl-dispenser/internal/mail"
	"gplaydl-dispenser/internal/store"
	"gplaydl-dispenser/web"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	box, err := crypto.NewBox(cfg.EncryptionKey)
	if err != nil {
		log.Error("crypto", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("store", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	gp := gplay.NewClient(cfg.MintConcurrency, cfg.MintTimeout, cfg.PublicURL+"/api/auth")

	mailer := mail.New(cfg.BrevoAPIKey, cfg.MailFrom, cfg.MailFromName, cfg.PublicURL, log)
	if !mailer.Enabled() {
		log.Warn("BREVO_API_KEY not set: email verification and password reset are disabled; new users are auto-verified")
	}

	server := api.NewServer(cfg, st, box, gp, mailer, web.Dist(), log)

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           server.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("gplaydl dispenser listening", "addr", cfg.Addr, "mintConcurrency", cfg.MintConcurrency)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server", "err", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
	log.Info("bye!")
}
