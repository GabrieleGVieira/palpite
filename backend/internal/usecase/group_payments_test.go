package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestAdminCanListPayments(t *testing.T) {
	now := time.Now()
	db := &paymentFakeDB{
		rows: []paymentFakeRow{{values: []any{true}}},
		queryRows: []paymentFakeRows{{
			values: [][]any{paymentRowValues(now, "member-id", "pending", 20.0, 0.0)},
		}},
	}

	payments, err := ListPayments(context.Background(), db, "owner-id", "group-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payments) != 1 {
		t.Fatalf("expected one payment, got %d", len(payments))
	}
	if payments[0].UserID != "member-id" || payments[0].Status != "pending" {
		t.Fatalf("unexpected payment: %#v", payments[0])
	}
}

func TestCommonUserCannotListPayments(t *testing.T) {
	db := &paymentFakeDB{
		rows: []paymentFakeRow{{values: []any{false}}},
	}

	_, err := ListPayments(context.Background(), db, "member-id", "group-id")
	if !apperrors.IsNotFound(err) {
		t.Fatalf("expected not found/forbidden-style error, got %v", err)
	}
}

func TestAdminCanMarkPaymentAsPaid(t *testing.T) {
	now := time.Now()
	db := &paymentFakeDB{
		rows: []paymentFakeRow{
			{values: []any{true}},
			{values: []any{true}},
			{values: paymentRowValues(now, "member-id", "paid", 20.0, 20.0)},
		},
	}

	payment, err := UpdatePayment(context.Background(), db, "owner-id", "group-id", "member-id", dto.UpdateGroupPaymentRequest{
		AmountPaid:    20,
		PaymentMethod: "Pix",
		Status:        "paid",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.Status != "paid" || payment.AmountPaid != 20 {
		t.Fatalf("unexpected payment: %#v", payment)
	}
}

func TestCommonUserCannotUpdatePayment(t *testing.T) {
	db := &paymentFakeDB{
		rows: []paymentFakeRow{{values: []any{false}}},
	}

	_, err := UpdatePayment(context.Background(), db, "member-id", "group-id", "member-id", dto.UpdateGroupPaymentRequest{Status: "paid"})
	if !apperrors.IsNotFound(err) {
		t.Fatalf("expected not found/forbidden-style error, got %v", err)
	}
}

func TestCannotUpdatePaymentForUserOutsideGroup(t *testing.T) {
	db := &paymentFakeDB{
		rows: []paymentFakeRow{
			{values: []any{true}},
			{values: []any{false}},
		},
	}

	_, err := UpdatePayment(context.Background(), db, "owner-id", "group-id", "outsider-id", dto.UpdateGroupPaymentRequest{Status: "paid"})
	if !errors.Is(err, ErrPaymentNotFound) {
		t.Fatalf("expected payment not found, got %v", err)
	}
}

func TestPaymentsSummaryCalculatesCountsAndAmounts(t *testing.T) {
	payments := []dto.GroupPaymentResponse{
		{Status: "paid", AmountExpected: 20, AmountPaid: 20},
		{Status: "pending", AmountExpected: 20, AmountPaid: 5},
		{Status: "exempt", AmountExpected: 0, AmountPaid: 0},
		{Status: "refunded", AmountExpected: 20, AmountPaid: 20},
	}

	summary := CalculatePaymentsSummary(payments)
	if summary.TotalParticipants != 4 ||
		summary.PaidCount != 1 ||
		summary.PendingCount != 1 ||
		summary.ExemptCount != 1 ||
		summary.RefundedCount != 1 ||
		summary.TotalExpected != 60 ||
		summary.TotalPaid != 45 ||
		summary.TotalPending != 15 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}

func paymentRowValues(now time.Time, userID string, status string, amountExpected float64, amountPaid float64) []any {
	adminID := "owner-id"
	return []any{
		"payment-id",
		"group-id",
		userID,
		"Participante",
		status,
		amountExpected,
		amountPaid,
		"Pix",
		now,
		adminID,
		"",
		now,
		now,
	}
}

type paymentFakeDB struct {
	rows      []paymentFakeRow
	queryRows []paymentFakeRows
}

func (db *paymentFakeDB) Ping(context.Context) error {
	return nil
}

func (db *paymentFakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (db *paymentFakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	if len(db.queryRows) == 0 {
		return nil, errors.New("unexpected query")
	}

	rows := db.queryRows[0]
	db.queryRows = db.queryRows[1:]
	return &rows, nil
}

func (db *paymentFakeDB) QueryRow(context.Context, string, ...any) pgx.Row {
	if len(db.rows) == 0 {
		return paymentFakeRow{err: errors.New("unexpected query row")}
	}

	row := db.rows[0]
	db.rows = db.rows[1:]
	return row
}

type paymentFakeRow struct {
	values []any
	err    error
}

func (row paymentFakeRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}

	assignScanValues(dest, row.values)
	return nil
}

type paymentFakeRows struct {
	index  int
	values [][]any
}

func (rows *paymentFakeRows) Close() {}

func (rows *paymentFakeRows) Err() error {
	return nil
}

func (rows *paymentFakeRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (rows *paymentFakeRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (rows *paymentFakeRows) Next() bool {
	if rows.index >= len(rows.values) {
		return false
	}
	rows.index++
	return true
}

func (rows *paymentFakeRows) Scan(dest ...any) error {
	assignScanValues(dest, rows.values[rows.index-1])
	return nil
}

func (rows *paymentFakeRows) Values() ([]any, error) {
	return rows.values[rows.index-1], nil
}

func (rows *paymentFakeRows) RawValues() [][]byte {
	return nil
}

func (rows *paymentFakeRows) Conn() *pgx.Conn {
	return nil
}

func assignScanValues(dest []any, values []any) {
	for index, value := range values {
		switch target := dest[index].(type) {
		case *bool:
			*target = value.(bool)
		case *string:
			*target = value.(string)
		case **string:
			if value == nil {
				*target = nil
			} else {
				nextValue := value.(string)
				*target = &nextValue
			}
		case *float64:
			*target = value.(float64)
		case *time.Time:
			*target = value.(time.Time)
		case **time.Time:
			if value == nil {
				*target = nil
			} else {
				nextValue := value.(time.Time)
				*target = &nextValue
			}
		default:
			panic("unsupported scan target")
		}
	}
}
