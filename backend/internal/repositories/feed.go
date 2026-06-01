package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"

	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
)

type FeedEventInput struct {
	ActorUserID *string
	EventType   string
	GroupID     string
	MatchID     *string
	Metadata    map[string]any
}

type RankingSnapshot struct {
	Position    int
	TotalPoints int
	UserID      string
	DisplayName string
}

func InsertFeedEvent(ctx context.Context, db Querier, input FeedEventInput) (string, bool, error) {
	metadata := input.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	payload, err := json.Marshal(metadata)
	if err != nil {
		return "", false, err
	}

	var id string
	err = db.QueryRow(ctx, `
		insert into group_feed_events (
			group_id,
			event_type,
			actor_user_id,
			match_id,
			metadata_json
		)
		select $1, $2, $3, $4, $5::jsonb
		where not exists (
			select 1
			from group_feed_events
			where group_id = $1
				and event_type = $2
				and actor_user_id is not distinct from $3::uuid
				and match_id is not distinct from $4::uuid
				and metadata_json = $5::jsonb
		)
		returning id
	`, input.GroupID, input.EventType, input.ActorUserID, input.MatchID, string(payload)).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	slog.Info("feed event created", "id", id, "group_id", input.GroupID, "event_type", input.EventType)

	return id, true, nil
}

func ListFeedEvents(ctx context.Context, db Querier, groupID string, userID string, page int, pageSize int) ([]dto.GroupFeedEventResponse, bool, error) {
	offset := (page - 1) * pageSize
	rows, err := db.Query(ctx, `
		select
			e.id,
			e.event_type,
			e.created_at,
			e.metadata_json::text,
			gm.user_id::text,
			gm.display_name,
			gm.avatar_url,
			coalesce(
				jsonb_agg(
					jsonb_build_object(
						'reactionType', reaction_counts.reaction_type,
						'count', reaction_counts.reaction_count,
						'reactedByMe', reaction_counts.reacted_by_me
					)
					order by reaction_counts.reaction_type
				) filter (where reaction_counts.reaction_type is not null),
				'[]'::jsonb
			)::text as reactions_json
		from group_feed_events e
		left join group_members gm on gm.group_id = e.group_id
			and gm.user_id = e.actor_user_id
		left join lateral (
			select
				r.reaction_type,
				count(*)::int as reaction_count,
				bool_or(r.user_id = $2)::bool as reacted_by_me
			from group_feed_event_reactions r
			where r.feed_event_id = e.id
			group by r.reaction_type
		) reaction_counts on true
		where e.group_id = $1
		group by
			e.id,
			e.event_type,
			e.created_at,
			e.metadata_json,
			gm.user_id,
			gm.display_name,
			gm.avatar_url
		order by e.created_at desc
		limit $3 offset $4
	`, groupID, userID, pageSize+1, offset)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	events := []dto.GroupFeedEventResponse{}
	for rows.Next() {
		var event dto.GroupFeedEventResponse
		var metadataJSON string
		var reactionsJSON string
		var actorID *string
		var actorName *string
		var actorAvatarURL *string
		if err := rows.Scan(
			&event.ID,
			&event.Type,
			&event.CreatedAt,
			&metadataJSON,
			&actorID,
			&actorName,
			&actorAvatarURL,
			&reactionsJSON,
		); err != nil {
			return nil, false, err
		}

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &event.Metadata); err != nil {
				return nil, false, err
			}
		}
		if reactionsJSON != "" {
			if err := json.Unmarshal([]byte(reactionsJSON), &event.Reactions); err != nil {
				return nil, false, err
			}
		}
		if event.Reactions == nil {
			event.Reactions = []dto.FeedReactionSummaryResponse{}
		}
		if actorID != nil {
			name := ""
			if actorName != nil {
				name = *actorName
			}
			event.Actor = &dto.GroupFeedActorResponse{
				AvatarURL: actorAvatarURL,
				ID:        *actorID,
				Name:      name,
			}
		}

		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(events) > pageSize
	if hasMore {
		events = events[:pageSize]
	}

	return events, hasMore, nil
}

func UpsertFeedReaction(ctx context.Context, db Querier, groupID string, eventID string, userID string, reactionType string) error {
	_, err := db.Exec(ctx, `
		insert into group_feed_event_reactions (
			feed_event_id,
			group_id,
			user_id,
			reaction_type
		)
		select e.id, e.group_id, $3, $4
		from group_feed_events e
		where e.id = $2
			and e.group_id = $1
		on conflict (feed_event_id, user_id, reaction_type)
		do update set
			reaction_type = excluded.reaction_type,
			updated_at = now()
	`, groupID, eventID, userID, reactionType)

	return err
}

func DeleteFeedReaction(ctx context.Context, db Querier, groupID string, eventID string, userID string, reactionType string) error {
	_, err := db.Exec(ctx, `
		delete from group_feed_event_reactions
		where group_id = $1
			and feed_event_id = $2
			and user_id = $3
			and ($4 = '' or reaction_type = $4)
	`, groupID, eventID, userID, reactionType)

	return err
}

func FeedEventExists(ctx context.Context, db Querier, groupID string, eventID string) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `
		select exists (
			select 1
			from group_feed_events
			where group_id = $1
				and id = $2
		)
	`, groupID, eventID).Scan(&exists)

	return exists, err
}

func GroupRankingSnapshot(ctx context.Context, db Querier, groupID string) ([]RankingSnapshot, error) {
	rows, err := db.Query(ctx, `
		with scores as (
			select
				gm.user_id::text as user_id,
				gm.display_name,
				coalesce(sum(p.points), 0)::int as total_points
			from group_members gm
			left join predictions p on p.group_id = gm.group_id
				and p.user_id = gm.user_id
			where gm.group_id = $1
				and gm.status = 'active'
			group by gm.user_id, gm.display_name
		)
		select
			rank() over (order by total_points desc, display_name asc)::int as position,
			user_id,
			display_name,
			total_points
		from scores
		order by position asc, display_name asc
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ranking := []RankingSnapshot{}
	for rows.Next() {
		var entry RankingSnapshot
		if err := rows.Scan(&entry.Position, &entry.UserID, &entry.DisplayName, &entry.TotalPoints); err != nil {
			return nil, err
		}
		ranking = append(ranking, entry)
	}

	return ranking, rows.Err()
}

func ExactScorePredictionsForMatch(ctx context.Context, db Querier, matchID string, homeScore int, awayScore int) ([]FeedEventInput, error) {
	rows, err := db.Query(ctx, `
		select
			p.group_id::text,
			p.user_id::text,
			p.match_id::text,
			m.home_team,
			m.away_team
		from predictions p
		join world_cup_matches m on m.id = p.match_id
		join group_members gm on gm.group_id = p.group_id
			and gm.user_id = p.user_id
			and gm.status = 'active'
		where p.match_id = $1
			and p.home_score = $2
			and p.away_score = $3
	`, matchID, homeScore, awayScore)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []FeedEventInput{}
	for rows.Next() {
		var groupID string
		var actorID string
		var scannedMatchID string
		var homeTeam string
		var awayTeam string
		if err := rows.Scan(&groupID, &actorID, &scannedMatchID, &homeTeam, &awayTeam); err != nil {
			return nil, err
		}

		events = append(events, FeedEventInput{
			ActorUserID: &actorID,
			EventType:   "exact_score",
			GroupID:     groupID,
			MatchID:     &scannedMatchID,
			Metadata: map[string]any{
				"awayScore": awayScore,
				"awayTeam":  awayTeam,
				"homeScore": homeScore,
				"homeTeam":  homeTeam,
				"score":     formatScore(homeScore, awayScore),
			},
		})
	}

	return events, rows.Err()
}

func GroupsForFinishedMatch(ctx context.Context, db Querier, matchID string) ([]FeedEventInput, error) {
	rows, err := db.Query(ctx, `
		select distinct
			g.id::text,
			m.id::text,
			m.home_team,
			m.away_team,
			m.home_score,
			m.away_score
		from groups g
		join group_members gm on gm.group_id = g.id and gm.status = 'active'
		join world_cup_matches m on m.id = $1
		left join predictions p on p.group_id = g.id and p.match_id = m.id
		where p.group_id is not null
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []FeedEventInput{}
	for rows.Next() {
		var groupID string
		var scannedMatchID string
		var homeTeam string
		var awayTeam string
		var homeScore *int
		var awayScore *int
		if err := rows.Scan(&groupID, &scannedMatchID, &homeTeam, &awayTeam, &homeScore, &awayScore); err != nil {
			return nil, err
		}

		metadata := map[string]any{
			"awayTeam": awayTeam,
			"homeTeam": homeTeam,
		}
		if homeScore != nil && awayScore != nil {
			metadata["homeScore"] = *homeScore
			metadata["awayScore"] = *awayScore
			metadata["score"] = formatScore(*homeScore, *awayScore)
		}
		events = append(events, FeedEventInput{
			EventType: "match_finished",
			GroupID:   groupID,
			MatchID:   &scannedMatchID,
			Metadata:  metadata,
		})
	}

	return events, rows.Err()
}

func formatScore(homeScore int, awayScore int) string {
	return strconv.Itoa(homeScore) + "x" + strconv.Itoa(awayScore)
}
