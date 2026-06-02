package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
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

type ApproveBetaTesterInput struct {
	TesterID   string
	ApprovedBy string
}

type ApproveBetaTesterResult struct {
	Status string
}

type ConfirmBetaTesterApprovalInput struct {
	TesterID   string
	Token      string
	ApprovedBy string
}

type BetaTesterApprovalPreview struct {
	TesterID string
	Name     string
	Email    string
	Status   string
}

type BetaAndroidUsecase struct {
	repo              repositories.BetaTesterAndroidRepository
	emailSender       emailservice.Sender
	notificationEmail string
	approvalBaseURL   string
	approvalSecret    string
	playStoreURL      string
	logger            *slog.Logger
}

func NewBetaAndroidUsecase(db Datastore, _ GoogleGroupMemberAdder, emailSender emailservice.Sender, notificationEmail string, approvalBaseURL string, approvalSecret string, playStoreURL string, logger *slog.Logger) BetaAndroidUsecase {
	if logger == nil {
		logger = slog.Default()
	}

	return BetaAndroidUsecase{
		repo:              repositories.NewBetaTesterAndroidRepository(db),
		emailSender:       emailSender,
		notificationEmail: strings.TrimSpace(notificationEmail),
		approvalBaseURL:   strings.TrimRight(strings.TrimSpace(approvalBaseURL), "/"),
		approvalSecret:    strings.TrimSpace(approvalSecret),
		playStoreURL:      strings.TrimSpace(playStoreURL),
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
		RedirectURL: uc.playStoreURL,
		Status:      tester.Status,
	}, nil
}

func (uc BetaAndroidUsecase) ApproveBetaTester(ctx context.Context, input ApproveBetaTesterInput) (ApproveBetaTesterResult, error) {
	testerID := strings.TrimSpace(input.TesterID)
	if testerID == "" {
		return ApproveBetaTesterResult{}, apperrors.NewValidation("Informe o beta tester.")
	}

	approvedBy := strings.TrimSpace(input.ApprovedBy)
	if approvedBy == "" {
		return ApproveBetaTesterResult{}, apperrors.NewValidation("Informe o responsavel pela aprovacao.")
	}

	existing, err := uc.repo.FindByID(ctx, testerID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ApproveBetaTesterResult{}, apperrors.NewNotFound("Beta tester nao encontrado.")
		}
		return ApproveBetaTesterResult{}, err
	}

	tester, err := uc.repo.Approve(ctx, existing.ID, approvedBy)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ApproveBetaTesterResult{}, apperrors.NewNotFound("Beta tester nao encontrado.")
		}
		return ApproveBetaTesterResult{}, err
	}

	uc.logger.Info("beta tester approved", "tester_id", tester.ID, "email", tester.Email)
	uc.sendApprovalEmail(ctx, tester)

	return ApproveBetaTesterResult{Status: tester.Status}, nil
}

func (uc BetaAndroidUsecase) PreviewBetaTesterApproval(ctx context.Context, testerID string, token string) (BetaTesterApprovalPreview, error) {
	testerID = strings.TrimSpace(testerID)
	if testerID == "" {
		return BetaTesterApprovalPreview{}, apperrors.NewValidation("Informe o beta tester.")
	}
	if err := uc.validateApprovalToken(testerID, token); err != nil {
		return BetaTesterApprovalPreview{}, err
	}

	tester, err := uc.repo.FindByID(ctx, testerID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return BetaTesterApprovalPreview{}, apperrors.NewNotFound("Beta tester nao encontrado.")
		}
		return BetaTesterApprovalPreview{}, err
	}

	return BetaTesterApprovalPreview{
		TesterID: tester.ID,
		Name:     tester.Name,
		Email:    tester.Email,
		Status:   tester.Status,
	}, nil
}

func (uc BetaAndroidUsecase) ConfirmBetaTesterApproval(ctx context.Context, input ConfirmBetaTesterApprovalInput) (ApproveBetaTesterResult, error) {
	testerID := strings.TrimSpace(input.TesterID)
	if testerID == "" {
		return ApproveBetaTesterResult{}, apperrors.NewValidation("Informe o beta tester.")
	}
	if err := uc.validateApprovalToken(testerID, input.Token); err != nil {
		return ApproveBetaTesterResult{}, err
	}

	approvedBy := strings.TrimSpace(input.ApprovedBy)
	if approvedBy == "" {
		approvedBy = "signed_approval_link"
	}

	return uc.ApproveBetaTester(ctx, ApproveBetaTesterInput{
		TesterID:   testerID,
		ApprovedBy: approvedBy,
	})
}

func (uc BetaAndroidUsecase) sendSignupNotification(ctx context.Context, tester domain.BetaTesterAndroid) {
	if uc.emailSender == nil || uc.notificationEmail == "" {
		return
	}

	approvalURL := uc.approvalConfirmationURL(tester.ID)
	message := betaSignupNotificationMessage(uc.notificationEmail, tester, approvalURL)
	if err := uc.emailSender.Send(ctx, message); err != nil {
		uc.logger.Warn("beta android notification email failed", "email", tester.Email, "error", err)
		return
	}

	uc.logger.Info("beta android notification email sent", "email", tester.Email)
}

func (uc BetaAndroidUsecase) sendApprovalEmail(ctx context.Context, tester domain.BetaTesterAndroid) {
	if uc.emailSender == nil {
		return
	}
	if uc.playStoreURL == "" {
		uc.logger.Warn("beta android play store url not configured", "tester_id", tester.ID, "email", tester.Email)
		return
	}

	message := betaApprovalEmailMessage(tester.Email, uc.playStoreURL)
	if err := uc.emailSender.Send(ctx, message); err != nil {
		uc.logger.Warn("beta approval email failed", "tester_id", tester.ID, "email", tester.Email, "error", err)
		return
	}

	uc.logger.Info("beta approval email sent", "tester_id", tester.ID, "email", tester.Email)
}

func betaSignupNotificationMessage(to string, tester domain.BetaTesterAndroid, approvalURL string) emailservice.Message {
	name := strings.TrimSpace(tester.Name)
	if name == "" {
		name = "Nao informado"
	}

	status := strings.TrimSpace(tester.Status)
	if status == "" {
		status = domain.BetaTesterStatusPendingApproval
	}

	createdAt := formatBetaSignupNotificationTime(tester.CreatedAt)
	textApprovalBlock := ""
	htmlApprovalBlock := ""
	if approvalURL != "" {
		textApprovalBlock = fmt.Sprintf(`
Voce esta prestes a aprovar:
Nome: %s
Email: %s

Confirme que este e-mail ja foi adicionado no Play Console:
%s
`, name, tester.Email, approvalURL)
		htmlApprovalBlock = fmt.Sprintf(`
<hr>

<p><strong>Voce esta prestes a aprovar:</strong></p>
<p><strong>Nome:</strong> %s</p>
<p><strong>Email:</strong> %s</p>

<p>Confirme que este e-mail ja foi adicionado no Play Console.</p>

<p>
  <a href="%s" style="display:inline-block;padding:12px 18px;background:#1f7a4a;color:#ffffff;text-decoration:none;border-radius:8px;font-weight:700;">
    Confirmar aprovacao
  </a>
</p>`,
			html.EscapeString(name),
			html.EscapeString(tester.Email),
			html.EscapeString(approvalURL),
		)
	}

	return emailservice.Message{
		To:      []string{to},
		Subject: "[Palpite!] Novo interessado no beta Android",
		Text: fmt.Sprintf(`Novo interessado cadastrado no beta Android do Palpite!

Nome: %s
Email: %s
Status: %s
Data: %s

Acesse o painel administrativo ou o banco de dados para acompanhar os cadastros.
%s`, name, tester.Email, status, createdAt, textApprovalBlock),
		HTML: fmt.Sprintf(`<h2>Novo interessado no beta Android do Palpite!</h2>

<p><strong>Nome:</strong> %s</p>
<p><strong>Email:</strong> %s</p>
<p><strong>Status:</strong> %s</p>
<p><strong>Data:</strong> %s</p>

<hr>

<p>Acesse o painel administrativo ou o banco de dados para acompanhar os cadastros.</p>%s`,
			html.EscapeString(name),
			html.EscapeString(tester.Email),
			html.EscapeString(status),
			html.EscapeString(createdAt),
			htmlApprovalBlock,
		),
	}
}

func formatBetaSignupNotificationTime(value time.Time) string {
	if value.IsZero() {
		return time.Now().UTC().Format(time.RFC3339)
	}

	return value.UTC().Format(time.RFC3339)
}

func betaApprovalEmailMessage(to string, playStoreURL string) emailservice.Message {
	escapedURL := html.EscapeString(playStoreURL)
	return emailservice.Message{
		To:      []string{to},
		Subject: "Seu acesso ao beta Android do Palpite! foi liberado",
		Text: fmt.Sprintf(`Olá!

Seu acesso à versão beta Android do Palpite! foi aprovado.

Clique no link abaixo para participar do programa de testes:

%s

Após aceitar o convite, você poderá instalar o aplicativo pela Play Store.

Obrigado por ajudar a testar o Palpite!

Equipe Palpite!
`, playStoreURL),
		HTML: fmt.Sprintf(`<h2>Seu acesso ao beta Android do Palpite! foi liberado</h2>

<p>Olá!</p>

<p>Seu acesso à versão beta Android do Palpite! foi aprovado.</p>

<p>
  <a href="%s">
    Participar do programa de testes
  </a>
</p>

<p>
Após aceitar o convite, você poderá instalar o aplicativo pela Play Store.
</p>

<p>Obrigado por ajudar a testar o Palpite!</p>

<p><strong>Equipe Palpite!</strong></p>`, escapedURL),
	}
}

func (uc BetaAndroidUsecase) approvalConfirmationURL(testerID string) string {
	if uc.approvalBaseURL == "" || uc.approvalSecret == "" || strings.TrimSpace(testerID) == "" {
		return ""
	}

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour).Unix()
	token := uc.approvalToken(testerID, expiresAt)
	return fmt.Sprintf("%s/admin/beta-testers/%s/approve/confirm?token=%s",
		uc.approvalBaseURL,
		url.PathEscape(testerID),
		url.QueryEscape(token),
	)
}

func (uc BetaAndroidUsecase) approvalToken(testerID string, expiresAt int64) string {
	payload := fmt.Sprintf("%s.%d", testerID, expiresAt)
	signature := uc.signApprovalPayload(payload)
	return payload + "." + signature
}

func (uc BetaAndroidUsecase) validateApprovalToken(testerID string, token string) error {
	token = strings.TrimSpace(token)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return apperrors.NewForbidden("Link de aprovacao invalido.")
	}

	tokenTesterID := parts[0]
	if tokenTesterID != testerID {
		return apperrors.NewForbidden("Link de aprovacao invalido.")
	}

	expiresAt, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return apperrors.NewForbidden("Link de aprovacao invalido.")
	}
	if time.Now().UTC().Unix() > expiresAt {
		return apperrors.NewForbidden("Link de aprovacao expirado.")
	}
	if uc.approvalSecret == "" {
		return apperrors.NewForbidden("Aprovacao por link nao configurada.")
	}

	payload := parts[0] + "." + parts[1]
	expected := uc.signApprovalPayload(payload)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return apperrors.NewForbidden("Link de aprovacao invalido.")
	}

	return nil
}

func (uc BetaAndroidUsecase) signApprovalPayload(payload string) string {
	mac := hmac.New(sha256.New, []byte(uc.approvalSecret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
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
