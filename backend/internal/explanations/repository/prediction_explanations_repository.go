package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/explanations/models"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db repositories.Querier
}

func New(db repositories.Querier) Repository {
	return Repository{db: db}
}

func (r Repository) FindPendingMatchesForExplanation(ctx context.Context, fromDate time.Time, toDate time.Time, limit int, promptVersion string) ([]models.ExplanationCandidate, error) {
	rows, err := r.db.Query(ctx, `
		select
			coalesce(mp.match_id, mgp.match_id)::text as match_id,
			mf.match_date,
			mf.stage,
			mf.home_team_id::text,
			mf.away_team_id::text,
			ht.name as home_team,
			at.name as away_team,
			mp.id::text as match_prediction_id,
			mgp.id::text as goal_prediction_id,
			mp.model_id::text as result_model_id,
			mp.home_win_probability::float8,
			mp.draw_probability::float8,
			mp.away_win_probability::float8,
			mp.predicted_label,
			mp.confidence,
			mgp.expected_home_goals::float8,
			mgp.expected_away_goals::float8,
			mgp.most_likely_home_score,
			mgp.most_likely_away_score,
			mgp.over_2_5_probability::float8,
			mgp.both_teams_score_probability::float8,
			mf.elo_diff::float8,
			mf.fifa_rank_diff::float8,
			mf.home_attack_score::float8,
			mf.away_attack_score::float8,
			mf.home_defense_score::float8,
			mf.away_defense_score::float8,
			mf.home_recent_form_score::float8,
			mf.away_recent_form_score::float8,
			mf.home_world_cup_history_score::float8,
			mf.away_world_cup_history_score::float8,
			pe.id::text as existing_explanation_id
		from match_features mf
		join world_cup_matches wm on wm.id = mf.match_id
		join teams ht on ht.id = mf.home_team_id
		join teams at on at.id = mf.away_team_id
		left join lateral (
			select *
			from match_predictions mp
			where mp.match_date = mf.match_date
				and mp.home_team_id = mf.home_team_id
				and mp.away_team_id = mf.away_team_id
			order by mp.created_at desc
			limit 1
		) mp on true
		left join lateral (
			select *
			from match_goal_predictions mgp
			where mgp.match_date = mf.match_date
				and mgp.home_team_id = mf.home_team_id
				and mgp.away_team_id = mf.away_team_id
			order by mgp.updated_at desc
			limit 1
		) mgp on true
		left join prediction_explanations pe on pe.match_date = mf.match_date
			and pe.home_team_id = mf.home_team_id
			and pe.away_team_id = mf.away_team_id
			and pe.prompt_version = $3
			and pe.status = 'generated'
		where mf.match_date between $1 and $2
			and lower(wm.status) in ('scheduled', 'schedule', 'timed')
			and (
				pe.id is null
				or pe.match_prediction_id is distinct from mp.id
				or pe.goal_prediction_id is distinct from mgp.id
			)
		order by wm.kickoff_at asc, mf.match_date asc, mf.id asc
		limit $4
	`, fromDate, toDate, promptVersion, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := []models.ExplanationCandidate{}
	for rows.Next() {
		var candidate models.ExplanationCandidate
		if err := rows.Scan(
			&candidate.MatchID,
			&candidate.MatchDate,
			&candidate.Stage,
			&candidate.HomeTeamID,
			&candidate.AwayTeamID,
			&candidate.HomeTeam,
			&candidate.AwayTeam,
			&candidate.MatchPredictionID,
			&candidate.GoalPredictionID,
			&candidate.ResultModelID,
			&candidate.HomeWinProbability,
			&candidate.DrawProbability,
			&candidate.AwayWinProbability,
			&candidate.PredictedLabel,
			&candidate.Confidence,
			&candidate.ExpectedHomeGoals,
			&candidate.ExpectedAwayGoals,
			&candidate.MostLikelyHomeScore,
			&candidate.MostLikelyAwayScore,
			&candidate.Over25Probability,
			&candidate.BothTeamsScoreProbability,
			&candidate.EloDiff,
			&candidate.FifaRankDiff,
			&candidate.HomeAttackScore,
			&candidate.AwayAttackScore,
			&candidate.HomeDefenseScore,
			&candidate.AwayDefenseScore,
			&candidate.HomeRecentFormScore,
			&candidate.AwayRecentFormScore,
			&candidate.HomeWorldCupHistoryScore,
			&candidate.AwayWorldCupHistoryScore,
			&candidate.ExistingExplanationID,
		); err != nil {
			return nil, err
		}
		if candidate.GoalPredictionID != nil {
			scores, err := r.ScoreProbabilities(ctx, *candidate.GoalPredictionID)
			if err != nil {
				return nil, err
			}
			candidate.TopScoreProbabilities = scores
		}
		candidates = append(candidates, candidate)
	}
	return candidates, rows.Err()
}

func (r Repository) ScoreProbabilities(ctx context.Context, goalPredictionID string) ([]models.ScoreProbability, error) {
	rows, err := r.db.Query(ctx, `
		select home_score, away_score, probability::float8
		from match_score_probabilities
		where match_goal_prediction_id = $1
		order by probability desc, home_score asc, away_score asc
		limit 10
	`, goalPredictionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scores []models.ScoreProbability
	for rows.Next() {
		var score models.ScoreProbability
		if err := rows.Scan(&score.HomeScore, &score.AwayScore, &score.Probability); err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}
	return scores, rows.Err()
}

func (r Repository) UpsertExplanation(ctx context.Context, params models.UpsertExplanationParams) (string, error) {
	reasons, err := json.Marshal(params.MainReasons)
	if err != nil {
		return "", err
	}
	var id string
	err = r.db.QueryRow(ctx, `
		insert into prediction_explanations (
			match_id, match_prediction_id, goal_prediction_id, home_team_id, away_team_id,
			match_date, summary, main_reasons, risk_alert, bet_style, user_tip,
			model_name, prompt_version, input_snapshot, raw_response, status, error_message
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10, $11, $12, $13, $14::jsonb, $15::jsonb, $16, $17)
		on conflict (match_date, home_team_id, away_team_id, prompt_version) do update set
			match_id = excluded.match_id,
			match_prediction_id = excluded.match_prediction_id,
			goal_prediction_id = excluded.goal_prediction_id,
			summary = excluded.summary,
			main_reasons = excluded.main_reasons,
			risk_alert = excluded.risk_alert,
			bet_style = excluded.bet_style,
			user_tip = excluded.user_tip,
			model_name = excluded.model_name,
			input_snapshot = excluded.input_snapshot,
			raw_response = excluded.raw_response,
			status = excluded.status,
			error_message = excluded.error_message,
			updated_at = now()
		where excluded.status = 'generated'
			or prediction_explanations.status <> 'generated'
		returning id::text
	`, params.MatchID, params.MatchPredictionID, params.GoalPredictionID, params.HomeTeamID, params.AwayTeamID,
		params.MatchDate, params.Summary, string(reasons), params.RiskAlert, params.BetStyle, params.UserTip,
		params.ModelName, params.PromptVersion, string(ensureJSON(params.InputSnapshot, `{}`)), string(ensureJSON(params.RawResponse, `null`)), params.Status, params.ErrorMessage).Scan(&id)
	return id, err
}

func (r Repository) MarkFailed(ctx context.Context, candidate models.ExplanationCandidate, promptVersion string, modelName string, inputSnapshot json.RawMessage, rawResponse json.RawMessage, message string) error {
	_, err := r.UpsertExplanation(ctx, models.UpsertExplanationParams{
		MatchID:           candidate.MatchID,
		MatchPredictionID: candidate.MatchPredictionID,
		GoalPredictionID:  candidate.GoalPredictionID,
		HomeTeamID:        candidate.HomeTeamID,
		AwayTeamID:        candidate.AwayTeamID,
		MatchDate:         candidate.MatchDate,
		Summary:           "Falha ao gerar explicacao.",
		MainReasons:       []string{},
		ModelName:         modelName,
		PromptVersion:     promptVersion,
		InputSnapshot:     inputSnapshot,
		RawResponse:       rawResponse,
		Status:            "failed",
		ErrorMessage:      &message,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

func (r Repository) MarkSkipped(ctx context.Context, candidate models.ExplanationCandidate, promptVersion string, reason string) error {
	_, err := r.UpsertExplanation(ctx, models.UpsertExplanationParams{
		MatchID:           candidate.MatchID,
		MatchPredictionID: candidate.MatchPredictionID,
		GoalPredictionID:  candidate.GoalPredictionID,
		HomeTeamID:        candidate.HomeTeamID,
		AwayTeamID:        candidate.AwayTeamID,
		MatchDate:         candidate.MatchDate,
		Summary:           "Explicacao ignorada por dados insuficientes.",
		MainReasons:       []string{},
		ModelName:         "none",
		PromptVersion:     promptVersion,
		InputSnapshot:     json.RawMessage(`{}`),
		RawResponse:       json.RawMessage(`null`),
		Status:            "skipped",
		ErrorMessage:      &reason,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

func (r Repository) FindByMatchID(ctx context.Context, matchID string, promptVersion string) (models.PredictionExplanation, error) {
	return r.findOne(ctx, "match_id = $1 and prompt_version = $2 and status = 'generated'", matchID, promptVersion)
}

func (r Repository) FindByMatchPredictionID(ctx context.Context, matchPredictionID string, promptVersion string) (models.PredictionExplanation, error) {
	return r.findOne(ctx, "match_prediction_id = $1 and prompt_version = $2 and status = 'generated'", matchPredictionID, promptVersion)
}

func (r Repository) findOne(ctx context.Context, where string, args ...any) (models.PredictionExplanation, error) {
	query := `
		select id::text, match_id::text, match_prediction_id::text, goal_prediction_id::text,
			home_team_id::text, away_team_id::text, match_date, summary, main_reasons,
			risk_alert, bet_style, user_tip, model_name, prompt_version,
			input_snapshot, raw_response, status, error_message, created_at, updated_at
		from prediction_explanations
		where ` + where + `
		order by created_at desc
		limit 1
	`
	var explanation models.PredictionExplanation
	var mainReasons []byte
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&explanation.ID,
		&explanation.MatchID,
		&explanation.MatchPredictionID,
		&explanation.GoalPredictionID,
		&explanation.HomeTeamID,
		&explanation.AwayTeamID,
		&explanation.MatchDate,
		&explanation.Summary,
		&mainReasons,
		&explanation.RiskAlert,
		&explanation.BetStyle,
		&explanation.UserTip,
		&explanation.ModelName,
		&explanation.PromptVersion,
		&explanation.InputSnapshot,
		&explanation.RawResponse,
		&explanation.Status,
		&explanation.ErrorMessage,
		&explanation.CreatedAt,
		&explanation.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.PredictionExplanation{}, repositories.ErrNotFound
	}
	if err != nil {
		return models.PredictionExplanation{}, err
	}
	_ = json.Unmarshal(mainReasons, &explanation.MainReasons)
	return explanation, nil
}

func ensureJSON(value json.RawMessage, fallback string) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage(fallback)
	}
	return value
}
