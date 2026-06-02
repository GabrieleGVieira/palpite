package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
)

const (
	resendAPIURL          = "https://api.resend.com/emails"
	resendDefaultFrom     = "Palpite! <noreply@palpite.app>"
	resendFallbackFrom    = "noreply@resend.dev"
	resendUserAgent       = "palpite-api/1.0"
	resendRequestTimeout  = 10 * time.Second
	resendMaxErrorBodyLen = 2048
)

type ResendSender struct {
	apiKey     string
	from       string
	httpClient *http.Client
	logger     *slog.Logger
	apiURL     string
}

func NewResendSender(cfg config.EmailConfig, logger *slog.Logger) Sender {
	if logger == nil {
		logger = slog.Default()
	}

	apiKey := strings.TrimSpace(cfg.ResendAPIKey)
	if apiKey == "" {
		logger.Warn("resend api key not configured; beta signup notification email disabled")
		return nil
	}

	return ResendSender{
		apiKey:     apiKey,
		from:       resendFrom(cfg.From),
		httpClient: &http.Client{Timeout: resendRequestTimeout},
		logger:     logger,
		apiURL:     resendAPIURL,
	}
}

func resendFrom(value string) string {
	from := strings.TrimSpace(value)
	if from == "" {
		return resendDefaultFrom
	}

	return from
}

func (sender ResendSender) Send(ctx context.Context, message Message) error {
	if strings.TrimSpace(sender.apiKey) == "" {
		return errors.New("resend api key not configured")
	}
	if len(message.To) == 0 {
		return errors.New("email recipients not configured")
	}

	statusCode, responseBody, err := sender.sendWithFrom(ctx, message, sender.from)
	if err != nil {
		return err
	}
	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		sender.log().Info("resend email accepted", "status", statusCode, "from", sender.from)
		return nil
	}

	if statusCode == http.StatusForbidden && strings.TrimSpace(sender.from) != resendFallbackFrom {
		sender.log().Warn("resend email forbidden with configured from, retrying fallback from", "from", sender.from, "fallback_from", resendFallbackFrom)
		statusCode, responseBody, err = sender.sendWithFrom(ctx, message, resendFallbackFrom)
		if err != nil {
			return err
		}
		if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
			sender.log().Info("resend email accepted", "status", statusCode, "from", resendFallbackFrom)
			return nil
		}
	}

	return fmt.Errorf("resend api returned %d: %s", statusCode, strings.TrimSpace(responseBody))
}

func (sender ResendSender) sendWithFrom(ctx context.Context, message Message, from string) (int, string, error) {
	payload, err := json.Marshal(resendEmailRequest{
		From:    from,
		To:      message.To,
		Subject: message.Subject,
		HTML:    message.HTML,
		Text:    message.Text,
	})
	if err != nil {
		return 0, "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.apiURL, bytes.NewReader(payload))
	if err != nil {
		return 0, "", err
	}
	request.Header.Set("Authorization", "Bearer "+sender.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", resendUserAgent)

	response, err := sender.client().Do(request)
	if err != nil {
		sender.log().Error("resend email request failed", "error", err, "from", from)
		return 0, "", err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(response.Body, resendMaxErrorBodyLen))
	return response.StatusCode, string(body), nil
}

func (sender ResendSender) log() *slog.Logger {
	if sender.logger == nil {
		return slog.Default()
	}

	return sender.logger
}

func (sender ResendSender) client() *http.Client {
	if sender.httpClient == nil {
		return &http.Client{Timeout: resendRequestTimeout}
	}

	return sender.httpClient
}

type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}
