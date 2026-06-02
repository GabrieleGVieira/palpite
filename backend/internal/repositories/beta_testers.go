package repositories

import (
	"context"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
)

type BetaTesterAndroidRepository struct {
	db Querier
}

func NewBetaTesterAndroidRepository(db Querier) BetaTesterAndroidRepository {
	return BetaTesterAndroidRepository{db: db}
}

func (repo BetaTesterAndroidRepository) UpsertLandingSignup(ctx context.Context, name string, email string) (domain.BetaTesterAndroid, error) {
	var tester domain.BetaTesterAndroid
	err := repo.db.QueryRow(ctx, `
		insert into beta_testers_android (name, email, source, status, error_message, updated_at)
		values (nullif($1, ''), $2, 'landing', 'pending', null, now())
		on conflict (email) do update
		set name = coalesce(nullif(excluded.name, ''), beta_testers_android.name),
			source = 'landing',
			status = case
				when beta_testers_android.status = 'added_to_google_group' then beta_testers_android.status
				else 'pending'
			end,
			error_message = case
				when beta_testers_android.status = 'added_to_google_group' then beta_testers_android.error_message
				else null
			end,
			updated_at = now()
		returning id::text, coalesce(name, ''), email, source, status,
			coalesce(error_message, ''), created_at, updated_at
	`, name, email).Scan(
		&tester.ID,
		&tester.Name,
		&tester.Email,
		&tester.Source,
		&tester.Status,
		&tester.ErrorMessage,
		&tester.CreatedAt,
		&tester.UpdatedAt,
	)
	return tester, err
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
