package repository

import (
	"context"
	"errors"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/goalpredictions/models"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
	"github.com/jackc/pgx/v5"
)

func ByMatchID(ctx context.Context, db repositories.Querier, matchID string) (models.MatchGoalPrediction, error) {
	prediction, err := scanGoalPrediction(ctx, db, `
		select
			id::text,
			match_id::text,
			match_date,
			home_team_id::text,
			away_team_id::text,
			goal_model_id::text,
			result_model_id::text,
			expected_home_goals::float8,
			expected_away_goals::float8,
			most_likely_home_score,
			most_likely_away_score,
			over_1_5_probability::float8,
			over_2_5_probability::float8,
			both_teams_score_probability::float8,
			calibration_method,
			score_probability_mass::float8,
			calibrated_at,
			model_version,
			source,
			created_at,
			updated_at
		from match_goal_predictions
		where match_id = $1
		order by created_at desc
		limit 1
	`, matchID)
	if err != nil {
		return models.MatchGoalPrediction{}, err
	}
	prediction.TopScoreProbabilities, err = ScoreProbabilities(ctx, db, prediction.ID)
	return prediction, err
}

func ByDateAndTeams(ctx context.Context, db repositories.Querier, matchDate time.Time, homeTeamID string, awayTeamID string) (models.MatchGoalPrediction, error) {
	prediction, err := scanGoalPrediction(ctx, db, `
		select
			id::text,
			match_id::text,
			match_date,
			home_team_id::text,
			away_team_id::text,
			goal_model_id::text,
			result_model_id::text,
			expected_home_goals::float8,
			expected_away_goals::float8,
			most_likely_home_score,
			most_likely_away_score,
			over_1_5_probability::float8,
			over_2_5_probability::float8,
			both_teams_score_probability::float8,
			calibration_method,
			score_probability_mass::float8,
			calibrated_at,
			model_version,
			source,
			created_at,
			updated_at
		from match_goal_predictions
		where match_date = $1 and home_team_id = $2 and away_team_id = $3
		order by created_at desc
		limit 1
	`, matchDate, homeTeamID, awayTeamID)
	if err != nil {
		return models.MatchGoalPrediction{}, err
	}
	prediction.TopScoreProbabilities, err = ScoreProbabilities(ctx, db, prediction.ID)
	return prediction, err
}

func ScoreProbabilities(ctx context.Context, db repositories.Querier, matchGoalPredictionID string) ([]models.MatchScoreProbability, error) {
	rows, err := db.Query(ctx, `
		select
			id::text,
			match_goal_prediction_id::text,
			home_score,
			away_score,
			probability::float8,
			created_at
		from match_score_probabilities
		where match_goal_prediction_id = $1
		order by probability desc, home_score asc, away_score asc
	`, matchGoalPredictionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	probabilities := []models.MatchScoreProbability{}
	for rows.Next() {
		var probability models.MatchScoreProbability
		if err := rows.Scan(
			&probability.ID,
			&probability.MatchGoalPredictionID,
			&probability.HomeScore,
			&probability.AwayScore,
			&probability.Probability,
			&probability.CreatedAt,
		); err != nil {
			return nil, err
		}
		probabilities = append(probabilities, probability)
	}
	return probabilities, rows.Err()
}

func scanGoalPrediction(ctx context.Context, db repositories.Querier, sql string, args ...any) (models.MatchGoalPrediction, error) {
	var prediction models.MatchGoalPrediction
	err := db.QueryRow(ctx, sql, args...).Scan(
		&prediction.ID,
		&prediction.MatchID,
		&prediction.MatchDate,
		&prediction.HomeTeamID,
		&prediction.AwayTeamID,
		&prediction.GoalModelID,
		&prediction.ResultModelID,
		&prediction.ExpectedHomeGoals,
		&prediction.ExpectedAwayGoals,
		&prediction.MostLikelyHomeScore,
		&prediction.MostLikelyAwayScore,
		&prediction.Over15Probability,
		&prediction.Over25Probability,
		&prediction.BothTeamsScoreProbability,
		&prediction.CalibrationMethod,
		&prediction.ScoreProbabilityMass,
		&prediction.CalibratedAt,
		&prediction.ModelVersion,
		&prediction.Source,
		&prediction.CreatedAt,
		&prediction.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.MatchGoalPrediction{}, repositories.ErrNotFound
	}
	return prediction, err
}
