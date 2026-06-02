package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/usecase"
)

type BetaAndroidSignupUsecase interface {
	Signup(ctx context.Context, input usecase.BetaAndroidSignupInput) (usecase.BetaAndroidSignupResult, error)
}

type BetaTesterApprovalUsecase interface {
	ApproveBetaTester(ctx context.Context, input usecase.ApproveBetaTesterInput) (usecase.ApproveBetaTesterResult, error)
	PreviewBetaTesterApproval(ctx context.Context, testerID string, token string) (usecase.BetaTesterApprovalPreview, error)
	ConfirmBetaTesterApproval(ctx context.Context, input usecase.ConfirmBetaTesterApprovalInput) (usecase.ApproveBetaTesterResult, error)
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

type betaTesterApprovalResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

type betaTesterApprovalPreviewResponse struct {
	TesterID string `json:"testerId"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Status   string `json:"status"`
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

func ApproveBetaTesterHandler(cfg config.Config, approvals BetaTesterApprovalUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, displayName, err := userIDAndDisplayNameFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Nao autorizado.")
			return
		}

		approvedBy := strings.TrimSpace(displayName)
		if approvedBy == "" {
			approvedBy = userID
		}

		result, err := approvals.ApproveBetaTester(r.Context(), usecase.ApproveBetaTesterInput{
			TesterID:   r.PathValue("id"),
			ApprovedBy: approvedBy,
		})
		if err == nil {
			writeJSON(w, http.StatusOK, betaTesterApprovalResponse{
				Success: true,
				Status:  result.Status,
			})
			return
		}

		switch {
		case apperrors.IsNotFound(err):
			writeError(w, http.StatusNotFound, "Beta tester nao encontrado.")
		case apperrors.IsValidation(err):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			slog.Error("beta tester approval failed", "error", err)
			writeError(w, http.StatusInternalServerError, "Nao foi possivel aprovar o beta tester agora.")
		}
	}
}

func BetaTesterApprovalConfirmationPageHandler(approvals BetaTesterApprovalUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		preview, err := approvals.PreviewBetaTesterApproval(r.Context(), r.PathValue("id"), r.URL.Query().Get("token"))
		if err != nil {
			writeApprovalHTML(w, http.StatusBadRequest, approvalErrorHTML(err))
			return
		}

		writeApprovalHTML(w, http.StatusOK, approvalConfirmationHTML(preview, r.URL.Query().Get("token")))
	}
}

func ConfirmBetaTesterApprovalByLinkHandler(approvals BetaTesterApprovalUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		preview, err := approvals.PreviewBetaTesterApproval(r.Context(), r.PathValue("id"), r.URL.Query().Get("token"))
		if err != nil {
			writeApprovalHTML(w, http.StatusBadRequest, approvalErrorHTML(err))
			return
		}

		result, err := approvals.ConfirmBetaTesterApproval(r.Context(), usecase.ConfirmBetaTesterApprovalInput{
			TesterID:   r.PathValue("id"),
			Token:      r.URL.Query().Get("token"),
			ApprovedBy: "signed_approval_link",
		})
		if err != nil {
			writeApprovalHTML(w, http.StatusBadRequest, approvalErrorHTML(err))
			return
		}

		writeApprovalHTML(w, http.StatusOK, approvalSuccessHTML(preview, result.Status))
	}
}

func BetaTesterApprovalPreviewAPIHandler(approvals BetaTesterApprovalUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		preview, err := approvals.PreviewBetaTesterApproval(r.Context(), r.PathValue("id"), r.URL.Query().Get("token"))
		if err != nil {
			writeApprovalAPIError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, betaTesterApprovalPreviewResponse{
			TesterID: preview.TesterID,
			Name:     preview.Name,
			Email:    preview.Email,
			Status:   preview.Status,
		})
	}
}

func ConfirmBetaTesterApprovalByLinkAPIHandler(approvals BetaTesterApprovalUsecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := approvals.ConfirmBetaTesterApproval(r.Context(), usecase.ConfirmBetaTesterApprovalInput{
			TesterID:   r.PathValue("id"),
			Token:      r.URL.Query().Get("token"),
			ApprovedBy: "signed_approval_link",
		})
		if err != nil {
			writeApprovalAPIError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, betaTesterApprovalResponse{
			Success: true,
			Status:  result.Status,
		})
	}
}

func writeApprovalAPIError(w http.ResponseWriter, err error) {
	switch {
	case apperrors.IsNotFound(err):
		writeError(w, http.StatusNotFound, "Beta tester nao encontrado.")
	case apperrors.IsForbidden(err):
		writeError(w, http.StatusForbidden, err.Error())
	case apperrors.IsValidation(err):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		slog.Error("beta tester approval link failed", "error", err)
		writeError(w, http.StatusInternalServerError, "Nao foi possivel processar a aprovacao agora.")
	}
}

func writeApprovalHTML(w http.ResponseWriter, statusCode int, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(body))
}

func approvalConfirmationHTML(preview usecase.BetaTesterApprovalPreview, token string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Confirmar aprovacao beta Android</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 0; padding: 32px; background: #f7f8f4; color: #1c2a22; }
    main { max-width: 560px; margin: 0 auto; background: #fff; border: 1px solid #dfe8d9; border-radius: 8px; padding: 24px; }
    button { border: 0; border-radius: 8px; padding: 12px 18px; background: #1f7a4a; color: #fff; font-weight: 700; cursor: pointer; }
    .muted { color: #66736a; }
  </style>
</head>
<body>
  <main>
    <h1>Confirmar aprovacao</h1>
    <p>Voce esta prestes a aprovar:</p>
    <p><strong>Nome:</strong> %s</p>
    <p><strong>Email:</strong> %s</p>
    <p><strong>Status atual:</strong> %s</p>
    <p class="muted">Confirme que este e-mail ja foi adicionado no Play Console.</p>
    <form method="post" action="/admin/beta-testers/%s/approve/confirm?token=%s">
      <button type="submit">Confirmar aprovacao</button>
    </form>
  </main>
</body>
</html>`,
		html.EscapeString(preview.Name),
		html.EscapeString(preview.Email),
		html.EscapeString(preview.Status),
		url.PathEscape(preview.TesterID),
		url.QueryEscape(token),
	)
}

func approvalSuccessHTML(preview usecase.BetaTesterApprovalPreview, status string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="pt-BR"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Aprovacao concluida</title></head>
<body style="font-family:Arial,sans-serif;padding:32px;background:#f7f8f4;color:#1c2a22;">
<main style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #dfe8d9;border-radius:8px;padding:24px;">
<h1>Aprovacao concluida</h1>
<p><strong>Email:</strong> %s</p>
<p><strong>Status:</strong> %s</p>
<p>A aprovacao foi salva. O backend processou o envio do e-mail de acesso ao beta para o usuario.</p>
</main></body></html>`, html.EscapeString(preview.Email), html.EscapeString(status))
}

func approvalErrorHTML(err error) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="pt-BR"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Link invalido</title></head>
<body style="font-family:Arial,sans-serif;padding:32px;background:#f7f8f4;color:#1c2a22;">
<main style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #dfe8d9;border-radius:8px;padding:24px;">
<h1>Nao foi possivel confirmar</h1>
<p>%s</p>
</main></body></html>`, html.EscapeString(err.Error()))
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
