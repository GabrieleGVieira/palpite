package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/usecase"
)

type PingService interface {
	Ping(ctx context.Context) error
}

func HealthHandler(db usecase.Datastore, redis PingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"database": "not_configured",
				"status":   "degraded",
			})
			return
		}

		if err := db.Ping(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"database": "unavailable",
				"status":   "degraded",
			})
			return
		}

		redisStatus := pingStatus(r.Context(), redis)
		if redisStatus == "unavailable" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"database": "ok",
				"redis":    redisStatus,
				"status":   "degraded",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"database": "ok",
			"redis":    redisStatus,
			"status":   "ok",
		})
	}
}

func StatusHandler(cfg config.Config, db usecase.Datastore, redis PingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		databaseStatus := "ok"
		redisStatus := pingStatus(r.Context(), redis)
		responseStatus := "ok"

		if db == nil {
			databaseStatus = "not_configured"
			responseStatus = "degraded"
		} else if err := db.Ping(r.Context()); err != nil {
			databaseStatus = "unavailable"
			responseStatus = "degraded"
		}
		if redisStatus == "unavailable" {
			responseStatus = "degraded"
		}

		writeJSON(w, http.StatusOK, dto.StatusResponse{
			App:       "palpitai-api",
			Database:  databaseStatus,
			Env:       cfg.Env,
			Redis:     redisStatus,
			Status:    responseStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func pingStatus(ctx context.Context, service PingService) string {
	if service == nil {
		return "not_configured"
	}

	if err := service.Ping(ctx); err != nil {
		return "unavailable"
	}

	return "ok"
}
