package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/ai"
	"github.com/gabrielevieira/palpitai/backend/internal/explanations/models"
)

type Repository interface {
	FindPendingMatchesForExplanation(ctx context.Context, fromDate time.Time, toDate time.Time, limit int, promptVersion string) ([]models.ExplanationCandidate, error)
	UpsertExplanation(ctx context.Context, params models.UpsertExplanationParams) (string, error)
	MarkFailed(ctx context.Context, candidate models.ExplanationCandidate, promptVersion string, modelName string, inputSnapshot json.RawMessage, rawResponse json.RawMessage, message string) error
	MarkSkipped(ctx context.Context, candidate models.ExplanationCandidate, promptVersion string, reason string) error
}

type GenerationSummary struct {
	Processed     int
	Generated     int
	Skipped       int
	Failed        int
	PromptVersion string
}

type ExplanationGenerationService struct {
	repository    Repository
	aiClient      ai.AIClient
	modelName     string
	promptVersion string
	logger        *slog.Logger
}

func NewExplanationGenerationService(repository Repository, aiClient ai.AIClient, modelName string, logger *slog.Logger) ExplanationGenerationService {
	if logger == nil {
		logger = slog.Default()
	}
	return ExplanationGenerationService{
		repository:    repository,
		aiClient:      aiClient,
		modelName:     modelName,
		promptVersion: ai.PromptVersionPredictionExplanationV1,
		logger:        logger,
	}
}

func (s ExplanationGenerationService) Generate(ctx context.Context, fromDate time.Time, toDate time.Time, limit int) (GenerationSummary, error) {
	if limit <= 0 {
		limit = 50
	}
	candidates, err := s.repository.FindPendingMatchesForExplanation(ctx, fromDate, toDate, limit, s.promptVersion)
	if err != nil {
		return GenerationSummary{}, err
	}

	summary := GenerationSummary{PromptVersion: s.promptVersion}
	for _, candidate := range candidates {
		summary.Processed++
		input, skipReason, err := BuildPromptInput(candidate, s.promptVersion)
		if err != nil {
			return summary, err
		}
		if skipReason != "" {
			if err := s.repository.MarkSkipped(ctx, candidate, s.promptVersion, skipReason); err != nil {
				return summary, err
			}
			summary.Skipped++
			continue
		}

		inputSnapshot, err := json.Marshal(input)
		if err != nil {
			return summary, err
		}
		response, err := s.aiClient.GeneratePredictionExplanation(ctx, input)
		if err != nil {
			message := err.Error()
			var invalidResponse ai.InvalidResponseError
			var rawResponse json.RawMessage
			if errors.As(err, &invalidResponse) {
				rawResponse = invalidResponse.RawResponse
			}
			if markErr := s.repository.MarkFailed(ctx, candidate, s.promptVersion, s.modelName, inputSnapshot, rawResponse, message); markErr != nil {
				return summary, markErr
			}
			s.logger.Warn("prediction explanation generation failed", "match_date", candidate.MatchDate, "home_team_id", candidate.HomeTeamID, "away_team_id", candidate.AwayTeamID, "error", err)
			summary.Failed++
			continue
		}
		rawResponse := response.RawResponse
		if len(rawResponse) == 0 {
			rawResponse, _ = json.Marshal(response)
		}
		_, err = s.repository.UpsertExplanation(ctx, models.UpsertExplanationParams{
			MatchID:           candidate.MatchID,
			MatchPredictionID: candidate.MatchPredictionID,
			GoalPredictionID:  candidate.GoalPredictionID,
			HomeTeamID:        candidate.HomeTeamID,
			AwayTeamID:        candidate.AwayTeamID,
			MatchDate:         candidate.MatchDate,
			Summary:           response.Summary,
			MainReasons:       response.MainReasons,
			RiskAlert:         response.RiskAlert,
			BetStyle:          response.BetStyle,
			UserTip:           response.UserTip,
			ModelName:         s.modelName,
			PromptVersion:     s.promptVersion,
			InputSnapshot:     inputSnapshot,
			RawResponse:       rawResponse,
			Status:            "generated",
		})
		if err != nil {
			return summary, err
		}
		summary.Generated++
	}
	return summary, nil
}

func BuildPromptInput(candidate models.ExplanationCandidate, promptVersion string) (ai.ExplanationPromptInput, string, error) {
	if candidate.MatchPredictionID == nil {
		return ai.ExplanationPromptInput{}, "missing match_prediction", nil
	}
	if candidate.GoalPredictionID == nil {
		return ai.ExplanationPromptInput{}, "missing goal_prediction", nil
	}
	if candidate.HomeWinProbability == nil || candidate.DrawProbability == nil || candidate.AwayWinProbability == nil || candidate.PredictedLabel == nil || candidate.Confidence == nil {
		return ai.ExplanationPromptInput{}, "missing result prediction fields", nil
	}
	if candidate.ExpectedHomeGoals == nil || candidate.ExpectedAwayGoals == nil || candidate.MostLikelyHomeScore == nil || candidate.MostLikelyAwayScore == nil {
		return ai.ExplanationPromptInput{}, "missing goal prediction fields", nil
	}

	topScores := make([]ai.ScoreProbabilityData, 0, len(candidate.TopScoreProbabilities))
	for _, score := range candidate.TopScoreProbabilities {
		topScores = append(topScores, ai.ScoreProbabilityData{
			Score:       strconv.Itoa(score.HomeScore) + "x" + strconv.Itoa(score.AwayScore),
			Probability: percent(score.Probability),
		})
	}
	input := ai.ExplanationPromptInput{
		Match: ai.MatchPromptData{
			MatchID:   candidate.MatchID,
			HomeTeam:  candidate.HomeTeam,
			AwayTeam:  candidate.AwayTeam,
			MatchDate: candidate.MatchDate,
			Stage:     candidate.Stage,
		},
		ResultPrediction: ai.ResultPredictionData{
			HomeWinProbability: percent(*candidate.HomeWinProbability),
			DrawProbability:    percent(*candidate.DrawProbability),
			AwayWinProbability: percent(*candidate.AwayWinProbability),
			PredictedLabel:     *candidate.PredictedLabel,
			Confidence:         *candidate.Confidence,
		},
		GoalsPrediction: ai.GoalsPredictionData{
			ExpectedHomeGoals:         round2(*candidate.ExpectedHomeGoals),
			ExpectedAwayGoals:         round2(*candidate.ExpectedAwayGoals),
			MostLikelyScore:           fmt.Sprintf("%dx%d", *candidate.MostLikelyHomeScore, *candidate.MostLikelyAwayScore),
			Over25Probability:         percentPtr(candidate.Over25Probability),
			BothTeamsScoreProbability: percentPtr(candidate.BothTeamsScoreProbability),
		},
		TopScoreProbabilities: topScores,
		KeyMetrics: ai.KeyMetricsData{
			EloDiff:                  candidate.EloDiff,
			FifaRankDiff:             candidate.FifaRankDiff,
			HomeAttackScore:          candidate.HomeAttackScore,
			AwayAttackScore:          candidate.AwayAttackScore,
			HomeDefenseScore:         candidate.HomeDefenseScore,
			AwayDefenseScore:         candidate.AwayDefenseScore,
			HomeRecentFormScore:      candidate.HomeRecentFormScore,
			AwayRecentFormScore:      candidate.AwayRecentFormScore,
			HomeWorldCupHistoryScore: candidate.HomeWorldCupHistoryScore,
			AwayWorldCupHistoryScore: candidate.AwayWorldCupHistoryScore,
		},
		PromptVersion: promptVersion,
	}
	return input, "", nil
}

func percent(value float64) float64 {
	if value <= 1 {
		return round2(value * 100)
	}
	return round2(value)
}

func percentPtr(value *float64) float64 {
	if value == nil {
		return 0
	}
	return percent(*value)
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
