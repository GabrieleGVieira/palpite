package usecase

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

var (
	ErrBetaAndroidInvalidEmail     = errors.New("invalid email")
	ErrBetaAndroidConsentRequired  = errors.New("consent required")
	ErrBetaAndroidGroupAddFailed   = errors.New("google group add failed")
	ErrBetaAndroidRedirectNotReady = errors.New("play store beta url not configured")
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
	groupAdder  GoogleGroupMemberAdder
	redirectURL string
	logger      *slog.Logger
}

func NewBetaAndroidUsecase(db Datastore, groupAdder GoogleGroupMemberAdder, redirectURL string, logger *slog.Logger) BetaAndroidUsecase {
	if logger == nil {
		logger = slog.Default()
	}

	return BetaAndroidUsecase{
		repo:        repositories.NewBetaTesterAndroidRepository(db),
		groupAdder:  groupAdder,
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

	if tester.Status == domain.BetaTesterStatusAddedToGoogleGroup {
		if uc.redirectURL == "" {
			uc.logger.Error("beta android redirect url missing", "email", email)
			return BetaAndroidSignupResult{}, ErrBetaAndroidRedirectNotReady
		}
		uc.logger.Info("beta android signup already added", "email", email)
		return BetaAndroidSignupResult{RedirectURL: uc.redirectURL, Status: tester.Status}, nil
	}

	if uc.groupAdder == nil {
		err := errors.New("google group adapter not configured")
		_ = uc.repo.MarkStatus(ctx, email, domain.BetaTesterStatusFailed, err.Error())
		uc.logger.Error("beta android google group adapter missing", "email", email, "error", err)
		return BetaAndroidSignupResult{Status: domain.BetaTesterStatusFailed}, ErrBetaAndroidGroupAddFailed
	}

	if err := uc.groupAdder.AddMember(ctx, email); err != nil {
		_ = uc.repo.MarkStatus(ctx, email, domain.BetaTesterStatusFailed, err.Error())
		uc.logger.Error("beta android google group add failed", "email", email, "error", err)
		return BetaAndroidSignupResult{Status: domain.BetaTesterStatusFailed}, ErrBetaAndroidGroupAddFailed
	}

	if err := uc.repo.MarkStatus(ctx, email, domain.BetaTesterStatusAddedToGoogleGroup, ""); err != nil {
		uc.logger.Error("beta android status update failed", "email", email, "error", err)
		return BetaAndroidSignupResult{}, err
	}

	if uc.redirectURL == "" {
		uc.logger.Error("beta android redirect url missing", "email", email)
		return BetaAndroidSignupResult{}, ErrBetaAndroidRedirectNotReady
	}

	uc.logger.Info("beta android signup completed", "email", email)
	return BetaAndroidSignupResult{
		RedirectURL: uc.redirectURL,
		Status:      domain.BetaTesterStatusAddedToGoogleGroup,
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
