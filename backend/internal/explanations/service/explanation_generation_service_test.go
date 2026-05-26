package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/ai"
	"github.com/gabrielevieira/palpitai/backend/internal/explanations/models"
)

type fakeRepository struct {
	candidates []models.ExplanationCandidate
	generated  int
	skipped    int
	failed     int
}

func (r *fakeRepository) FindPendingMatchesForExplanation(ctx context.Context, fromDate time.Time, toDate time.Time, limit int, promptVersion string) ([]models.ExplanationCandidate, error) {
	return r.candidates, nil
}

func (r *fakeRepository) UpsertExplanation(ctx context.Context, params models.UpsertExplanationParams) (string, error) {
	r.generated++
	return "explanation-id", nil
}

func (r *fakeRepository) MarkFailed(ctx context.Context, candidate models.ExplanationCandidate, promptVersion string, modelName string, inputSnapshot json.RawMessage, rawResponse json.RawMessage, message string) error {
	r.failed++
	return nil
}

func (r *fakeRepository) MarkSkipped(ctx context.Context, candidate models.ExplanationCandidate, promptVersion string, reason string) error {
	r.skipped++
	return nil
}

type fakeAIClient struct {
	err error
}

func (c fakeAIClient) GeneratePredictionExplanation(ctx context.Context, input ai.ExplanationPromptInput) (*ai.ExplanationAIResponse, error) {
	if c.err != nil {
		return nil, c.err
	}
	risk := "Ha incerteza relevante."
	return &ai.ExplanationAIResponse{
		Summary:     "Brasil aparece como leve favorito.",
		MainReasons: []string{"Tem leve vantagem nas metricas.", "Os placares indicam equilibrio."},
		RiskAlert:   &risk,
		BetStyle:    "moderate",
		UserTip:     "Use como leitura de tendencia, sem promessa de acerto.",
		RawResponse: []byte(`{"summary":"Brasil aparece como leve favorito.","main_reasons":["Tem leve vantagem nas metricas.","Os placares indicam equilibrio."],"risk_alert":"Ha incerteza relevante.","bet_style":"moderate","user_tip":"Use como leitura de tendencia, sem promessa de acerto."}`),
	}, nil
}

func TestServiceMarksSkippedWhenMissingMatchPrediction(t *testing.T) {
	repo := &fakeRepository{candidates: []models.ExplanationCandidate{validCandidate(func(c *models.ExplanationCandidate) {
		c.MatchPredictionID = nil
	})}}
	service := NewExplanationGenerationService(repo, fakeAIClient{}, "gpt-test", slog.Default())
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if summary.Skipped != 1 || repo.skipped != 1 {
		t.Fatalf("Skipped = %d repo=%d", summary.Skipped, repo.skipped)
	}
}

func TestServiceMarksSkippedWhenMissingGoalPrediction(t *testing.T) {
	repo := &fakeRepository{candidates: []models.ExplanationCandidate{validCandidate(func(c *models.ExplanationCandidate) {
		c.GoalPredictionID = nil
	})}}
	service := NewExplanationGenerationService(repo, fakeAIClient{}, "gpt-test", slog.Default())
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if summary.Skipped != 1 {
		t.Fatalf("Skipped = %d", summary.Skipped)
	}
}

func TestServiceDoesNotRegenerateWhenRepositoryReturnsNoPending(t *testing.T) {
	repo := &fakeRepository{}
	service := NewExplanationGenerationService(repo, fakeAIClient{}, "gpt-test", slog.Default())
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if summary.Processed != 0 || summary.Generated != 0 {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestServiceSavesFailedWhenAIReturnsInvalidTwice(t *testing.T) {
	repo := &fakeRepository{candidates: []models.ExplanationCandidate{validCandidate(nil)}}
	service := NewExplanationGenerationService(repo, fakeAIClient{err: errors.New("invalid response after retry")}, "gpt-test", slog.Default())
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if summary.Failed != 1 || repo.failed != 1 {
		t.Fatalf("Failed = %d repo=%d", summary.Failed, repo.failed)
	}
}

func validCandidate(mutator func(*models.ExplanationCandidate)) models.ExplanationCandidate {
	matchPredictionID := "mp"
	goalPredictionID := "gp"
	homeWin := 0.46
	draw := 0.27
	awayWin := 0.27
	label := "HOME_WIN"
	confidence := "medium"
	expectedHome := 1.8
	expectedAway := 1.2
	homeScore := 2
	awayScore := 1
	over25 := 0.48
	btts := 0.52
	candidate := models.ExplanationCandidate{
		MatchDate:                 time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC),
		HomeTeamID:                "home",
		AwayTeamID:                "away",
		HomeTeam:                  "Brasil",
		AwayTeam:                  "Franca",
		MatchPredictionID:         &matchPredictionID,
		GoalPredictionID:          &goalPredictionID,
		HomeWinProbability:        &homeWin,
		DrawProbability:           &draw,
		AwayWinProbability:        &awayWin,
		PredictedLabel:            &label,
		Confidence:                &confidence,
		ExpectedHomeGoals:         &expectedHome,
		ExpectedAwayGoals:         &expectedAway,
		MostLikelyHomeScore:       &homeScore,
		MostLikelyAwayScore:       &awayScore,
		Over25Probability:         &over25,
		BothTeamsScoreProbability: &btts,
		TopScoreProbabilities:     []models.ScoreProbability{{HomeScore: 2, AwayScore: 1, Probability: 0.14}},
	}
	if mutator != nil {
		mutator(&candidate)
	}
	return candidate
}
