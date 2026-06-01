package controller

import (
	"net/http"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/usecase"
)

type RedisStatusService interface{}

func HealthHandler(db usecase.Datastore, redis RedisStatusService) http.HandlerFunc {
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

		writeJSON(w, http.StatusOK, map[string]string{
			"database": "ok",
			"redis":    redisStatus(redis),
			"status":   "ok",
		})
	}
}

func StatusHandler(cfg config.Config, db usecase.Datastore, redis RedisStatusService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		databaseStatus := "ok"
		responseStatus := "ok"

		if db == nil {
			databaseStatus = "not_configured"
			responseStatus = "degraded"
		} else if err := db.Ping(r.Context()); err != nil {
			databaseStatus = "unavailable"
			responseStatus = "degraded"
		}

		writeJSON(w, http.StatusOK, dto.StatusResponse{
			App:       "palpite-api",
			Database:  databaseStatus,
			Env:       cfg.Env,
			Redis:     redisStatus(redis),
			Status:    responseStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func redisStatus(service RedisStatusService) string {
	if service == nil {
		return "not_configured"
	}

	return "configured"
}
