package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
)

func CreateChallenge(ctx context.Context, db Querier, creatorUserID string, opponentUserID string, matchID string, stakeAmount int) (domain.PalpicoinChallenge, error) {
	var challenge domain.PalpicoinChallenge
	err := db.QueryRow(ctx, `
		insert into palpicoin_challenges (creator_user_id, opponent_user_id, match_id, stake_amount, status)
		values ($1, $2, $3, $4, 'PENDING')
		returning id::text, creator_user_id::text, opponent_user_id::text, match_id::text, stake_amount,
			creator_prediction_id::text, opponent_prediction_id::text, creator_points, opponent_points,
			winner_user_id::text, status, created_at, accepted_at, settled_at, updated_at
	`, creatorUserID, opponentUserID, matchID, stakeAmount).Scan(
		&challenge.ID,
		&challenge.CreatorUserID,
		&challenge.OpponentUserID,
		&challenge.MatchID,
		&challenge.StakeAmount,
		&challenge.CreatorPredictionID,
		&challenge.OpponentPredictionID,
		&challenge.CreatorPoints,
		&challenge.OpponentPoints,
		&challenge.WinnerUserID,
		&challenge.Status,
		&challenge.CreatedAt,
		&challenge.AcceptedAt,
		&challenge.SettledAt,
		&challenge.UpdatedAt,
	)
	return challenge, err
}

func GetChallenge(ctx context.Context, db Querier, challengeID string) (domain.PalpicoinChallenge, error) {
	var challenge domain.PalpicoinChallenge
	err := db.QueryRow(ctx, `
		select id::text, creator_user_id::text, opponent_user_id::text, match_id::text, stake_amount,
			creator_prediction_id::text, opponent_prediction_id::text, creator_points, opponent_points,
			winner_user_id::text, status, created_at, accepted_at, settled_at, updated_at
		from palpicoin_challenges
		where id = $1
	`, challengeID).Scan(
		&challenge.ID,
		&challenge.CreatorUserID,
		&challenge.OpponentUserID,
		&challenge.MatchID,
		&challenge.StakeAmount,
		&challenge.CreatorPredictionID,
		&challenge.OpponentPredictionID,
		&challenge.CreatorPoints,
		&challenge.OpponentPoints,
		&challenge.WinnerUserID,
		&challenge.Status,
		&challenge.CreatedAt,
		&challenge.AcceptedAt,
		&challenge.SettledAt,
		&challenge.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PalpicoinChallenge{}, ErrNotFound
	}
	return challenge, err
}

func UpdateChallengeStatus(ctx context.Context, db Querier, challengeID string, status domain.ChallengeStatus) (domain.PalpicoinChallenge, error) {
	var acceptedAt string
	if status == domain.ChallengeStatusAccepted {
		acceptedAt = "now()"
	} else {
		acceptedAt = "accepted_at"
	}
	query := strings.ReplaceAll(`
		update palpicoin_challenges
		set status = $2, accepted_at = ACCEPTED_AT_EXPR, updated_at = now()
		where id = $1
		returning id::text, creator_user_id::text, opponent_user_id::text, match_id::text, stake_amount,
			creator_prediction_id::text, opponent_prediction_id::text, creator_points, opponent_points,
			winner_user_id::text, status, created_at, accepted_at, settled_at, updated_at
	`, "ACCEPTED_AT_EXPR", acceptedAt)

	var challenge domain.PalpicoinChallenge
	err := db.QueryRow(ctx, query, challengeID, status).Scan(
		&challenge.ID,
		&challenge.CreatorUserID,
		&challenge.OpponentUserID,
		&challenge.MatchID,
		&challenge.StakeAmount,
		&challenge.CreatorPredictionID,
		&challenge.OpponentPredictionID,
		&challenge.CreatorPoints,
		&challenge.OpponentPoints,
		&challenge.WinnerUserID,
		&challenge.Status,
		&challenge.CreatedAt,
		&challenge.AcceptedAt,
		&challenge.SettledAt,
		&challenge.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PalpicoinChallenge{}, ErrNotFound
	}
	return challenge, err
}

func ListChallenges(ctx context.Context, db Querier, userID string, status string, challengeType string) ([]dto.ChallengeResponse, error) {
	whereType := "and ($3 = 'all' or ($3 = 'sent' and c.creator_user_id = $1) or ($3 = 'received' and c.opponent_user_id = $1))"
	rows, err := db.Query(ctx, `
		with profiles as (
			select distinct on (gm.user_id)
				gm.user_id,
				gm.display_name,
				gm.avatar_url
			from group_members gm
			where gm.status = 'active'
			order by gm.user_id, gm.joined_at desc
		)
		select
			c.id::text,
			c.creator_user_id::text,
			c.opponent_user_id::text,
			c.match_id::text,
			c.stake_amount,
			c.creator_prediction_id::text,
			c.opponent_prediction_id::text,
			c.creator_points,
			c.opponent_points,
			c.winner_user_id::text,
			c.status,
			c.created_at,
			c.accepted_at,
			c.settled_at,
			c.updated_at,
			m.home_team,
			m.away_team,
			m.kickoff_at,
			coalesce(nullif(p.display_name, ''), 'Palpiteiro'),
			p.avatar_url
		from palpicoin_challenges c
		join world_cup_matches m on m.id = c.match_id
		left join profiles p on p.user_id = case when c.creator_user_id = $1 then c.opponent_user_id else c.creator_user_id end
		where (c.creator_user_id = $1 or c.opponent_user_id = $1)
			and ($2 = '' or c.status = $2)
			`+whereType+`
		order by c.created_at desc
	`, userID, status, challengeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	challenges := []dto.ChallengeResponse{}
	for rows.Next() {
		var challenge dto.ChallengeResponse
		if err := rows.Scan(
			&challenge.ID,
			&challenge.CreatorUserID,
			&challenge.OpponentUserID,
			&challenge.MatchID,
			&challenge.StakeAmount,
			&challenge.CreatorPredictionID,
			&challenge.OpponentPredictionID,
			&challenge.CreatorPoints,
			&challenge.OpponentPoints,
			&challenge.WinnerUserID,
			&challenge.Status,
			&challenge.CreatedAt,
			&challenge.AcceptedAt,
			&challenge.SettledAt,
			&challenge.UpdatedAt,
			&challenge.HomeTeam,
			&challenge.AwayTeam,
			&challenge.KickoffAt,
			&challenge.FriendName,
			&challenge.FriendAvatarURL,
		); err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}
	return challenges, rows.Err()
}

func ChallengeDetail(ctx context.Context, db Querier, userID string, challengeID string) (dto.ChallengeResponse, error) {
	challenges, err := ListChallenges(ctx, db, userID, "", "all")
	if err != nil {
		return dto.ChallengeResponse{}, err
	}
	for _, challenge := range challenges {
		if challenge.ID == challengeID {
			return challenge, nil
		}
	}
	return dto.ChallengeResponse{}, ErrNotFound
}

func ListChallengesBetweenUsers(ctx context.Context, db Querier, userID string, otherUserID string) ([]dto.ChallengeResponse, error) {
	rows, err := db.Query(ctx, `
		with profiles as (
			select distinct on (gm.user_id)
				gm.user_id,
				gm.display_name,
				gm.avatar_url
			from group_members gm
			where gm.status = 'active'
			order by gm.user_id, gm.joined_at desc
		)
		select
			c.id::text,
			c.creator_user_id::text,
			c.opponent_user_id::text,
			c.match_id::text,
			c.stake_amount,
			c.creator_prediction_id::text,
			c.opponent_prediction_id::text,
			c.creator_points,
			c.opponent_points,
			c.winner_user_id::text,
			c.status,
			c.created_at,
			c.accepted_at,
			c.settled_at,
			c.updated_at,
			m.home_team,
			m.away_team,
			m.kickoff_at,
			coalesce(nullif(p.display_name, ''), 'Palpiteiro'),
			p.avatar_url
		from palpicoin_challenges c
		join world_cup_matches m on m.id = c.match_id
		left join profiles p on p.user_id = case when c.creator_user_id = $1 then c.opponent_user_id else c.creator_user_id end
		where least(c.creator_user_id, c.opponent_user_id) = least($1::uuid, $2::uuid)
			and greatest(c.creator_user_id, c.opponent_user_id) = greatest($1::uuid, $2::uuid)
		order by c.created_at desc
		limit 20
	`, userID, otherUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	challenges := []dto.ChallengeResponse{}
	for rows.Next() {
		var challenge dto.ChallengeResponse
		if err := rows.Scan(
			&challenge.ID,
			&challenge.CreatorUserID,
			&challenge.OpponentUserID,
			&challenge.MatchID,
			&challenge.StakeAmount,
			&challenge.CreatorPredictionID,
			&challenge.OpponentPredictionID,
			&challenge.CreatorPoints,
			&challenge.OpponentPoints,
			&challenge.WinnerUserID,
			&challenge.Status,
			&challenge.CreatedAt,
			&challenge.AcceptedAt,
			&challenge.SettledAt,
			&challenge.UpdatedAt,
			&challenge.HomeTeam,
			&challenge.AwayTeam,
			&challenge.KickoffAt,
			&challenge.FriendName,
			&challenge.FriendAvatarURL,
		); err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}
	return challenges, rows.Err()
}

func ListRefundableChallengesBetweenUsers(ctx context.Context, db Querier, userID string, otherUserID string) ([]domain.PalpicoinChallenge, error) {
	rows, err := db.Query(ctx, `
		select c.id::text, c.creator_user_id::text, c.opponent_user_id::text, c.match_id::text, c.stake_amount,
			c.creator_prediction_id::text, c.opponent_prediction_id::text, c.creator_points, c.opponent_points,
			c.winner_user_id::text, c.status, c.created_at, c.accepted_at, c.settled_at, c.updated_at
		from palpicoin_challenges c
		join world_cup_matches m on m.id = c.match_id
		where least(c.creator_user_id, c.opponent_user_id) = least($1::uuid, $2::uuid)
			and greatest(c.creator_user_id, c.opponent_user_id) = greatest($1::uuid, $2::uuid)
			and c.status in ('PENDING', 'ACCEPTED')
			and m.status = 'scheduled'
			and m.kickoff_at > now()
	`, userID, otherUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	challenges := []domain.PalpicoinChallenge{}
	for rows.Next() {
		var challenge domain.PalpicoinChallenge
		if err := rows.Scan(
			&challenge.ID,
			&challenge.CreatorUserID,
			&challenge.OpponentUserID,
			&challenge.MatchID,
			&challenge.StakeAmount,
			&challenge.CreatorPredictionID,
			&challenge.OpponentPredictionID,
			&challenge.CreatorPoints,
			&challenge.OpponentPoints,
			&challenge.WinnerUserID,
			&challenge.Status,
			&challenge.CreatedAt,
			&challenge.AcceptedAt,
			&challenge.SettledAt,
			&challenge.UpdatedAt,
		); err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}
	return challenges, rows.Err()
}

func DeleteChallengesBetweenUsers(ctx context.Context, db Querier, userID string, otherUserID string) error {
	_, err := db.Exec(ctx, `
		delete from palpicoin_challenges
		where least(creator_user_id, opponent_user_id) = least($1::uuid, $2::uuid)
			and greatest(creator_user_id, opponent_user_id) = greatest($1::uuid, $2::uuid)
	`, userID, otherUserID)
	return err
}

func SettleAcceptedChallengesForMatch(ctx context.Context, db Querier, matchID string) ([]domain.PalpicoinChallenge, error) {
	rows, err := db.Query(ctx, `
		with scored as (
			select
				c.id,
				coalesce(max(p.points) filter (where p.user_id = c.creator_user_id), 0)::int as creator_points,
				coalesce(max(p.points) filter (where p.user_id = c.opponent_user_id), 0)::int as opponent_points
			from palpicoin_challenges c
			left join predictions p on p.match_id = c.match_id
				and p.user_id in (c.creator_user_id, c.opponent_user_id)
			where c.match_id = $1
				and c.status = 'ACCEPTED'
			group by c.id
		)
		update palpicoin_challenges c
		set
			creator_points = s.creator_points,
			opponent_points = s.opponent_points,
			winner_user_id = case
				when s.creator_points > s.opponent_points then c.creator_user_id
				when s.opponent_points > s.creator_points then c.opponent_user_id
				else null
			end,
			status = 'SETTLED',
			settled_at = now(),
			updated_at = now()
		from scored s
		where c.id = s.id
		returning c.id::text, c.creator_user_id::text, c.opponent_user_id::text, c.match_id::text, c.stake_amount,
			c.creator_prediction_id::text, c.opponent_prediction_id::text, c.creator_points, c.opponent_points,
			c.winner_user_id::text, c.status, c.created_at, c.accepted_at, c.settled_at, c.updated_at
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	challenges := []domain.PalpicoinChallenge{}
	for rows.Next() {
		var challenge domain.PalpicoinChallenge
		if err := rows.Scan(
			&challenge.ID,
			&challenge.CreatorUserID,
			&challenge.OpponentUserID,
			&challenge.MatchID,
			&challenge.StakeAmount,
			&challenge.CreatorPredictionID,
			&challenge.OpponentPredictionID,
			&challenge.CreatorPoints,
			&challenge.OpponentPoints,
			&challenge.WinnerUserID,
			&challenge.Status,
			&challenge.CreatedAt,
			&challenge.AcceptedAt,
			&challenge.SettledAt,
			&challenge.UpdatedAt,
		); err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}
	return challenges, rows.Err()
}
