package models

import (
	"encoding/json"
	"time"
)

type PredictionExplanation struct {
	ID                string
	MatchID           *string
	MatchPredictionID *string
	GoalPredictionID  *string
	HomeTeamID        string
	AwayTeamID        string
	MatchDate         time.Time
	Summary           string
	MainReasons       []string
	RiskAlert         *string
	BetStyle          *string
	UserTip           *string
	ModelName         string
	PromptVersion     string
	InputSnapshot     json.RawMessage
	RawResponse       json.RawMessage
	Status            string
	ErrorMessage      *string
	GeneratedAt       *time.Time
	Version           int
	RetryCount        int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ExplanationCandidate struct {
	MatchID                   *string
	MatchDate                 time.Time
	Stage                     *string
	HomeTeamID                string
	AwayTeamID                string
	HomeTeam                  string
	AwayTeam                  string
	MatchPredictionID         *string
	GoalPredictionID          *string
	ResultModelID             *string
	HomeWinProbability        *float64
	DrawProbability           *float64
	AwayWinProbability        *float64
	PredictedLabel            *string
	Confidence                *string
	ExpectedHomeGoals         *float64
	ExpectedAwayGoals         *float64
	MostLikelyHomeScore       *int
	MostLikelyAwayScore       *int
	Over25Probability         *float64
	BothTeamsScoreProbability *float64
	EloDiff                   *float64
	FifaRankDiff              *float64
	HomeAttackScore           *float64
	AwayAttackScore           *float64
	HomeDefenseScore          *float64
	AwayDefenseScore          *float64
	HomeRecentFormScore       *float64
	AwayRecentFormScore       *float64
	HomeWorldCupHistoryScore  *float64
	AwayWorldCupHistoryScore  *float64
	TopScoreProbabilities     []ScoreProbability
	ExistingExplanationID     *string
	ExistingGeneratedAt       *time.Time
	ExistingStatus            *string
	ExistingRetryCount        int
}

type ScoreProbability struct {
	HomeScore   int
	AwayScore   int
	Probability float64
}

type UpsertExplanationParams struct {
	MatchID           *string
	MatchPredictionID *string
	GoalPredictionID  *string
	HomeTeamID        string
	AwayTeamID        string
	MatchDate         time.Time
	Summary           string
	MainReasons       []string
	RiskAlert         *string
	BetStyle          string
	UserTip           string
	ModelName         string
	PromptVersion     string
	InputSnapshot     json.RawMessage
	RawResponse       json.RawMessage
	Status            string
	ErrorMessage      *string
	RetryCount        int
}
