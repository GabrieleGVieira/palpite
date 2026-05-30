package repositories

import (
	"context"
	"errors"

	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
)

const DeletedUserDisplayName = "Usuário excluído"

func UserOwnedGroupCount(ctx context.Context, db Querier, userID string) (int, error) {
	var count int
	err := db.QueryRow(ctx, `
		select count(*)::int
		from groups
		where owner_id = $1
	`, userID).Scan(&count)

	return count, err
}

func AnonymizeAccountData(ctx context.Context, db Querier, userID string) error {
	if _, err := db.Exec(ctx, `
		update group_members
		set
			display_name = $2,
			status = 'deleted'
		where user_id = $1
	`, userID, DeletedUserDisplayName); err != nil {
		return err
	}

	_, err := db.Exec(ctx, `
		update groups
		set updated_at = now()
		where owner_id = $1
	`, userID)

	return err
}

func UserProfile(ctx context.Context, db Querier, userID string) (dto.ProfileResponse, error) {
	var profile dto.ProfileResponse
	err := db.QueryRow(ctx, `
		select display_name, avatar_url
		from group_members
		where user_id = $1
			and status = 'active'
		order by joined_at desc
		limit 1
	`, userID).Scan(&profile.DisplayName, &profile.AvatarURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.ProfileResponse{}, ErrNotFound
	}

	return profile, err
}

func UpdateUserProfile(ctx context.Context, db Querier, userID string, displayName string, avatarURL *string) (dto.ProfileResponse, error) {
	var profile dto.ProfileResponse
	err := db.QueryRow(ctx, `
		with updated_members as (
			update group_members
			set
				display_name = $2,
				avatar_url = $3
			where user_id = $1
				and status = 'active'
			returning display_name, avatar_url, joined_at
		)
		select display_name, avatar_url
		from updated_members
		order by joined_at desc
		limit 1
	`, userID, displayName, avatarURL).Scan(&profile.DisplayName, &profile.AvatarURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.ProfileResponse{}, ErrNotFound
	}

	return profile, err
}
