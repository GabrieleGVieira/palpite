package email

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
)

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type Message struct {
	To      []string
	Subject string
	Text    string
	HTML    string
	ReplyTo string
}

type NoopSender struct {
	logger *slog.Logger
}

func NewSender(cfg config.EmailConfig, env string, logger *slog.Logger) (Sender, error) {
	if logger == nil {
		logger = slog.Default()
	}

	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch provider {
	case "":
		logger.Info("email sending disabled; EMAIL_PROVIDER not configured", "env", env)
		return NoopSender{logger: logger}, nil
	case "brevo":
		brevo, err := NewBrevoSMTPSender(cfg, env, logger)
		if err != nil {
			return nil, err
		}
		return LoggingSender{sender: brevo, provider: "brevo", logger: logger}, nil
	default:
		return nil, errors.New("unsupported email provider: " + provider)
	}
}

func (sender NoopSender) Send(_ context.Context, message Message) error {
	logger := sender.logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("email sending disabled; skipping email", "to_count", len(message.To), "subject", message.Subject)
	return nil
}

type LoggingSender struct {
	sender   Sender
	provider string
	logger   *slog.Logger
}

func (sender LoggingSender) Send(ctx context.Context, message Message) error {
	logger := sender.logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("email send started", "provider", sender.provider, "to_count", len(message.To), "subject", message.Subject)
	if err := sender.sender.Send(ctx, message); err != nil {
		logger.Error("email send failed", "provider", sender.provider, "to_count", len(message.To), "subject", message.Subject, "error", err)
		return err
	}

	logger.Info("email send success", "provider", sender.provider, "to_count", len(message.To), "subject", message.Subject)
	return nil
}
