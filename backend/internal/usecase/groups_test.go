package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestNormalizeCreateGroupRequestRequiresSelectedTeams(t *testing.T) {
	limit := 10
	_, err := NormalizeCreateGroupRequest(dto.CreateGroupRequest{
		Name:             "Copa da firma",
		MatchScope:       "selected",
		ParticipantLimit: &limit,
	})
	if !apperrors.IsValidation(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestNormalizeCreateGroupRequestClearsTeamsForAllMatches(t *testing.T) {
	request, err := NormalizeCreateGroupRequest(dto.CreateGroupRequest{
		HasUnlimitedParticipants: true,
		MatchScope:               "all",
		Name:                     "Copa da firma",
		SelectedTeams:            []string{"Brasil"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(request.SelectedTeams) != 0 {
		t.Fatalf("expected selected teams to be cleared, got %v", request.SelectedTeams)
	}
	if request.ParticipantLimit != nil {
		t.Fatalf("expected participant limit to be nil")
	}
}

func TestNormalizeInviteCode(t *testing.T) {
	got := NormalizeInviteCode(" abcd-1234 ")
	if got != "ABCD1234" {
		t.Fatalf("expected ABCD1234, got %q", got)
	}
}

func TestListMemberSummariesRequiresExistingGroup(t *testing.T) {
	db := &groupSocialFakeDB{
		rows: []groupSocialFakeRow{
			{values: []any{false}},
		},
	}

	_, err := ListMemberSummaries(context.Background(), db, "user-id", "group-id")
	if !apperrors.IsNotFound(err) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestListMemberSummariesRequiresActiveMembership(t *testing.T) {
	db := &groupSocialFakeDB{
		rows: []groupSocialFakeRow{
			{values: []any{true}},
			{values: []any{false}},
		},
	}

	_, err := ListMemberSummaries(context.Background(), db, "user-id", "group-id")
	if !apperrors.IsForbidden(err) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestListMemberSummariesAllowsActiveMember(t *testing.T) {
	db := &groupSocialFakeDB{
		rows: []groupSocialFakeRow{
			{values: []any{true}},
			{values: []any{true}},
		},
		queryRows: groupSocialFakeRows{},
	}

	members, err := ListMemberSummaries(context.Background(), db, "user-id", "group-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 0 {
		t.Fatalf("expected empty members, got %d", len(members))
	}
}

func TestMemberDetailReturnsNotFoundForUserOutsideGroup(t *testing.T) {
	db := &groupSocialFakeDB{
		rows: []groupSocialFakeRow{
			{values: []any{true}},
			{values: []any{true}},
			{err: pgx.ErrNoRows},
		},
	}

	_, err := MemberDetail(context.Background(), db, "viewer-id", "group-id", "target-id")
	if !apperrors.IsNotFound(err) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

type groupSocialFakeDB struct {
	rows      []groupSocialFakeRow
	queryRows groupSocialFakeRows
}

func (db *groupSocialFakeDB) Ping(context.Context) error {
	return nil
}

func (db *groupSocialFakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected exec")
}

func (db *groupSocialFakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return db.queryRows, nil
}

func (db *groupSocialFakeDB) QueryRow(context.Context, string, ...any) pgx.Row {
	if len(db.rows) == 0 {
		return groupSocialFakeRow{err: errors.New("unexpected query row")}
	}

	row := db.rows[0]
	db.rows = db.rows[1:]
	return row
}

type groupSocialFakeRow struct {
	values []any
	err    error
}

func (row groupSocialFakeRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}

	for index, value := range row.values {
		switch target := dest[index].(type) {
		case *bool:
			*target = value.(bool)
		default:
			return errors.New("unsupported scan target")
		}
	}

	return nil
}

type groupSocialFakeRows struct{}

func (rows groupSocialFakeRows) Close() {}

func (rows groupSocialFakeRows) Err() error {
	return nil
}

func (rows groupSocialFakeRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (rows groupSocialFakeRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (rows groupSocialFakeRows) Next() bool {
	return false
}

func (rows groupSocialFakeRows) Scan(...any) error {
	return nil
}

func (rows groupSocialFakeRows) Values() ([]any, error) {
	return nil, nil
}

func (rows groupSocialFakeRows) RawValues() [][]byte {
	return nil
}

func (rows groupSocialFakeRows) Conn() *pgx.Conn {
	return nil
}
