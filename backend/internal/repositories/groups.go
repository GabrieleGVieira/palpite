package repositories

import (
	"context"
	"errors"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
)

type GroupInviteSummary struct {
	ID               string
	IsPrivate        bool
	IsPaid           bool
	PaymentAmount    float64
	ParticipantLimit *int
	MemberCount      int
}

type GroupCapacity struct {
	ParticipantLimit *int
	MemberCount      int
}

func ListActiveUserGroups(ctx context.Context, db Querier, userID string) ([]dto.GroupListItemResponse, error) {
	rows, err := db.Query(ctx, `
		select
			g.id,
			case when owner_member.status = 'deleted' then '' else g.owner_id::text end as owner_id,
			g.name,
			g.description,
			g.match_scope,
			g.selected_teams,
			g.participant_limit,
			g.is_private,
			g.is_paid,
			g.payment_amount::float8,
			g.block_pending_predictions,
			g.invite_code,
			g.created_at,
			gm.role,
			gm.status,
			count(distinct all_members.user_id)::int as member_count,
			count(distinct pending_members.user_id)::int as pending_requests_count
		from groups g
		join group_members gm on gm.group_id = g.id and gm.user_id = $1 and gm.status = 'active'
		left join group_members owner_member on owner_member.group_id = g.id
			and owner_member.user_id = g.owner_id
		left join group_members all_members on all_members.group_id = g.id and all_members.status = 'active'
		left join group_members pending_members on pending_members.group_id = g.id and pending_members.status = 'pending'
		group by
			g.id,
			g.owner_id,
			g.name,
			g.description,
			g.match_scope,
			g.selected_teams,
			g.participant_limit,
			g.is_private,
			g.is_paid,
			g.payment_amount,
			g.block_pending_predictions,
			g.invite_code,
			g.created_at,
			gm.role,
			gm.status,
			owner_member.status
		order by g.created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []dto.GroupListItemResponse{}
	for rows.Next() {
		var group dto.GroupListItemResponse
		if err := rows.Scan(
			&group.ID,
			&group.OwnerID,
			&group.Name,
			&group.Description,
			&group.MatchScope,
			&group.SelectedTeams,
			&group.ParticipantLimit,
			&group.IsPrivate,
			&group.IsPaid,
			&group.PaymentAmount,
			&group.BlockPendingPredictions,
			&group.InviteCode,
			&group.CreatedAt,
			&group.Role,
			&group.Status,
			&group.MemberCount,
			&group.PendingRequestsCount,
		); err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	return groups, rows.Err()
}

func GroupInviteSummaryByCode(ctx context.Context, db Querier, inviteCode string) (GroupInviteSummary, error) {
	var group GroupInviteSummary
	err := db.QueryRow(ctx, `
		select
			g.id,
			g.is_private,
			g.is_paid,
			g.payment_amount::float8,
			g.participant_limit,
			count(gm.user_id)::int as member_count
		from groups g
		left join group_members gm on gm.group_id = g.id and gm.status = 'active'
		where g.invite_code = $1
		group by g.id, g.is_private, g.is_paid, g.payment_amount, g.participant_limit
	`, inviteCode).Scan(&group.ID, &group.IsPrivate, &group.IsPaid, &group.PaymentAmount, &group.ParticipantLimit, &group.MemberCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return GroupInviteSummary{}, ErrNotFound
	}

	return group, err
}

func GroupListItemByID(ctx context.Context, db Querier, groupID string) (dto.GroupListItemResponse, error) {
	var group dto.GroupListItemResponse
	err := db.QueryRow(ctx, `
		select
			g.id,
			case when owner_member.status = 'deleted' then '' else g.owner_id::text end as owner_id,
			g.name,
			g.description,
			g.match_scope,
			g.selected_teams,
			g.participant_limit,
			g.is_private,
			g.is_paid,
			g.payment_amount::float8,
			g.block_pending_predictions,
			g.invite_code,
			g.created_at,
			count(distinct all_members.user_id)::int as member_count,
			count(distinct pending_members.user_id)::int as pending_requests_count
		from groups g
		left join group_members owner_member on owner_member.group_id = g.id
			and owner_member.user_id = g.owner_id
		left join group_members all_members on all_members.group_id = g.id and all_members.status = 'active'
		left join group_members pending_members on pending_members.group_id = g.id and pending_members.status = 'pending'
		where g.id = $1
		group by
			g.id,
			g.owner_id,
			g.name,
			g.description,
			g.match_scope,
			g.selected_teams,
			g.participant_limit,
			g.is_private,
			g.is_paid,
			g.payment_amount,
			g.block_pending_predictions,
			g.invite_code,
			g.created_at,
			owner_member.status
	`, groupID).Scan(
		&group.ID,
		&group.OwnerID,
		&group.Name,
		&group.Description,
		&group.MatchScope,
		&group.SelectedTeams,
		&group.ParticipantLimit,
		&group.IsPrivate,
		&group.IsPaid,
		&group.PaymentAmount,
		&group.BlockPendingPredictions,
		&group.InviteCode,
		&group.CreatedAt,
		&group.MemberCount,
		&group.PendingRequestsCount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.GroupListItemResponse{}, ErrNotFound
	}

	return group, err
}

func OwnerGroupCapacity(ctx context.Context, db Querier, ownerID string, groupID string) (GroupCapacity, error) {
	var capacity GroupCapacity
	err := db.QueryRow(ctx, `
		with locked_group as (
			select id, participant_limit
			from groups
			where id = $1 and owner_id = $2
			for update
		)
		select
			locked_group.participant_limit,
			count(gm.user_id)::int as member_count
		from locked_group
		left join group_members gm on gm.group_id = locked_group.id and gm.status = 'active'
		group by locked_group.id, locked_group.participant_limit
	`, groupID, ownerID).Scan(&capacity.ParticipantLimit, &capacity.MemberCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return GroupCapacity{}, ErrNotFound
	}

	return capacity, err
}

func InsertGroupWithOwner(ctx context.Context, db Querier, userID string, displayName string, request dto.CreateGroupRequest, inviteCode string) (dto.GroupResponse, error) {
	var group dto.GroupResponse
	err := db.QueryRow(ctx, `
		with inserted_group as (
			insert into groups (
				owner_id,
				name,
				description,
				match_scope,
				selected_teams,
				participant_limit,
				is_private,
				is_paid,
				payment_amount,
				block_pending_predictions,
				invite_code
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			returning
				id,
				owner_id,
				name,
				description,
				match_scope,
				selected_teams,
				participant_limit,
				is_private,
				is_paid,
				payment_amount::float8,
				block_pending_predictions,
				invite_code,
				created_at
		),
		inserted_member as (
			insert into group_members (group_id, user_id, role, display_name)
			select id, owner_id, 'owner', $12 from inserted_group
		)
		select
			id,
			owner_id,
			name,
			description,
			match_scope,
			selected_teams,
			participant_limit,
			is_private,
			is_paid,
			payment_amount,
			block_pending_predictions,
			invite_code,
			created_at
		from inserted_group
	`,
		userID,
		request.Name,
		request.Description,
		request.MatchScope,
		request.SelectedTeams,
		request.ParticipantLimit,
		request.IsPrivate,
		request.IsPaid,
		request.PaymentAmount,
		request.BlockPendingPredictions,
		inviteCode,
		displayName,
	).Scan(
		&group.ID,
		&group.OwnerID,
		&group.Name,
		&group.Description,
		&group.MatchScope,
		&group.SelectedTeams,
		&group.ParticipantLimit,
		&group.IsPrivate,
		&group.IsPaid,
		&group.PaymentAmount,
		&group.BlockPendingPredictions,
		&group.InviteCode,
		&group.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.GroupResponse{}, ErrNotFound
	}

	return group, err
}

func UpdateOwnedGroup(ctx context.Context, db Querier, ownerID string, groupID string, request dto.UpdateGroupRequest) (dto.GroupResponse, error) {
	var group dto.GroupResponse
	err := db.QueryRow(ctx, `
		update groups
		set
			name = $3,
			description = $4,
			participant_limit = $5,
			is_private = $6,
			is_paid = $7,
			payment_amount = $8,
			block_pending_predictions = $9,
			updated_at = now()
		where id = $1 and owner_id = $2
		returning
			id,
			owner_id,
			name,
			description,
			match_scope,
			selected_teams,
			participant_limit,
			is_private,
			is_paid,
			payment_amount::float8,
			block_pending_predictions,
			invite_code,
			created_at
	`, groupID, ownerID, request.Name, request.Description, request.ParticipantLimit, request.IsPrivate, request.IsPaid, request.PaymentAmount, request.BlockPendingPredictions).Scan(
		&group.ID,
		&group.OwnerID,
		&group.Name,
		&group.Description,
		&group.MatchScope,
		&group.SelectedTeams,
		&group.ParticipantLimit,
		&group.IsPrivate,
		&group.IsPaid,
		&group.PaymentAmount,
		&group.BlockPendingPredictions,
		&group.InviteCode,
		&group.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.GroupResponse{}, ErrNotFound
	}

	return group, err
}

func GroupsAffectedByMatch(ctx context.Context, db Querier, matchID string) ([]domain.GroupSummary, error) {
	rows, err := db.Query(ctx, `
		select distinct g.id::text, g.name
		from groups g
		join predictions p on p.group_id = g.id
		where p.match_id = $1
		order by g.name asc
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []domain.GroupSummary{}
	for rows.Next() {
		var group domain.GroupSummary
		if err := rows.Scan(&group.ID, &group.Name); err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	return groups, rows.Err()
}
