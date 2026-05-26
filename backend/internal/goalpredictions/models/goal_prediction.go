package models

import "time"

type MatchGoalPrediction struct {
	ID                        string
	MatchID                   *string
	MatchDate                 time.Time
	HomeTeamID                string
	AwayTeamID                string
	GoalModelID               *string
	ResultModelID             *string
	ExpectedHomeGoals         float64
	ExpectedAwayGoals         float64
	MostLikelyHomeScore       *int
	MostLikelyAwayScore       *int
	Over15Probability         *float64
	Over25Probability         *float64
	BothTeamsScoreProbability *float64
	CalibrationMethod         *string
	ScoreProbabilityMass      *float64
	CalibratedAt              *time.Time
	ModelVersion              string
	Source                    string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	TopScoreProbabilities     []MatchScoreProbability
}

type MatchScoreProbability struct {
	ID                    string
	MatchGoalPredictionID string
	HomeScore             int
	AwayScore             int
	Probability           float64
	CreatedAt             time.Time
}
