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

type predictionFakeDB struct {
	rows []predictionFakeRow
}

func (db *predictionFakeDB) Ping(context.Context) error {
	return nil
}

func (db *predictionFakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected exec")
}

func (db *predictionFakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected query")
}

func (db *predictionFakeDB) QueryRow(context.Context, string, ...any) pgx.Row {
	if len(db.rows) == 0 {
		return predictionFakeRow{err: errors.New("unexpected query row")}
	}

	row := db.rows[0]
	db.rows = db.rows[1:]
	return row
}

type predictionFakeRow struct {
	values []any
	err    error
}

func (row predictionFakeRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}

	for index, value := range row.values {
		switch target := dest[index].(type) {
		case *bool:
			*target = value.(bool)
		case *time.Time:
			*target = value.(time.Time)
		default:
			return errors.New("unsupported scan target")
		}
	}

	return nil
}

func TestSavePredictionRequiresActiveMembership(t *testing.T) {
	db := &predictionFakeDB{
		rows: []predictionFakeRow{
			{values: []any{false}},
		},
	}

	_, err := SavePrediction(context.Background(), db, "user-id", "group-id", "match-id", dto.PredictionRequest{})
	if !apperrors.IsForbidden(err) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestSavePredictionRejectsMatchAlreadyStarted(t *testing.T) {
	db := &predictionFakeDB{
		rows: []predictionFakeRow{
			{values: []any{true}},
			{values: []any{true}},
			{values: []any{time.Now().Add(-time.Minute)}},
		},
	}

	_, err := SavePrediction(context.Background(), db, "user-id", "group-id", "match-id", dto.PredictionRequest{})
	if !apperrors.IsConflict(err) {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestSavePredictionRequiresConfirmedPaymentWhenGroupBlocksPendingPayments(t *testing.T) {
	db := &predictionFakeDB{
		rows: []predictionFakeRow{
			{values: []any{true}},
			{values: []any{false}},
		},
	}

	_, err := SavePrediction(context.Background(), db, "user-id", "group-id", "match-id", dto.PredictionRequest{})
	if !errors.Is(err, ErrPaymentRequired) {
		t.Fatalf("expected payment required error, got %v", err)
	}
}
