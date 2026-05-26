package repository

import (
	"context"
	"errors"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/predictions/models"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
	"github.com/jackc/pgx/v5"
)

func ByMatchID(ctx context.Context, db repositories.Querier, matchID string) (models.MatchPrediction, error) {
	var prediction models.MatchPrediction
	err := db.QueryRow(ctx, `
		select
			id::text,
			match_id::text,
			match_date,
			home_team_id::text,
			away_team_id::text,
			model_id::text,
			home_win_probability::float8,
			draw_probability::float8,
			away_win_probability::float8,
			predicted_label,
			confidence,
			suggested_home_score,
			suggested_away_score,
			model_version,
			source,
			created_at,
			updated_at
		from match_predictions
		where match_id = $1
		order by created_at desc
		limit 1
	`, matchID).Scan(
		&prediction.ID,
		&prediction.MatchID,
		&prediction.MatchDate,
		&prediction.HomeTeamID,
		&prediction.AwayTeamID,
		&prediction.ModelID,
		&prediction.HomeWinProbability,
		&prediction.DrawProbability,
		&prediction.AwayWinProbability,
		&prediction.PredictedLabel,
		&prediction.Confidence,
		&prediction.SuggestedHomeScore,
		&prediction.SuggestedAwayScore,
		&prediction.ModelVersion,
		&prediction.Source,
		&prediction.CreatedAt,
		&prediction.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.MatchPrediction{}, repositories.ErrNotFound
	}
	return prediction, err
}

func ByDateAndTeams(ctx context.Context, db repositories.Querier, matchDate time.Time, homeTeamID string, awayTeamID string) (models.MatchPrediction, error) {
	var prediction models.MatchPrediction
	err := db.QueryRow(ctx, `
		select
			id::text,
			match_id::text,
			match_date,
			home_team_id::text,
			away_team_id::text,
			model_id::text,
			home_win_probability::float8,
			draw_probability::float8,
			away_win_probability::float8,
			predicted_label,
			confidence,
			suggested_home_score,
			suggested_away_score,
			model_version,
			source,
			created_at,
			updated_at
		from match_predictions
		where match_date = $1 and home_team_id = $2 and away_team_id = $3
		order by created_at desc
		limit 1
	`, matchDate, homeTeamID, awayTeamID).Scan(
		&prediction.ID,
		&prediction.MatchID,
		&prediction.MatchDate,
		&prediction.HomeTeamID,
		&prediction.AwayTeamID,
		&prediction.ModelID,
		&prediction.HomeWinProbability,
		&prediction.DrawProbability,
		&prediction.AwayWinProbability,
		&prediction.PredictedLabel,
		&prediction.Confidence,
		&prediction.SuggestedHomeScore,
		&prediction.SuggestedAwayScore,
		&prediction.ModelVersion,
		&prediction.Source,
		&prediction.CreatedAt,
		&prediction.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.MatchPrediction{}, repositories.ErrNotFound
	}
	return prediction, err
}
