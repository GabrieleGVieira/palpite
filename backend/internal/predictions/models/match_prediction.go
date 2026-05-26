package models

import "time"

type MatchPrediction struct {
	ID                 string
	MatchID            *string
	MatchDate          time.Time
	HomeTeamID         string
	AwayTeamID         string
	ModelID            *string
	HomeWinProbability float64
	DrawProbability    float64
	AwayWinProbability float64
	PredictedLabel     string
	Confidence         string
	SuggestedHomeScore *int
	SuggestedAwayScore *int
	ModelVersion       string
	Source             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
