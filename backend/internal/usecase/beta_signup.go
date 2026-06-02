package usecase

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"

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
	repo        repositories.BetaTesterAndroidRepository
	redirectURL string
	logger      *slog.Logger
}

func NewBetaAndroidUsecase(db Datastore, _ GoogleGroupMemberAdder, redirectURL string, logger *slog.Logger) BetaAndroidUsecase {
	if logger == nil {
		logger = slog.Default()
	}

	return BetaAndroidUsecase{
		repo:        repositories.NewBetaTesterAndroidRepository(db),
		redirectURL: strings.TrimSpace(redirectURL),
		logger:      logger,
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

	uc.logger.Info("beta android signup received", "email", email, "status", tester.Status)
	return BetaAndroidSignupResult{
		RedirectURL: uc.redirectURL,
		Status:      tester.Status,
	}, nil
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
