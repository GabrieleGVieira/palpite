package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

type BetaTesterAndroidRepository struct {
	db Querier
}

func NewBetaTesterAndroidRepository(db Querier) BetaTesterAndroidRepository {
	return BetaTesterAndroidRepository{db: db}
}

func (repo BetaTesterAndroidRepository) UpsertLandingSignup(ctx context.Context, name string, email string) (domain.BetaTesterAndroid, error) {
	return repo.scanOne(ctx, `
		insert into beta_testers_android (name, email, source, platform, status, error_message, updated_at)
		values (nullif($1, ''), $2, $3, $4, $5, null, now())
		on conflict (email) do update
		set name = coalesce(nullif(excluded.name, ''), beta_testers_android.name),
			source = excluded.source,
			platform = excluded.platform,
			status = case
				when beta_testers_android.status in ('added_to_google_group', 'approved', 'exported') then beta_testers_android.status
				else excluded.status
			end,
			error_message = case
				when beta_testers_android.status in ('added_to_google_group', 'approved', 'exported') then beta_testers_android.error_message
				else null
			end,
			updated_at = now()
		returning id::text, coalesce(name, ''), email, source, platform, status,
			coalesce(error_message, ''), approved_at, approved_by, created_at, updated_at
	`, name, email, domain.BetaTesterSourceLanding, domain.BetaTesterPlatformAndroid, domain.BetaTesterStatusPendingApproval)
}

func (repo BetaTesterAndroidRepository) FindByID(ctx context.Context, testerID string) (domain.BetaTesterAndroid, error) {
	return repo.scanOne(ctx, `
		select id::text, coalesce(name, ''), email, source, platform, status,
			coalesce(error_message, ''), approved_at, approved_by, created_at, updated_at
		from beta_testers_android
		where id = $1
	`, testerID)
}

func (repo BetaTesterAndroidRepository) Approve(ctx context.Context, testerID string, approvedBy string) (domain.BetaTesterAndroid, error) {
	return repo.scanOne(ctx, `
		update beta_testers_android
		set status = $2,
			approved_at = now(),
			approved_by = nullif($3, ''),
			error_message = null,
			updated_at = now()
		where id = $1
		returning id::text, coalesce(name, ''), email, source, platform, status,
			coalesce(error_message, ''), approved_at, approved_by, created_at, updated_at
	`, testerID, domain.BetaTesterStatusApproved, approvedBy)
}

func (repo BetaTesterAndroidRepository) ListByStatus(ctx context.Context, status string) ([]domain.BetaTesterAndroid, error) {
	query := `
		select id::text, coalesce(name, ''), email, source, platform, status,
			coalesce(error_message, ''), approved_at, approved_by, created_at, updated_at
		from beta_testers_android
		where $1 = '' or status = $1
		order by created_at desc
	`
	rows, err := repo.db.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	testers := make([]domain.BetaTesterAndroid, 0)
	for rows.Next() {
		tester, err := scanBetaTester(rows.Scan)
		if err != nil {
			return nil, err
		}
		testers = append(testers, tester)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return testers, nil
}

func (repo BetaTesterAndroidRepository) MarkStatus(ctx context.Context, email string, status string, errorMessage string) error {
	_, err := repo.db.Exec(ctx, `
		update beta_testers_android
		set status = $2,
			error_message = nullif($3, ''),
			updated_at = now()
		where email = $1
	`, email, status, errorMessage)
	return err
}

func (repo BetaTesterAndroidRepository) scanOne(ctx context.Context, query string, args ...any) (domain.BetaTesterAndroid, error) {
	tester, err := scanBetaTester(func(dest ...any) error {
		return repo.db.QueryRow(ctx, query, args...).Scan(dest...)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.BetaTesterAndroid{}, ErrNotFound
	}

	return tester, err
}

type scannerFunc func(dest ...any) error

func scanBetaTester(scan scannerFunc) (domain.BetaTesterAndroid, error) {
	var tester domain.BetaTesterAndroid
	var approvedAt sql.NullTime
	var approvedBy sql.NullString

	err := scan(
		&tester.ID,
		&tester.Name,
		&tester.Email,
		&tester.Source,
		&tester.Platform,
		&tester.Status,
		&tester.ErrorMessage,
		&approvedAt,
		&approvedBy,
		&tester.CreatedAt,
		&tester.UpdatedAt,
	)
	if err != nil {
		return domain.BetaTesterAndroid{}, err
	}

	if approvedAt.Valid {
		tester.ApprovedAt = &approvedAt.Time
	}
	if approvedBy.Valid {
		tester.ApprovedBy = approvedBy.String
	}

	return tester, nil
}
