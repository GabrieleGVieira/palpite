package usecase

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	emailservice "github.com/gabrielevieira/palpitai/backend/internal/email"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

var (
	ErrBetaAndroidInvalidEmail    = errors.New("invalid email")
	ErrBetaAndroidConsentRequired = errors.New("consent required")
)

type GoogleGroupMemberAdder interface {
	AddMember(ctx context.Context, email string) error
}

type BetaAndroidSignupInput struct {
	Name    string
	Email   string
	Consent bool
}

type BetaAndroidSignupResult struct {
	RedirectURL string
	Status      string
}

type BetaAndroidUsecase struct {
	repo              repositories.BetaTesterAndroidRepository
	emailSender       emailservice.Sender
	notificationEmail string
	redirectURL       string
	logger            *slog.Logger
}

func NewBetaAndroidUsecase(db Datastore, _ GoogleGroupMemberAdder, emailSender emailservice.Sender, notificationEmail string, redirectURL string, logger *slog.Logger) BetaAndroidUsecase {
	if logger == nil {
		logger = slog.Default()
	}

	return BetaAndroidUsecase{
		repo:              repositories.NewBetaTesterAndroidRepository(db),
		emailSender:       emailSender,
		notificationEmail: strings.TrimSpace(notificationEmail),
		redirectURL:       strings.TrimSpace(redirectURL),
		logger:            logger,
	}
}

func (uc BetaAndroidUsecase) Signup(ctx context.Context, input BetaAndroidSignupInput) (BetaAndroidSignupResult, error) {
	if !input.Consent {
		return BetaAndroidSignupResult{}, ErrBetaAndroidConsentRequired
	}

	email, err := normalizeEmail(input.Email)
	if err != nil {
		return BetaAndroidSignupResult{}, err
	}

	name := strings.TrimSpace(input.Name)
	tester, err := uc.repo.UpsertLandingSignup(ctx, name, email)
	if err != nil {
		uc.logger.Error("beta android signup persistence failed", "email", email, "error", err)
		return BetaAndroidSignupResult{}, err
	}

	uc.logger.Info("beta android signup created", "email", email)
	uc.sendSignupNotification(ctx, tester)

	return BetaAndroidSignupResult{
		RedirectURL: uc.redirectURL,
		Status:      tester.Status,
	}, nil
}

func (uc BetaAndroidUsecase) sendSignupNotification(ctx context.Context, tester domain.BetaTesterAndroid) {
	if uc.emailSender == nil || uc.notificationEmail == "" {
		return
	}

	message := betaSignupNotificationMessage(uc.notificationEmail, tester)
	if err := uc.emailSender.Send(ctx, message); err != nil {
		uc.logger.Warn("beta android notification email failed", "email", tester.Email, "error", err)
		return
	}

	uc.logger.Info("beta android notification email sent", "email", tester.Email)
}

func betaSignupNotificationMessage(to string, tester domain.BetaTesterAndroid) emailservice.Message {
	name := strings.TrimSpace(tester.Name)
	if name == "" {
		name = "Nao informado"
	}

	status := strings.TrimSpace(tester.Status)
	if status == "" {
		status = domain.BetaTesterStatusPendingApproval
	}

	createdAt := formatBetaSignupNotificationTime(tester.CreatedAt)
	return emailservice.Message{
		To:      []string{to},
		Subject: "[Palpite!] Novo interessado no beta Android",
		Text: fmt.Sprintf(`Novo interessado cadastrado no beta Android do Palpite!

Nome: %s
Email: %s
Status: %s
Data: %s

Acesse o painel administrativo ou o banco de dados para acompanhar os cadastros.
`, name, tester.Email, status, createdAt),
		HTML: fmt.Sprintf(`<h2>Novo interessado no beta Android do Palpite!</h2>

<p><strong>Nome:</strong> %s</p>
<p><strong>Email:</strong> %s</p>
<p><strong>Status:</strong> %s</p>
<p><strong>Data:</strong> %s</p>

<hr>

<p>Acesse o painel administrativo ou o banco de dados para acompanhar os cadastros.</p>`,
			html.EscapeString(name),
			html.EscapeString(tester.Email),
			html.EscapeString(status),
			html.EscapeString(createdAt),
		),
	}
}

func formatBetaSignupNotificationTime(value time.Time) string {
	if value.IsZero() {
		return time.Now().UTC().Format(time.RFC3339)
	}

	return value.UTC().Format(time.RFC3339)
}

func normalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	if email == "" {
		return "", ErrBetaAndroidInvalidEmail
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || !strings.Contains(email, "@") {
		return "", ErrBetaAndroidInvalidEmail
	}

	return email, nil
}
