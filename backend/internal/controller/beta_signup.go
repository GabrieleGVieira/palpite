package controller

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/usecase"
)

type BetaAndroidSignupUsecase interface {
	Signup(ctx context.Context, input usecase.BetaAndroidSignupInput) (usecase.BetaAndroidSignupResult, error)
}

type betaAndroidSignupRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Consent bool   `json:"consent"`
}

type betaAndroidSignupResponse struct {
	Success     bool   `json:"success"`
	RedirectURL string `json:"redirectUrl,omitempty"`
	Status      string `json:"status,omitempty"`
	Message     string `json:"message,omitempty"`
}

var betaAndroidLimiter = newSimpleRateLimiter(5, 10*time.Minute)

func BetaAndroidSignupHandler(signups BetaAndroidSignupUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !betaAndroidLimiter.Allow(clientIP(r)) {
			writeError(w, http.StatusTooManyRequests, "Muitas tentativas. Tente novamente em alguns minutos.")
			return
		}

		var request betaAndroidSignupRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "Dados invalidos.")
			return
		}

		result, err := signups.Signup(r.Context(), usecase.BetaAndroidSignupInput{
			Name:    request.Name,
			Email:   request.Email,
			Consent: request.Consent,
		})
		if err == nil {
			writeJSON(w, http.StatusOK, betaAndroidSignupResponse{
				Success:     true,
				RedirectURL: result.RedirectURL,
				Status:      result.Status,
			})
			return
		}

		switch {
		case errors.Is(err, usecase.ErrBetaAndroidInvalidEmail):
			writeError(w, http.StatusBadRequest, "Informe um e-mail valido.")
		case errors.Is(err, usecase.ErrBetaAndroidConsentRequired):
			writeError(w, http.StatusBadRequest, "Confirme o consentimento para receber acesso beta e comunicacoes sobre o app.")
		default:
			slog.Error("beta android signup failed", "error", err)
			writeError(w, http.StatusInternalServerError, "Nao foi possivel processar o cadastro agora.")
		}
	}
}

type simpleRateLimiter struct {
	limit  int
	window time.Duration
	mutex  sync.Mutex
	hits   map[string][]time.Time
}

func newSimpleRateLimiter(limit int, window time.Duration) *simpleRateLimiter {
	return &simpleRateLimiter{
		limit:  limit,
		window: window,
		hits:   make(map[string][]time.Time),
	}
}

func (limiter *simpleRateLimiter) Allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-limiter.window)

	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	recent := limiter.hits[key][:0]
	for _, hit := range limiter.hits[key] {
		if hit.After(cutoff) {
			recent = append(recent, hit)
		}
	}

	if len(recent) >= limiter.limit {
		limiter.hits[key] = recent
		return false
	}

	limiter.hits[key] = append(recent, now)
	return true
}

func clientIP(r *http.Request) string {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
