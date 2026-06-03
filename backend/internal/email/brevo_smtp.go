package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
)

const (
	defaultBrevoSMTPHost = "smtp-relay.brevo.com"
	defaultBrevoSMTPPort = 587
	smtpTimeout          = 10 * time.Second
)

type BrevoSMTPSender struct {
	host        string
	port        int
	username    string
	password    string
	fromName    string
	fromAddress string
}

func NewBrevoSMTPSender(cfg config.EmailConfig, env string, logger *slog.Logger) (Sender, error) {
	sender := BrevoSMTPSender{
		host:        firstNonEmpty(cfg.BrevoSMTPHost, defaultBrevoSMTPHost),
		port:        cfg.BrevoSMTPPort,
		username:    strings.TrimSpace(cfg.BrevoSMTPUser),
		password:    strings.TrimSpace(cfg.BrevoSMTPPassword),
		fromName:    strings.TrimSpace(cfg.FromName),
		fromAddress: strings.TrimSpace(cfg.FromAddress),
	}
	if sender.port == 0 {
		sender.port = defaultBrevoSMTPPort
	}

	if err := sender.validate(); err != nil {
		if isProduction(env) {
			return nil, err
		}
		if logger == nil {
			logger = slog.Default()
		}
		logger.Info("email sending disabled; Brevo SMTP config incomplete", "env", env)
		return NoopSender{logger: logger}, nil
	}

	return sender, nil
}

func (sender BrevoSMTPSender) Send(ctx context.Context, message Message) error {
	if err := sender.validate(); err != nil {
		return err
	}
	if len(message.To) == 0 {
		return errors.New("email recipients not configured")
	}

	payload, err := buildSMTPMessage(sender.from(), message)
	if err != nil {
		return err
	}

	addr := net.JoinHostPort(sender.host, strconv.Itoa(sender.port))
	auth := smtp.PlainAuth("", sender.username, sender.password, sender.host)
	return sendMailWithStartTLS(ctx, addr, sender.host, auth, sender.fromAddress, message.To, payload)
}

func (sender BrevoSMTPSender) validate() error {
	missing := make([]string, 0)
	if strings.TrimSpace(sender.host) == "" {
		missing = append(missing, "BREVO_SMTP_HOST")
	}
	if sender.port <= 0 {
		missing = append(missing, "BREVO_SMTP_PORT")
	}
	if strings.TrimSpace(sender.username) == "" {
		missing = append(missing, "BREVO_SMTP_USER")
	}
	if strings.TrimSpace(sender.password) == "" {
		missing = append(missing, "BREVO_SMTP_PASSWORD")
	}
	if strings.TrimSpace(sender.fromAddress) == "" {
		missing = append(missing, "EMAIL_FROM_ADDRESS")
	}
	if len(missing) > 0 {
		return errors.New("missing required Brevo email config: " + strings.Join(missing, ", "))
	}

	return nil
}

func (sender BrevoSMTPSender) from() mail.Address {
	return mail.Address{Name: sender.fromName, Address: sender.fromAddress}
}

func buildSMTPMessage(from mail.Address, message Message) ([]byte, error) {
	if len(message.To) == 0 {
		return nil, errors.New("email recipients not configured")
	}
	if strings.TrimSpace(message.Subject) == "" {
		return nil, errors.New("email subject not configured")
	}
	if strings.TrimSpace(message.Text) == "" && strings.TrimSpace(message.HTML) == "" {
		return nil, errors.New("email body not configured")
	}

	var buffer bytes.Buffer
	writeHeader(&buffer, "From", from.String())
	writeHeader(&buffer, "To", strings.Join(message.To, ", "))
	writeHeader(&buffer, "Subject", message.Subject)
	if strings.TrimSpace(message.ReplyTo) != "" {
		writeHeader(&buffer, "Reply-To", strings.TrimSpace(message.ReplyTo))
	}
	writeHeader(&buffer, "MIME-Version", "1.0")

	if strings.TrimSpace(message.HTML) == "" {
		writeHeader(&buffer, "Content-Type", "text/plain; charset=UTF-8")
		writeHeader(&buffer, "Content-Transfer-Encoding", "quoted-printable")
		buffer.WriteString("\r\n")
		writeQuotedPrintable(&buffer, message.Text)
		return buffer.Bytes(), nil
	}

	writer := multipart.NewWriter(&buffer)
	writeHeader(&buffer, "Content-Type", "multipart/alternative; boundary="+writer.Boundary())
	buffer.WriteString("\r\n")

	if strings.TrimSpace(message.Text) != "" {
		part, err := writer.CreatePart(map[string][]string{
			"Content-Type":              {"text/plain; charset=UTF-8"},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		if err != nil {
			return nil, err
		}
		writeQuotedPrintable(part, message.Text)
	}

	part, err := writer.CreatePart(map[string][]string{
		"Content-Type":              {"text/html; charset=UTF-8"},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	if err != nil {
		return nil, err
	}
	writeQuotedPrintable(part, message.HTML)
	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func writeHeader(buffer *bytes.Buffer, key string, value string) {
	buffer.WriteString(key)
	buffer.WriteString(": ")
	buffer.WriteString(value)
	buffer.WriteString("\r\n")
}

func writeQuotedPrintable(buffer interface{ Write([]byte) (int, error) }, value string) {
	writer := quotedprintable.NewWriter(buffer)
	_, _ = writer.Write([]byte(value))
	_ = writer.Close()
}

func sendMailWithStartTLS(ctx context.Context, addr string, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	dialer := net.Dialer{Timeout: smtpTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Quit()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return err
	}
	return writer.Close()
}

func firstNonEmpty(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}

	return value
}

func isProduction(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "production", "prod":
		return true
	default:
		return false
	}
}
