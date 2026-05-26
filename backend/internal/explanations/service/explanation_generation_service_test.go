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
	candidates  []models.ExplanationCandidate
	generated   int
	failed      int
	generating  int
	retryCounts []int
}

func (r *fakeRepository) FindPendingMatchesForExplanation(ctx context.Context, fromDate time.Time, toDate time.Time, limit int, promptVersion string, staleBefore time.Time) ([]models.ExplanationCandidate, error) {
	return r.candidates, nil
}

func (r *fakeRepository) MarkGenerating(ctx context.Context, candidate models.ExplanationCandidate, promptVersion string, modelName string, inputSnapshot json.RawMessage) error {
	r.generating++
	return nil
}

func (r *fakeRepository) UpsertExplanation(ctx context.Context, params models.UpsertExplanationParams) (string, error) {
	r.generated++
	return "explanation-id", nil
}

func (r *fakeRepository) MarkFailed(ctx context.Context, candidate models.ExplanationCandidate, promptVersion string, modelName string, inputSnapshot json.RawMessage, rawResponse json.RawMessage, message string) error {
	r.failed++
	r.retryCounts = append(r.retryCounts, candidate.ExistingRetryCount)
	return nil
}

type fakeAIClient struct {
	err               error
	errs              []error
	predictions       []ai.BatchPredictionExplanation
	predictionBatches [][]ai.BatchPredictionExplanation
	callMatchIDs      [][]string
	calls             int
}

func (c *fakeAIClient) GeneratePredictionExplanations(ctx context.Context, inputs []ai.ExplanationPromptInput) (*ai.BatchExplanationAIResponse, error) {
	matchIDs := make([]string, 0, len(inputs))
	for _, input := range inputs {
		matchIDs = append(matchIDs, *input.Match.MatchID)
	}
	c.callMatchIDs = append(c.callMatchIDs, matchIDs)
	if c.calls < len(c.errs) {
		err := c.errs[c.calls]
		c.calls++
		if err != nil {
			return nil, err
		}
	} else {
		c.calls++
	}
	if c.err != nil {
		return nil, c.err
	}
	if c.calls-1 < len(c.predictionBatches) {
		return &ai.BatchExplanationAIResponse{
			Predictions: c.predictionBatches[c.calls-1],
			RawResponse: []byte(`{"predictions":[]}`),
		}, nil
	}
	if c.predictions != nil {
		return &ai.BatchExplanationAIResponse{
			Predictions: c.predictions,
			RawResponse: []byte(`{"predictions":[]}`),
		}, nil
	}
	predictions := make([]ai.BatchPredictionExplanation, 0, len(inputs))
	for _, input := range inputs {
		predictions = append(predictions, ai.BatchPredictionExplanation{
			MatchID:     *input.Match.MatchID,
			Explanation: "Brasil aparece como leve favorito.",
			KeyFactors:  []string{"Tem leve vantagem nas metricas.", "Os placares indicam equilibrio."},
			RiskLevel:   "medium",
		})
	}
	return &ai.BatchExplanationAIResponse{
		Predictions: predictions,
		RawResponse: []byte(`{"predictions":[]}`),
	}, nil
}

func TestServiceSavesReturnedMatchesAndFailsMissingMatches(t *testing.T) {
	secondMatchID := "match-2"
	repo := &fakeRepository{candidates: []models.ExplanationCandidate{
		validCandidate(nil),
		validCandidate(func(c *models.ExplanationCandidate) {
			c.MatchID = &secondMatchID
		}),
	}}
	client := &fakeAIClient{predictions: []ai.BatchPredictionExplanation{
		{
			MatchID:     "match",
			Explanation: "Brasil aparece como leve favorito.",
			KeyFactors:  []string{"Tem leve vantagem nas metricas.", "Os placares indicam equilibrio."},
			RiskLevel:   "medium",
		},
	}}
	service := NewExplanationGenerationService(repo, client, "gemini-test", slog.Default())
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10, 5, time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if summary.Generated != 1 || repo.generated != 1 {
		t.Fatalf("Generated = %d repo=%d", summary.Generated, repo.generated)
	}
	if summary.Failed != 1 || repo.failed != 1 {
		t.Fatalf("Failed = %d repo=%d", summary.Failed, repo.failed)
	}
}

func TestServiceRetriesOnlyMissingMatchesWithFallbackBatchSize(t *testing.T) {
	secondMatchID := "match-2"
	thirdMatchID := "match-3"
	repo := &fakeRepository{candidates: []models.ExplanationCandidate{
		validCandidate(nil),
		validCandidate(func(c *models.ExplanationCandidate) {
			c.MatchID = &secondMatchID
		}),
		validCandidate(func(c *models.ExplanationCandidate) {
			c.MatchID = &thirdMatchID
		}),
	}}
	client := &fakeAIClient{predictionBatches: [][]ai.BatchPredictionExplanation{
		{
			{
				MatchID:     "match",
				Explanation: "Brasil aparece como leve favorito.",
				KeyFactors:  []string{"Tem leve vantagem nas metricas.", "Os placares indicam equilibrio."},
				RiskLevel:   "medium",
			},
		},
		{
			{
				MatchID:     "match-2",
				Explanation: "Brasil aparece como leve favorito.",
				KeyFactors:  []string{"Tem leve vantagem nas metricas.", "Os placares indicam equilibrio."},
				RiskLevel:   "medium",
			},
		},
		{
			{
				MatchID:     "match-3",
				Explanation: "Brasil aparece como leve favorito.",
				KeyFactors:  []string{"Tem leve vantagem nas metricas.", "Os placares indicam equilibrio."},
				RiskLevel:   "medium",
			},
		},
	}}
	service := NewExplanationGenerationService(repo, client, "gemini-test", slog.Default()).
		WithMissingRetry(1, true, 2)
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10, 3, time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if summary.Generated != 3 || repo.generated != 3 || summary.Failed != 0 {
		t.Fatalf("summary = %+v repo generated=%d failed=%d", summary, repo.generated, repo.failed)
	}
	if len(client.callMatchIDs) != 3 {
		t.Fatalf("callMatchIDs = %+v", client.callMatchIDs)
	}
	if got := client.callMatchIDs[1]; len(got) != 1 || got[0] != "match-2" {
		t.Fatalf("second call processed %+v, want [match-2]", got)
	}
	if got := client.callMatchIDs[2]; len(got) != 1 || got[0] != "match-3" {
		t.Fatalf("third call processed %+v, want [match-3]", got)
	}
}

func TestServiceMarksFailedWhenMissingMatchPrediction(t *testing.T) {
	repo := &fakeRepository{candidates: []models.ExplanationCandidate{validCandidate(func(c *models.ExplanationCandidate) {
		c.MatchPredictionID = nil
	})}}
	service := NewExplanationGenerationService(repo, &fakeAIClient{}, "gpt-test", slog.Default())
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10, 5, time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if summary.Failed != 1 || repo.failed != 1 {
		t.Fatalf("Failed = %d repo=%d", summary.Failed, repo.failed)
	}
}

func TestServiceMarksFailedWhenMissingGoalPrediction(t *testing.T) {
	repo := &fakeRepository{candidates: []models.ExplanationCandidate{validCandidate(func(c *models.ExplanationCandidate) {
		c.GoalPredictionID = nil
	})}}
	service := NewExplanationGenerationService(repo, &fakeAIClient{}, "gpt-test", slog.Default())
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10, 5, time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if summary.Failed != 1 {
		t.Fatalf("Failed = %d", summary.Failed)
	}
}

func TestServiceDoesNotRegenerateWhenRepositoryReturnsNoPending(t *testing.T) {
	repo := &fakeRepository{}
	service := NewExplanationGenerationService(repo, &fakeAIClient{}, "gpt-test", slog.Default())
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10, 5, time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if summary.Processed != 0 || summary.Generated != 0 {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestServiceSavesFailedWhenAIReturnsInvalidTwice(t *testing.T) {
	repo := &fakeRepository{candidates: []models.ExplanationCandidate{validCandidate(nil)}}
	service := NewExplanationGenerationService(repo, &fakeAIClient{err: errors.New("invalid response after retry")}, "gpt-test", slog.Default())
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10, 5, time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if summary.Failed != 1 || repo.failed != 1 {
		t.Fatalf("Failed = %d repo=%d", summary.Failed, repo.failed)
	}
}

func TestServiceStopsWithoutMarkingFailedWhenRateLimited(t *testing.T) {
	repo := &fakeRepository{candidates: []models.ExplanationCandidate{
		validCandidate(nil),
		validCandidate(func(c *models.ExplanationCandidate) {
			matchID := "match-2"
			c.MatchID = &matchID
		}),
	}}
	client := &fakeAIClient{err: ai.RateLimitError{RetryAfter: time.Minute, Message: "quota exceeded"}}
	service := NewExplanationGenerationService(repo, client, "gemini-test", slog.Default())
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10, 5, time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !summary.RateLimited {
		t.Fatalf("expected RateLimited summary")
	}
	if summary.Processed != 2 {
		t.Fatalf("Processed = %d", summary.Processed)
	}
	if summary.Failed != 2 || repo.failed != 2 {
		t.Fatalf("Failed = %d repo=%d", summary.Failed, repo.failed)
	}
	if repo.generated != 0 {
		t.Fatalf("generated = %d", repo.generated)
	}
}

func TestServiceWaitsAndRetriesSameCandidateWhenRateLimited(t *testing.T) {
	repo := &fakeRepository{candidates: []models.ExplanationCandidate{validCandidate(nil)}}
	client := &fakeAIClient{errs: []error{
		ai.RateLimitError{RetryAfter: time.Millisecond, Message: "quota exceeded"},
		nil,
	}}
	service := NewExplanationGenerationService(repo, client, "gemini-test", slog.Default()).
		WithRateLimitCooldown(time.Millisecond, 1)
	summary, err := service.Generate(context.Background(), time.Now(), time.Now(), 10, 5, time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !summary.RateLimited || summary.RateLimitWaits != 1 {
		t.Fatalf("summary = %+v", summary)
	}
	if summary.Generated != 1 || repo.generated != 1 {
		t.Fatalf("Generated = %d repo=%d", summary.Generated, repo.generated)
	}
	if client.calls != 2 {
		t.Fatalf("calls = %d", client.calls)
	}
}

func validCandidate(mutator func(*models.ExplanationCandidate)) models.ExplanationCandidate {
	matchID := "match"
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
		MatchID:                   &matchID,
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
