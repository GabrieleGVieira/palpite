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

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/database"
	"github.com/gabrielevieira/palpitai/backend/internal/httpapi"
	"github.com/gabrielevieira/palpitai/backend/internal/matchsync"
	"github.com/gabrielevieira/palpitai/backend/internal/realtime"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()

	db, err := database.NewPostgresPool(startupCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(startupCtx, db); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	syncCtx, syncCancel := context.WithCancel(context.Background())
	defer syncCancel()
	realtimeHub := realtime.NewHub(logger)

	if syncer, enabled := matchsync.New(cfg, db, logger); enabled {
		syncer.SetPublisher(realtimeHub)
		go syncer.Run(syncCtx)
		logger.Info("match sync enabled", "provider", "football-data.org")
	} else {
		logger.Info("match sync disabled", "reason", "FOOTBALL_DATA_TOKEN not configured")
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(cfg, db, realtimeHub),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api server started", "addr", server.Addr, "env", cfg.Env, "database", "connected")

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-shutdownCtx.Done()
	logger.Info("api server shutting down")
	syncCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("api server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("api server stopped")
}
