package repositories

import (
	"context"
	"reflect"
	"testing"

	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestInsertGroupWithOwnerUsesPayloadFieldsInOrder(t *testing.T) {
	limit := 20
	db := &captureQuerier{}
	request := dto.CreateGroupRequest{
		Name:             "Copa da firma",
		Description:      "Bolao interno",
		MatchScope:       "selected",
		SelectedTeams:    []string{"Brasil", "Argentina"},
		ParticipantLimit: &limit,
		IsPrivate:        true,
	}

	_, err := InsertGroupWithOwner(
		context.Background(),
		db,
		"11111111-1111-1111-1111-111111111111",
		"Gabriele Vieira",
		request,
		"ABCD1234",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []any{
		"11111111-1111-1111-1111-111111111111",
		"Copa da firma",
		"Bolao interno",
		"selected",
		[]string{"Brasil", "Argentina"},
		&limit,
		true,
		"ABCD1234",
		"Gabriele Vieira",
	}

	if !reflect.DeepEqual(db.args, want) {
		t.Fatalf("unexpected query args:\nwant: %#v\n got: %#v", want, db.args)
	}
}

type captureQuerier struct {
	args []any
}

func (db *captureQuerier) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (db *captureQuerier) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (db *captureQuerier) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	db.args = args
	return noopRow{}
}

type noopRow struct{}

func (noopRow) Scan(...any) error {
	return nil
}
