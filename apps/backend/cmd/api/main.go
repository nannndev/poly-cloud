// Command api adalah entrypoint backend Poly Cloud: merakit config, DB,
// engine rclone, service, dan HTTP server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/polycloud/platform/apps/backend/internal/auth"
	"github.com/polycloud/platform/apps/backend/internal/config"
	"github.com/polycloud/platform/apps/backend/internal/engine"
	"github.com/polycloud/platform/apps/backend/internal/httpapi"
	"github.com/polycloud/platform/apps/backend/internal/index"
	"github.com/polycloud/platform/apps/backend/internal/routing"
	"github.com/polycloud/platform/apps/backend/internal/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error("startup gagal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := newLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := index.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer store.Close()

	// v1 single-user: pastikan baris pemilik ada sebelum request pertama masuk.
	if err := store.EnsureUser(ctx, cfg.DefaultUserID, cfg.DefaultUserEmail); err != nil {
		return err
	}

	cipher, err := auth.NewCipher(cfg.TokenEncKey)
	if err != nil {
		return err
	}

	rclone := engine.NewRcloneDaemon(cfg.RcloneRCURL, cfg.RcloneRCUser, cfg.RcloneRCPass)
	if err := rclone.Ping(ctx); err != nil {
		// Daemon belum siap bukan alasan gagal start: browse index tetap jalan,
		// operasi file akan melapor PROVIDER_ERROR sampai daemon hidup.
		log.Warn("rclone daemon belum merespons", "url", cfg.RcloneRCURL, "err", err)
	}

	deps := &storage.Deps{
		Cfg:      cfg,
		Engine:   rclone,
		Accounts: store.Accounts(),
		Files:    store.Files(),
		Folders:  store.Folders(),
		Router:   routing.New(cfg.RoutingStrategy),
		OAuth:    auth.NewManager(cfg),
		Cipher:   cipher,
		Log:      log,
		BaseDir:  cfg.RemoteBaseDir,
	}

	// Model A aktif. Mengganti ke Model B = menukar impl di baris ini saja.
	files := storage.NewWholeFileStore(deps)
	accounts := storage.NewAccountService(deps)
	hub := httpapi.NewHub()

	api := httpapi.NewAPI(cfg, files, accounts, store.Folders(), store.Keys(), hub, log)
	handler := httpapi.NewRouter(cfg, api, hub, store.Keys(), httpapi.Health{
		DB: func() error {
			pctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			return store.Ping(pctx)
		},
		Engine: func() error {
			pctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			return rclone.Ping(pctx)
		},
	}, log)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 15 * time.Second,
		// Tanpa Read/WriteTimeout: upload & download bisa berjalan lama;
		// batasnya diatur lewat context request.
		IdleTimeout: 2 * time.Minute,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("polycloud API listening", "port", cfg.Port, "routing", cfg.RoutingStrategy)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutdown diminta, menutup koneksi")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
