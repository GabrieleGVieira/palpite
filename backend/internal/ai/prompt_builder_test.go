package ai

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestBuildPredictionExplanationPromptContainsEssentialData(t *testing.T) {
	input := validPromptInput()
	prompt, err := BuildPredictionExplanationPrompt(input)
	if err != nil {
		t.Fatalf("BuildPredictionExplanationPrompt() error = %v", err)
	}
	if !strings.Contains(prompt.System, "Nao calcule vencedor") {
		t.Fatalf("system prompt must prevent model from calculating prediction")
	}
	if !strings.Contains(prompt.User, "Brasil") || !strings.Contains(prompt.User, "Franca") {
		t.Fatalf("user prompt must include match teams")
	}
	if !strings.Contains(prompt.User, "home_win_probability") {
		t.Fatalf("user prompt must include result probabilities")
	}
}

func validPromptInput() ExplanationPromptInput {
	risk := "Jogo equilibrado."
	return ExplanationPromptInput{
		Match: MatchPromptData{
			HomeTeam:  "Brasil",
			AwayTeam:  "Franca",
			MatchDate: time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC),
		},
		ResultPrediction: ResultPredictionData{
			HomeWinProbability: 46,
			DrawProbability:    27,
			AwayWinProbability: 27,
			PredictedLabel:     "HOME_WIN",
			Confidence:         "medium",
		},
		GoalsPrediction: GoalsPredictionData{
			ExpectedHomeGoals:         1.8,
			ExpectedAwayGoals:         1.2,
			MostLikelyScore:           "2x1",
			Over25Probability:         48,
			BothTeamsScoreProbability: 52,
		},
		TopScoreProbabilities: []ScoreProbabilityData{{Score: "1x0", Probability: 14}},
		KeyMetrics:            KeyMetricsData{},
		PromptVersion:         PromptVersionPredictionExplanationV1,
		Raw: map[string]json.RawMessage{
			"risk": json.RawMessage(`"` + risk + `"`),
		},
	}
}
