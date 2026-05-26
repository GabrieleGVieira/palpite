package ml

type PredictRequest struct {
	MatchDate  string         `json:"match_date"`
	HomeTeamID string         `json:"home_team_id"`
	AwayTeamID string         `json:"away_team_id"`
	Features   map[string]any `json:"features"`
}

type PredictResponse struct {
	HomeWinProbability float64 `json:"home_win_probability"`
	DrawProbability    float64 `json:"draw_probability"`
	AwayWinProbability float64 `json:"away_win_probability"`
	PredictedLabel     string  `json:"predicted_label"`
	Confidence         string  `json:"confidence"`
	ModelVersion       string  `json:"model_version"`
}
