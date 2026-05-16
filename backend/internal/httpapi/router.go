package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type statusResponse struct {
	App       string `json:"app"`
	Database  string `json:"database"`
	Env       string `json:"env"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type datastore interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func NewRouter(cfg config.Config, db datastore) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler(db))
	mux.HandleFunc("GET /api/v1/status", statusHandler(cfg, db))
	mux.HandleFunc("GET /api/v1/groups", listGroupsHandler(cfg, db))
	mux.HandleFunc("POST /api/v1/groups", createGroupHandler(cfg, db))
	mux.HandleFunc("POST /api/v1/groups/join", joinGroupHandler(cfg, db))

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func healthHandler(db datastore) http.HandlerFunc {
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
			"status":   "ok",
		})
	}
}

func statusHandler(cfg config.Config, db datastore) http.HandlerFunc {
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

		writeJSON(w, http.StatusOK, statusResponse{
			App:       "palpitai-api",
			Database:  databaseStatus,
			Env:       cfg.Env,
			Status:    responseStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
