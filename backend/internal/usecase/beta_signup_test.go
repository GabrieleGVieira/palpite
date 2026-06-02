package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestBetaAndroidSignupPersistsPendingApprovalWithoutAddingGoogleGroup(t *testing.T) {
	now := time.Now()
	db := &betaSignupFakeDB{
		row: betaSignupFakeRow{values: []any{
			"tester-id",
			"Gabriele",
			"user@example.com",
			domain.BetaTesterSourceLanding,
			domain.BetaTesterPlatformAndroid,
			domain.BetaTesterStatusPendingApproval,
			"",
			now,
			now,
		}},
	}
	groupAdder := &countingGroupAdder{}

	result, err := NewBetaAndroidUsecase(db, groupAdder, "https://play.example/beta", nil).Signup(context.Background(), BetaAndroidSignupInput{
		Name:    " Gabriele ",
		Email:   " USER@Example.COM ",
		Consent: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != domain.BetaTesterStatusPendingApproval {
		t.Fatalf("expected status %q, got %q", domain.BetaTesterStatusPendingApproval, result.Status)
	}
	if result.RedirectURL != "https://play.example/beta" {
		t.Fatalf("expected redirect url to be preserved for clients that need it, got %q", result.RedirectURL)
	}
	if groupAdder.calls != 0 {
		t.Fatalf("expected google group adder not to be called, got %d calls", groupAdder.calls)
	}
	if len(db.queryArgs) < 2 || db.queryArgs[0] != "Gabriele" || db.queryArgs[1] != "user@example.com" {
		t.Fatalf("expected trimmed name and normalized email in repository args, got %#v", db.queryArgs)
	}
}

func TestBetaAndroidSignupAllowsMissingGoogleGroupAdapter(t *testing.T) {
	now := time.Now()
	db := &betaSignupFakeDB{
		row: betaSignupFakeRow{values: []any{
			"tester-id",
			"",
			"user@example.com",
			domain.BetaTesterSourceLanding,
			domain.BetaTesterPlatformAndroid,
			domain.BetaTesterStatusPendingApproval,
			"",
			now,
			now,
		}},
	}

	_, err := NewBetaAndroidUsecase(db, nil, "", nil).Signup(context.Background(), BetaAndroidSignupInput{
		Email:   "user@example.com",
		Consent: true,
	})
	if err != nil {
		t.Fatalf("expected signup to succeed without google group adapter, got %v", err)
	}
}

func TestBetaAndroidSignupRequiresConsent(t *testing.T) {
	db := &betaSignupFakeDB{}

	_, err := NewBetaAndroidUsecase(db, nil, "", nil).Signup(context.Background(), BetaAndroidSignupInput{
		Email: "user@example.com",
	})
	if !errors.Is(err, ErrBetaAndroidConsentRequired) {
		t.Fatalf("expected consent error, got %v", err)
	}
	if db.queryCalls != 0 {
		t.Fatalf("expected no persistence without consent, got %d query calls", db.queryCalls)
	}
}

func TestBetaAndroidSignupRejectsInvalidEmail(t *testing.T) {
	db := &betaSignupFakeDB{}

	_, err := NewBetaAndroidUsecase(db, nil, "", nil).Signup(context.Background(), BetaAndroidSignupInput{
		Email:   "invalid-email",
		Consent: true,
	})
	if !errors.Is(err, ErrBetaAndroidInvalidEmail) {
		t.Fatalf("expected invalid email error, got %v", err)
	}
	if db.queryCalls != 0 {
		t.Fatalf("expected no persistence for invalid email, got %d query calls", db.queryCalls)
	}
}

type countingGroupAdder struct {
	calls int
}

func (adder *countingGroupAdder) AddMember(context.Context, string) error {
	adder.calls++
	return nil
}

type betaSignupFakeDB struct {
	row        betaSignupFakeRow
	queryArgs  []any
	queryCalls int
}

func (db *betaSignupFakeDB) Ping(context.Context) error {
	return nil
}

func (db *betaSignupFakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (db *betaSignupFakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected query")
}

func (db *betaSignupFakeDB) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	db.queryCalls++
	db.queryArgs = args
	return db.row
}

type betaSignupFakeRow struct {
	values []any
	err    error
}

func (row betaSignupFakeRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(dest) != len(row.values) {
		return errors.New("unexpected scan destination count")
	}
	for i, value := range row.values {
		switch pointer := dest[i].(type) {
		case *string:
			*pointer = value.(string)
		case *time.Time:
			*pointer = value.(time.Time)
		default:
			return errors.New("unexpected scan destination type")
		}
	}
	return nil
}
