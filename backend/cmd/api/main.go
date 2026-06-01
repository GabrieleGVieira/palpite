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

	"github.com/gabrielevieira/palpitai/backend/internal/cache"
	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/database"
	"github.com/gabrielevieira/palpitai/backend/internal/realtime"
	"github.com/gabrielevieira/palpitai/backend/internal/route"
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

	redisClient, err := cache.NewRedisClient(startupCtx, cfg.RedisURL)
	if err != nil {
		logger.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("redis close failed", "error", err)
		}
	}()

	if err := database.Migrate(startupCtx, db); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	realtimeHub := realtime.NewHub(logger)
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go realtime.SubscribeRedis(appCtx, redisClient, realtimeHub, logger)

	server := &http.Server{
		Addr:              "0.0.0.0:" + cfg.Port,
		Handler:           route.NewRouter(cfg, db, route.Services{Realtime: realtimeHub, Redis: redisClient}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api server started", "addr", server.Addr, "env", cfg.Env, "database", "connected", "redis", "connected")

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-appCtx.Done()
	logger.Info("api server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("api server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("api server stopped")
}
