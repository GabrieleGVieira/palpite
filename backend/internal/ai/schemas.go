package ai

import (
	"context"
	"encoding/json"
	"time"
)

const PromptVersionPredictionExplanationV1 = "prediction-explanation-v1"

type AIClient interface {
	GeneratePredictionExplanations(ctx context.Context, inputs []ExplanationPromptInput) (*BatchExplanationAIResponse, error)
}

type ExplanationPromptInput struct {
	Match                 MatchPromptData            `json:"match"`
	ResultPrediction      ResultPredictionData       `json:"result_prediction"`
	GoalsPrediction       GoalsPredictionData        `json:"goals_prediction"`
	TopScoreProbabilities []ScoreProbabilityData     `json:"top_score_probabilities"`
	KeyMetrics            KeyMetricsData             `json:"key_metrics"`
	PromptVersion         string                     `json:"prompt_version"`
	Raw                   map[string]json.RawMessage `json:"-"`
}

type MatchPromptData struct {
	MatchID   *string   `json:"match_id,omitempty"`
	HomeTeam  string    `json:"home_team"`
	AwayTeam  string    `json:"away_team"`
	MatchDate time.Time `json:"match_date"`
	Stage     *string   `json:"stage,omitempty"`
}

type ResultPredictionData struct {
	HomeWinProbability float64 `json:"home_win_probability"`
	DrawProbability    float64 `json:"draw_probability"`
	AwayWinProbability float64 `json:"away_win_probability"`
	PredictedLabel     string  `json:"predicted_label"`
	Confidence         string  `json:"confidence"`
}

type GoalsPredictionData struct {
	ExpectedHomeGoals         float64 `json:"expected_home_goals"`
	ExpectedAwayGoals         float64 `json:"expected_away_goals"`
	MostLikelyScore           string  `json:"most_likely_score"`
	Over25Probability         float64 `json:"over_2_5_probability"`
	BothTeamsScoreProbability float64 `json:"both_teams_score_probability"`
}

type ScoreProbabilityData struct {
	Score       string  `json:"score"`
	Probability float64 `json:"probability"`
}

type KeyMetricsData struct {
	EloDiff                  *float64 `json:"elo_diff,omitempty"`
	FifaRankDiff             *float64 `json:"fifa_rank_diff,omitempty"`
	HomeAttackScore          *float64 `json:"home_attack_score,omitempty"`
	AwayAttackScore          *float64 `json:"away_attack_score,omitempty"`
	HomeDefenseScore         *float64 `json:"home_defense_score,omitempty"`
	AwayDefenseScore         *float64 `json:"away_defense_score,omitempty"`
	HomeRecentFormScore      *float64 `json:"home_recent_form_score,omitempty"`
	AwayRecentFormScore      *float64 `json:"away_recent_form_score,omitempty"`
	HomeWorldCupHistoryScore *float64 `json:"home_world_cup_history_score,omitempty"`
	AwayWorldCupHistoryScore *float64 `json:"away_world_cup_history_score,omitempty"`
}

type ExplanationAIResponse struct {
	Summary     string   `json:"summary"`
	MainReasons []string `json:"main_reasons"`
	RiskAlert   *string  `json:"risk_alert"`
	BetStyle    string   `json:"bet_style"`
	UserTip     string   `json:"user_tip"`
	RawResponse []byte   `json:"-"`
}

type BatchExplanationAIResponse struct {
	Predictions []BatchPredictionExplanation `json:"predictions"`
	RawResponse []byte                       `json:"-"`
}

type BatchPredictionExplanation struct {
	MatchID     string   `json:"match_id"`
	Explanation string   `json:"explanation"`
	KeyFactors  []string `json:"key_factors"`
	RiskLevel   string   `json:"risk_level"`
}
