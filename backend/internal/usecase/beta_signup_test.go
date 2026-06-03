package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	emailservice "github.com/gabrielevieira/palpitai/backend/internal/email"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestBetaAndroidSignupPersistsPendingApprovalAndSendsNotification(t *testing.T) {
	now := time.Now()
	db := &betaSignupFakeDB{
		row: betaSignupFakeRow{values: betaSignupRowValues(now, "Gabriele", "user@example.com", domain.BetaTesterStatusPendingApproval)},
	}
	groupAdder := &countingGroupAdder{}
	emailSender := &fakeEmailSender{}

	result, err := NewBetaAndroidUsecase(db, groupAdder, emailSender, "admin@example.com", "https://api.example.com", "approval-secret", "https://play.example/beta", nil).Signup(context.Background(), BetaAndroidSignupInput{
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
	if len(emailSender.messages) != 1 {
		t.Fatalf("expected one notification email, got %d", len(emailSender.messages))
	}
	message := emailSender.messages[0]
	if message.Subject != "[Palpite!] Novo interessado no beta Android" {
		t.Fatalf("unexpected notification subject: %q", message.Subject)
	}
	if len(message.To) != 1 || message.To[0] != "admin@example.com" {
		t.Fatalf("unexpected notification recipients: %#v", message.To)
	}
	if !strings.Contains(message.Text, "Email: user@example.com") ||
		!strings.Contains(message.Text, "Status: pending_approval") ||
		!strings.Contains(message.HTML, "<strong>Email:</strong> user@example.com") ||
		!strings.Contains(message.HTML, "Confirmar aprovacao") ||
		!strings.Contains(message.HTML, "/admin/beta-testers/tester-id/approve/confirm?token=") {
		t.Fatalf("unexpected notification body: %#v", message)
	}
}

func TestBetaAndroidSignupSucceedsWhenNotificationEmailFails(t *testing.T) {
	now := time.Now()
	db := &betaSignupFakeDB{
		row: betaSignupFakeRow{values: betaSignupRowValues(now, "Gabriele", "user@example.com", domain.BetaTesterStatusPendingApproval)},
	}
	emailSender := &fakeEmailSender{err: errors.New("email provider unavailable")}

	result, err := NewBetaAndroidUsecase(db, nil, emailSender, "admin@example.com", "", "", "", nil).Signup(context.Background(), BetaAndroidSignupInput{
		Name:    "Gabriele",
		Email:   "user@example.com",
		Consent: true,
	})
	if err != nil {
		t.Fatalf("expected signup to succeed when notification fails, got %v", err)
	}
	if result.Status != domain.BetaTesterStatusPendingApproval {
		t.Fatalf("expected status %q, got %q", domain.BetaTesterStatusPendingApproval, result.Status)
	}
	if len(emailSender.messages) != 1 {
		t.Fatalf("expected notification attempt, got %d", len(emailSender.messages))
	}
}

func TestApproveBetaTesterApprovesAndSendsApprovalEmail(t *testing.T) {
	now := time.Now()
	approvedAt := now.Add(time.Minute)
	db := &betaSignupFakeDB{
		rows: []betaSignupFakeRow{
			{values: betaSignupRowValues(now, "Gabriele", "user@example.com", domain.BetaTesterStatusPendingApproval)},
			{values: betaSignupApprovedRowValues(now, approvedAt, "Gabriele", "user@example.com", "admin@example.com")},
		},
	}
	emailSender := &fakeEmailSender{}

	result, err := NewBetaAndroidUsecase(db, nil, emailSender, "", "", "", "https://play.google.com/apps/testing/com.gabrielevieira.palpite", nil).ApproveBetaTester(context.Background(), ApproveBetaTesterInput{
		TesterID:   "tester-id",
		ApprovedBy: "admin@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != domain.BetaTesterStatusApproved {
		t.Fatalf("expected status %q, got %q", domain.BetaTesterStatusApproved, result.Status)
	}
	if len(db.queryArgsHistory) != 2 {
		t.Fatalf("expected find and approve queries, got %d", len(db.queryArgsHistory))
	}
	if len(db.queryArgsHistory[1]) != 3 || db.queryArgsHistory[1][0] != "tester-id" || db.queryArgsHistory[1][2] != "admin@example.com" {
		t.Fatalf("unexpected approve query args: %#v", db.queryArgsHistory[1])
	}
	if len(emailSender.messages) != 1 {
		t.Fatalf("expected one approval email, got %d", len(emailSender.messages))
	}
	message := emailSender.messages[0]
	if message.Subject != "Seu acesso ao beta Android do Palpite! foi liberado" {
		t.Fatalf("unexpected approval subject: %q", message.Subject)
	}
	if len(message.To) != 1 || message.To[0] != "user@example.com" {
		t.Fatalf("unexpected approval recipients: %#v", message.To)
	}
	if !strings.Contains(message.Text, "https://play.google.com/apps/testing/com.gabrielevieira.palpite") ||
		!strings.Contains(message.HTML, "Participar do programa de testes") {
		t.Fatalf("unexpected approval email body: %#v", message)
	}
}

func TestApproveBetaTesterSucceedsWhenApprovalEmailFails(t *testing.T) {
	now := time.Now()
	db := &betaSignupFakeDB{
		rows: []betaSignupFakeRow{
			{values: betaSignupRowValues(now, "Gabriele", "user@example.com", domain.BetaTesterStatusPendingApproval)},
			{values: betaSignupApprovedRowValues(now, now.Add(time.Minute), "Gabriele", "user@example.com", "admin@example.com")},
		},
	}
	emailSender := &fakeEmailSender{err: errors.New("email provider unavailable")}

	result, err := NewBetaAndroidUsecase(db, nil, emailSender, "", "", "", "https://play.google.com/apps/testing/com.gabrielevieira.palpite", nil).ApproveBetaTester(context.Background(), ApproveBetaTesterInput{
		TesterID:   "tester-id",
		ApprovedBy: "admin@example.com",
	})
	if err != nil {
		t.Fatalf("expected approval to succeed when email fails, got %v", err)
	}
	if result.Status != domain.BetaTesterStatusApproved {
		t.Fatalf("expected status %q, got %q", domain.BetaTesterStatusApproved, result.Status)
	}
	if len(emailSender.messages) != 1 {
		t.Fatalf("expected approval email attempt, got %d", len(emailSender.messages))
	}
}

func TestConfirmBetaTesterApprovalWithSignedToken(t *testing.T) {
	now := time.Now()
	db := &betaSignupFakeDB{
		rows: []betaSignupFakeRow{
			{values: betaSignupRowValues(now, "Joao", "joao@gmail.com", domain.BetaTesterStatusPendingApproval)},
			{values: betaSignupApprovedRowValues(now, now.Add(time.Minute), "Joao", "joao@gmail.com", "signed_approval_link")},
		},
	}
	emailSender := &fakeEmailSender{}
	uc := NewBetaAndroidUsecase(db, nil, emailSender, "", "https://api.example.com", "approval-secret", "https://play.google.com/apps/testing/com.gabrielevieira.palpite", nil)
	token := uc.approvalToken("tester-id", time.Now().Add(time.Hour).Unix())

	result, err := uc.ConfirmBetaTesterApproval(context.Background(), ConfirmBetaTesterApprovalInput{
		TesterID: "tester-id",
		Token:    token,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.BetaTesterStatusApproved {
		t.Fatalf("expected status %q, got %q", domain.BetaTesterStatusApproved, result.Status)
	}
	if len(emailSender.messages) != 1 || emailSender.messages[0].To[0] != "joao@gmail.com" {
		t.Fatalf("expected approval email to tester, got %#v", emailSender.messages)
	}
}

func TestBetaAndroidSignupAllowsMissingGoogleGroupAdapter(t *testing.T) {
	now := time.Now()
	db := &betaSignupFakeDB{
		row: betaSignupFakeRow{values: betaSignupRowValues(now, "", "user@example.com", domain.BetaTesterStatusPendingApproval)},
	}

	_, err := NewBetaAndroidUsecase(db, nil, nil, "", "", "", "", nil).Signup(context.Background(), BetaAndroidSignupInput{
		Email:   "user@example.com",
		Consent: true,
	})
	if err != nil {
		t.Fatalf("expected signup to succeed without google group adapter, got %v", err)
	}
}

func TestBetaAndroidSignupRequiresConsent(t *testing.T) {
	db := &betaSignupFakeDB{}

	_, err := NewBetaAndroidUsecase(db, nil, nil, "", "", "", "", nil).Signup(context.Background(), BetaAndroidSignupInput{
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

	_, err := NewBetaAndroidUsecase(db, nil, nil, "", "", "", "", nil).Signup(context.Background(), BetaAndroidSignupInput{
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

type fakeEmailSender struct {
	messages []emailservice.Message
	err      error
}

func (sender *fakeEmailSender) Send(_ context.Context, message emailservice.Message) error {
	sender.messages = append(sender.messages, message)
	return sender.err
}

func betaSignupRowValues(now time.Time, name string, email string, status string) []any {
	return []any{
		"tester-id",
		name,
		email,
		domain.BetaTesterSourceLanding,
		domain.BetaTesterPlatformAndroid,
		status,
		"",
		sql.NullTime{},
		sql.NullString{},
		now,
		now,
	}
}

func betaSignupApprovedRowValues(createdAt time.Time, approvedAt time.Time, name string, email string, approvedBy string) []any {
	return []any{
		"tester-id",
		name,
		email,
		domain.BetaTesterSourceLanding,
		domain.BetaTesterPlatformAndroid,
		domain.BetaTesterStatusApproved,
		"",
		sql.NullTime{Time: approvedAt, Valid: true},
		sql.NullString{String: approvedBy, Valid: true},
		createdAt,
		approvedAt,
	}
}

type betaSignupFakeDB struct {
	row              betaSignupFakeRow
	rows             []betaSignupFakeRow
	queryArgs        []any
	queryArgsHistory [][]any
	queryCalls       int
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
	db.queryArgsHistory = append(db.queryArgsHistory, args)
	if len(db.rows) > 0 {
		row := db.rows[0]
		db.rows = db.rows[1:]
		return row
	}
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
		case *sql.NullTime:
			*pointer = value.(sql.NullTime)
		case *sql.NullString:
			*pointer = value.(sql.NullString)
		default:
			return errors.New("unexpected scan destination type")
		}
	}
	return nil
}
