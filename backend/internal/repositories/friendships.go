package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
)

type FriendshipRepository struct {
	db Querier
}

func NewFriendshipRepository(db Querier) FriendshipRepository {
	return FriendshipRepository{db: db}
}

func (repo FriendshipRepository) CreateRequest(ctx context.Context, requesterUserID string, addresseeUserID string) (domain.Friendship, error) {
	var friendship domain.Friendship
	err := repo.db.QueryRow(ctx, `
		insert into friendships (requester_user_id, addressee_user_id, status)
		values ($1, $2, 'PENDING')
		on conflict (
			(least(requester_user_id, addressee_user_id)),
			(greatest(requester_user_id, addressee_user_id))
		)
		do update set
			requester_user_id = excluded.requester_user_id,
			addressee_user_id = excluded.addressee_user_id,
			status = 'PENDING',
			updated_at = now()
		where friendships.status = 'DECLINED'
		returning id::text, requester_user_id::text, addressee_user_id::text, status, created_at, updated_at
	`, requesterUserID, addresseeUserID).Scan(
		&friendship.ID,
		&friendship.RequesterUserID,
		&friendship.AddresseeUserID,
		&friendship.Status,
		&friendship.CreatedAt,
		&friendship.UpdatedAt,
	)
	return friendship, err
}

func (repo FriendshipRepository) GetFriendship(ctx context.Context, userID string, otherUserID string) (domain.Friendship, error) {
	var friendship domain.Friendship
	err := repo.db.QueryRow(ctx, `
		select id::text, requester_user_id::text, addressee_user_id::text, status, created_at, updated_at
		from friendships
		where least(requester_user_id, addressee_user_id) = least($1::uuid, $2::uuid)
			and greatest(requester_user_id, addressee_user_id) = greatest($1::uuid, $2::uuid)
	`, userID, otherUserID).Scan(
		&friendship.ID,
		&friendship.RequesterUserID,
		&friendship.AddresseeUserID,
		&friendship.Status,
		&friendship.CreatedAt,
		&friendship.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Friendship{}, ErrNotFound
	}
	return friendship, err
}

func (repo FriendshipRepository) GetByID(ctx context.Context, friendshipID string) (domain.Friendship, error) {
	var friendship domain.Friendship
	err := repo.db.QueryRow(ctx, `
		select id::text, requester_user_id::text, addressee_user_id::text, status, created_at, updated_at
		from friendships
		where id = $1
	`, friendshipID).Scan(
		&friendship.ID,
		&friendship.RequesterUserID,
		&friendship.AddresseeUserID,
		&friendship.Status,
		&friendship.CreatedAt,
		&friendship.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Friendship{}, ErrNotFound
	}
	return friendship, err
}

func (repo FriendshipRepository) Accept(ctx context.Context, friendshipID string) (domain.Friendship, error) {
	return repo.updateStatus(ctx, friendshipID, domain.FriendshipStatusAccepted)
}

func (repo FriendshipRepository) Decline(ctx context.Context, friendshipID string) (domain.Friendship, error) {
	return repo.updateStatus(ctx, friendshipID, domain.FriendshipStatusDeclined)
}

func (repo FriendshipRepository) Block(ctx context.Context, friendshipID string) (domain.Friendship, error) {
	return repo.updateStatus(ctx, friendshipID, domain.FriendshipStatusBlocked)
}

func (repo FriendshipRepository) updateStatus(ctx context.Context, friendshipID string, status domain.FriendshipStatus) (domain.Friendship, error) {
	var friendship domain.Friendship
	err := repo.db.QueryRow(ctx, `
		update friendships
		set status = $2, updated_at = now()
		where id = $1
		returning id::text, requester_user_id::text, addressee_user_id::text, status, created_at, updated_at
	`, friendshipID, status).Scan(
		&friendship.ID,
		&friendship.RequesterUserID,
		&friendship.AddresseeUserID,
		&friendship.Status,
		&friendship.CreatedAt,
		&friendship.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Friendship{}, ErrNotFound
	}
	return friendship, err
}

func (repo FriendshipRepository) Delete(ctx context.Context, friendshipID string) error {
	commandTag, err := repo.db.Exec(ctx, `delete from friendships where id = $1`, friendshipID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (repo FriendshipRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := repo.db.QueryRow(ctx, `
		select exists (
			select 1
			from group_members
			where user_id = $1
				and status = 'active'
		)
	`, userID).Scan(&exists)
	return exists, err
}

func (repo FriendshipRepository) ListFriends(ctx context.Context, userID string) ([]dto.FriendResponse, error) {
	rows, err := repo.db.Query(ctx, `
		with friend_rows as (
			select
				f.id,
				case when f.requester_user_id = $1 then f.addressee_user_id else f.requester_user_id end as friend_user_id,
				f.created_at
			from friendships f
			where f.status = 'ACCEPTED'
				and (f.requester_user_id = $1 or f.addressee_user_id = $1)
		),
		profiles as (
			select distinct on (gm.user_id)
				gm.user_id,
				gm.display_name,
				gm.avatar_url
			from group_members gm
			where gm.status = 'active'
			order by gm.user_id, gm.joined_at desc
		)
		select fr.id::text, fr.friend_user_id::text, coalesce(p.display_name, ''), p.avatar_url, fr.created_at
		from friend_rows fr
		left join profiles p on p.user_id = fr.friend_user_id
		order by coalesce(p.display_name, '') asc, fr.created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	friends := []dto.FriendResponse{}
	for rows.Next() {
		var friend dto.FriendResponse
		if err := rows.Scan(&friend.ID, &friend.UserID, &friend.Name, &friend.AvatarURL, &friend.CreatedAt); err != nil {
			return nil, err
		}
		if strings.TrimSpace(friend.Name) == "" {
			friend.Name = "Palpiteiro"
		}
		friends = append(friends, friend)
	}
	return friends, rows.Err()
}

func (repo FriendshipRepository) ListPendingRequests(ctx context.Context, userID string) ([]dto.PendingFriendRequestResponse, error) {
	rows, err := repo.db.Query(ctx, `
		with profiles as (
			select distinct on (gm.user_id)
				gm.user_id,
				gm.display_name,
				gm.avatar_url
			from group_members gm
			where gm.status = 'active'
			order by gm.user_id, gm.joined_at desc
		)
		select f.id::text, f.requester_user_id::text, coalesce(p.display_name, ''), p.avatar_url, f.created_at
		from friendships f
		left join profiles p on p.user_id = f.requester_user_id
		where f.addressee_user_id = $1
			and f.status = 'PENDING'
		order by f.created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []dto.PendingFriendRequestResponse{}
	for rows.Next() {
		var request dto.PendingFriendRequestResponse
		if err := rows.Scan(&request.FriendshipID, &request.RequesterID, &request.Name, &request.AvatarURL, &request.CreatedAt); err != nil {
			return nil, err
		}
		if strings.TrimSpace(request.Name) == "" {
			request.Name = "Palpiteiro"
		}
		requests = append(requests, request)
	}
	return requests, rows.Err()
}

func (repo FriendshipRepository) SearchUsers(ctx context.Context, requesterUserID string, query string, limit int) ([]dto.UserSearchResponse, error) {
	rows, err := repo.db.Query(ctx, `
		with profiles as (
			select distinct on (gm.user_id)
				gm.user_id,
				gm.display_name,
				gm.avatar_url
			from group_members gm
			left join user_social_settings uss on uss.user_id = gm.user_id
			where gm.status = 'active'
				and gm.user_id <> $1
				and coalesce(uss.is_public_profile, true) = true
				and ($2 = '' or gm.display_name ilike '%' || $2 || '%')
			order by gm.user_id, gm.joined_at desc
		)
		select
			p.user_id::text,
			coalesce(p.display_name, ''),
			p.avatar_url,
			f.status
		from profiles p
		left join friendships f
			on least(f.requester_user_id, f.addressee_user_id) = least($1::uuid, p.user_id)
			and greatest(f.requester_user_id, f.addressee_user_id) = greatest($1::uuid, p.user_id)
			and f.status in ('PENDING', 'ACCEPTED', 'BLOCKED')
		order by coalesce(p.display_name, '') asc
		limit $3
	`, requesterUserID, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []dto.UserSearchResponse{}
	for rows.Next() {
		var user dto.UserSearchResponse
		if err := rows.Scan(&user.ID, &user.Name, &user.AvatarURL, &user.FriendshipStatus); err != nil {
			return nil, err
		}
		if strings.TrimSpace(user.Name) == "" {
			user.Name = "Palpiteiro"
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (repo FriendshipRepository) PublicProfile(ctx context.Context, userID string) (dto.PublicProfileResponse, error) {
	var profile dto.PublicProfileResponse
	err := repo.db.QueryRow(ctx, `
		with profile as (
			select distinct on (gm.user_id)
				gm.user_id,
				gm.display_name,
				gm.avatar_url,
				gm.joined_at
			from group_members gm
			where gm.user_id = $1
				and gm.status = 'active'
			order by gm.user_id, gm.joined_at asc
		),
		stats as (
			select
				count(distinct gm.group_id)::int as groups_count,
				count(distinct p.match_id)::int as predictions_count,
				coalesce(sum(p.points), 0)::int as total_points
			from group_members gm
			left join predictions p on p.group_id = gm.group_id
				and p.user_id = gm.user_id
			where gm.user_id = $1
				and gm.status = 'active'
		),
		ranking as (
			select user_id, rank() over (order by total_points desc, display_name asc)::int as position
			from (
				select
					gm.user_id,
					max(gm.display_name) as display_name,
					coalesce(sum(p.points), 0)::int as total_points
				from group_members gm
				left join predictions p on p.group_id = gm.group_id
					and p.user_id = gm.user_id
				where gm.status = 'active'
				group by gm.user_id
			) scores
		)
		select
			p.user_id::text,
			coalesce(p.display_name, ''),
			p.avatar_url,
			p.joined_at,
			s.groups_count,
			s.predictions_count,
			s.total_points,
			r.position
		from profile p
		cross join stats s
		left join ranking r on r.user_id = p.user_id
	`, userID).Scan(
		&profile.UserID,
		&profile.Name,
		&profile.AvatarURL,
		&profile.JoinedAt,
		&profile.GroupsCount,
		&profile.PredictionsCount,
		&profile.TotalPoints,
		&profile.GlobalRanking,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.PublicProfileResponse{}, ErrNotFound
	}
	if strings.TrimSpace(profile.Name) == "" {
		profile.Name = "Palpiteiro"
	}
	return profile, err
}
