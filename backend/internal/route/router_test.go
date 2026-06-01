package route

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeDB struct {
	err error
}

func (db fakeDB) Ping(context.Context) error {
	return db.err
}

func (db fakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, db.err
}

func (db fakeDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return fakeRow{err: db.err}
}

func (db fakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return fakeRows{err: db.err}, db.err
}

type fakeRow struct {
	err error
}

func (row fakeRow) Scan(...any) error {
	return row.err
}

type fakeRows struct {
	err error
}

func (rows fakeRows) Close() {}

func (rows fakeRows) Err() error {
	return rows.err
}

func (rows fakeRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (rows fakeRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (rows fakeRows) Next() bool {
	return false
}

func (rows fakeRows) Scan(...any) error {
	return rows.err
}

func (rows fakeRows) Values() ([]any, error) {
	return nil, rows.err
}

func (rows fakeRows) RawValues() [][]byte {
	return nil
}

func (rows fakeRows) Conn() *pgx.Conn {
	return nil
}

func TestHealthHandler(t *testing.T) {
	router := NewRouter(config.Config{Env: "test", Port: "3000"}, fakeDB{})
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload map[string]string
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", payload["status"])
	}

	if payload["database"] != "ok" {
		t.Fatalf("expected database ok, got %q", payload["database"])
	}
}

func TestHealthHandlerWhenDatabaseIsUnavailable(t *testing.T) {
	router := NewRouter(config.Config{Env: "test", Port: "3000"}, fakeDB{err: errors.New("down")})
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}
}

func TestStatusHandler(t *testing.T) {
	router := NewRouter(config.Config{Env: "test", Port: "3000"}, fakeDB{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload dto.StatusResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.App != "palpite-api" {
		t.Fatalf("expected app palpite-api, got %q", payload.App)
	}

	if payload.Env != "test" {
		t.Fatalf("expected env test, got %q", payload.Env)
	}

	if payload.Database != "ok" {
		t.Fatalf("expected database ok, got %q", payload.Database)
	}
}

func TestGetMatchPredictionRouteRequiresAuth(t *testing.T) {
	router := NewRouter(config.Config{Env: "test", Port: "3000"}, fakeDB{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/matches/match-1/prediction", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestDeleteMeRouteRequiresAuth(t *testing.T) {
	router := NewRouter(config.Config{Env: "test", Port: "3000"}, fakeDB{})
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/me", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}
