package email

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
)

func TestBuildSMTPMessageIncludesTextAndHTMLParts(t *testing.T) {
	payload, err := buildSMTPMessage(BrevoSMTPSender{
		fromName:    "Palpite!",
		fromAddress: "noreply@example.com",
	}.from(), Message{
		To:      []string{"admin@example.com"},
		Subject: "Subject",
		Text:    "Hello text",
		HTML:    "<p>Hello html</p>",
		ReplyTo: "reply@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := string(payload)
	for _, expected := range []string{
		`From: "Palpite!" <noreply@example.com>`,
		"To: admin@example.com",
		"Subject: Subject",
		"Reply-To: reply@example.com",
		"MIME-Version: 1.0",
		"multipart/alternative",
		"text/plain; charset=UTF-8",
		"text/html; charset=UTF-8",
		"Hello text",
		"<p>Hello html</p>",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected SMTP payload to contain %q, got:\n%s", expected, body)
		}
	}
}

func TestNewSenderSelectsBrevoProvider(t *testing.T) {
	sender, err := NewSender(config.EmailConfig{
		Provider:          "brevo",
		FromName:          "Palpite!",
		FromAddress:       "noreply@example.com",
		BrevoSMTPHost:     "smtp-relay.brevo.com",
		BrevoSMTPPort:     587,
		BrevoSMTPUser:     "smtp-user",
		BrevoSMTPPassword: "smtp-password",
	}, "production", slog.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logging, ok := sender.(LoggingSender)
	if !ok {
		t.Fatalf("expected LoggingSender, got %T", sender)
	}
	if logging.provider != "brevo" {
		t.Fatalf("expected brevo provider, got %q", logging.provider)
	}
	if _, ok := logging.sender.(BrevoSMTPSender); !ok {
		t.Fatalf("expected BrevoSMTPSender, got %T", logging.sender)
	}
}

func TestNoopSenderReturnsNil(t *testing.T) {
	err := NoopSender{logger: slog.Default()}.Send(context.Background(), Message{
		To:      []string{"admin@example.com"},
		Subject: "Subject",
		Text:    "Hello",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestNewSenderReturnsNoopWhenProviderIsEmpty(t *testing.T) {
	sender, err := NewSender(config.EmailConfig{}, "test", slog.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := sender.(NoopSender); !ok {
		t.Fatalf("expected NoopSender, got %T", sender)
	}
}

func TestNewSenderErrorsWhenBrevoConfigIsMissingInProduction(t *testing.T) {
	_, err := NewSender(config.EmailConfig{Provider: "brevo"}, "production", slog.Default())
	if err == nil {
		t.Fatal("expected missing Brevo config error")
	}
	for _, secret := range []string{"smtp-password", "BREVO_SMTP_PASSWORD="} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error leaked secret material: %q", err.Error())
		}
	}
}
