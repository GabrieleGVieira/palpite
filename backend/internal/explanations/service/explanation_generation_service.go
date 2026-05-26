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
	FindPendingMatchesForExplanation(ctx context.Context, fromDate time.Time, toDate time.Time, limit int, promptVersion string, staleBefore time.Time) ([]models.ExplanationCandidate, error)
	MarkGenerating(ctx context.Context, candidate models.ExplanationCandidate, promptVersion string, modelName string, inputSnapshot json.RawMessage) error
	UpsertExplanation(ctx context.Context, params models.UpsertExplanationParams) (string, error)
	MarkFailed(ctx context.Context, candidate models.ExplanationCandidate, promptVersion string, modelName string, inputSnapshot json.RawMessage, rawResponse json.RawMessage, message string) error
}

type GenerationSummary struct {
	Processed      int
	Generated      int
	Failed         int
	RateLimited    bool
	RateLimitWaits int
	PromptVersion  string
}

type ExplanationGenerationService struct {
	repository        Repository
	aiClient          ai.AIClient
	modelName         string
	promptVersion     string
	requestDelay      time.Duration
	rateLimitCooldown time.Duration
	maxRateLimitWaits int
	minBatchSize      int
	retryMissing      bool
	maxMissingRetries int
	logger            *slog.Logger
}

func NewExplanationGenerationService(repository Repository, aiClient ai.AIClient, modelName string, logger *slog.Logger) ExplanationGenerationService {
	if logger == nil {
		logger = slog.Default()
	}
	return ExplanationGenerationService{
		repository:        repository,
		aiClient:          aiClient,
		modelName:         modelName,
		promptVersion:     ai.PromptVersionPredictionExplanationV1,
		minBatchSize:      1,
		retryMissing:      true,
		maxMissingRetries: 2,
		logger:            logger,
	}
}

func (s ExplanationGenerationService) WithRequestDelay(delay time.Duration) ExplanationGenerationService {
	if delay > 0 {
		s.requestDelay = delay
	}
	return s
}

func (s ExplanationGenerationService) WithRateLimitCooldown(delay time.Duration, maxWaits int) ExplanationGenerationService {
	if delay > 0 {
		s.rateLimitCooldown = delay
	}
	if maxWaits > 0 {
		s.maxRateLimitWaits = maxWaits
	}
	return s
}

func (s ExplanationGenerationService) WithMissingRetry(minBatchSize int, retryMissing bool, maxRetries int) ExplanationGenerationService {
	if minBatchSize > 0 {
		s.minBatchSize = minBatchSize
	}
	s.retryMissing = retryMissing
	if maxRetries >= 0 {
		s.maxMissingRetries = maxRetries
	}
	return s
}

func (s ExplanationGenerationService) Generate(ctx context.Context, fromDate time.Time, toDate time.Time, limit int, batchSize int, maxAge time.Duration) (GenerationSummary, error) {
	if limit <= 0 {
		limit = 50
	}
	if batchSize <= 0 {
		batchSize = 2
	}
	if s.minBatchSize <= 0 {
		s.minBatchSize = 1
	}
	if batchSize < s.minBatchSize {
		batchSize = s.minBatchSize
	}
	if maxAge <= 0 {
		maxAge = 24 * time.Hour
	}
	staleBefore := time.Now().Add(-maxAge)
	candidates, err := s.repository.FindPendingMatchesForExplanation(ctx, fromDate, toDate, limit, s.promptVersion, staleBefore)
	if err != nil {
		return GenerationSummary{}, err
	}

	summary := GenerationSummary{PromptVersion: s.promptVersion}
	s.logger.Info("ai explanation candidates loaded", "count", len(candidates), "from_date", fromDate, "to_date", toDate, "batch_size", batchSize, "min_batch_size", s.minBatchSize, "retry_missing", s.retryMissing, "max_missing_retries", s.maxMissingRetries, "stale_before", staleBefore)
	aiRequests := 0
	items := make([]explanationWorkItem, 0, len(candidates))
	for _, candidate := range candidates {
		summary.Processed++
		input, skipReason, err := BuildPromptInput(candidate, s.promptVersion)
		if err != nil {
			return summary, err
		}
		if skipReason != "" {
			if markErr := s.repository.MarkFailed(ctx, candidate, s.promptVersion, s.modelName, json.RawMessage(`{}`), json.RawMessage(`null`), skipReason); markErr != nil {
				return summary, markErr
			}
			s.logger.Warn("prediction explanation candidate failed validation", "match_id", stringValue(candidate.MatchID), "reason", skipReason)
			summary.Failed++
			continue
		}
		items = append(items, explanationWorkItem{
			candidate:  candidate,
			input:      input,
			matchID:    stringValue(input.Match.MatchID),
			retryCount: candidate.ExistingRetryCount,
		})
	}

	currentBatchSize := batchSize
	for start := 0; start < len(items); {
		end := start + currentBatchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[start:end]
		start = end

		inputs := make([]ai.ExplanationPromptInput, 0, len(batch))
		inputByMatchID := map[string]ai.ExplanationPromptInput{}
		candidateByMatchID := map[string]models.ExplanationCandidate{}
		retryByMatchID := map[string]int{}
		for _, item := range batch {
			inputs = append(inputs, item.input)
			inputByMatchID[item.matchID] = item.input
			item.candidate.ExistingRetryCount = item.retryCount
			candidateByMatchID[item.matchID] = item.candidate
			retryByMatchID[item.matchID] = item.retryCount
		}
		inputSnapshot, err := json.Marshal(inputs)
		if err != nil {
			return summary, err
		}
		matchIDs := keys(candidateByMatchID)
		promptBytes := len(inputSnapshot)
		if prompt, err := ai.BuildBatchPredictionExplanationPrompt(inputs); err == nil {
			promptBytes = len(prompt.System) + len(prompt.User)
		}
		s.logger.Info("[AI] batch starting", "sent", len(inputs), "batch_size", currentBatchSize, "match_ids", matchIDs, "retry_counts", retryByMatchID, "prompt_bytes", promptBytes)
		for _, candidate := range candidateByMatchID {
			if err := s.repository.MarkGenerating(ctx, candidate, s.promptVersion, s.modelName, inputSnapshot); err != nil {
				return summary, err
			}
		}
		if aiRequests > 0 && s.requestDelay > 0 {
			if err := sleep(ctx, s.requestDelay); err != nil {
				return summary, err
			}
		}
		aiRequests++

		var response *ai.BatchExplanationAIResponse
		requestStarted := time.Now()
		for {
			response, err = s.aiClient.GeneratePredictionExplanations(ctx, inputs)
			if err == nil {
				break
			}

			var rateLimit ai.RateLimitError
			if errors.As(err, &rateLimit) {
				summary.RateLimited = true
				if summary.RateLimitWaits >= s.maxRateLimitWaits || s.rateLimitCooldown <= 0 {
					s.logger.Warn("[AI] batch skipped after gemini rate limit", "sent", len(inputs), "match_ids", matchIDs, "retry_after", rateLimit.RetryAfter, "error", err)
					for _, candidate := range candidateByMatchID {
						if markErr := s.repository.MarkFailed(ctx, candidate, s.promptVersion, s.modelName, inputSnapshot, json.RawMessage(`null`), err.Error()); markErr != nil {
							return summary, markErr
						}
						summary.Failed++
					}
					response = nil
					break
				}

				delay := s.rateLimitCooldown
				if rateLimit.RetryAfter > delay {
					delay = rateLimit.RetryAfter
				}
				summary.RateLimitWaits++
				s.logger.Warn("[AI] batch waiting after gemini rate limit", "sent", len(inputs), "match_ids", matchIDs, "delay", delay, "waits", summary.RateLimitWaits, "error", err)
				if err := sleep(ctx, delay); err != nil {
					return summary, err
				}
				continue
			}

			message := err.Error()
			var invalidResponse ai.InvalidResponseError
			var rawResponse json.RawMessage
			if errors.As(err, &invalidResponse) {
				rawResponse = invalidResponse.RawResponse
			}
			requeued := []string{}
			failed := []string{}
			for _, item := range batch {
				retryCount := item.retryCount + 1
				candidate := candidateByMatchID[item.matchID]
				if s.retryMissing && retryCount <= s.maxMissingRetries {
					item.retryCount = retryCount
					candidate.ExistingRetryCount = retryCount
					items = append(items, item)
					requeued = append(requeued, item.matchID)
					continue
				}
				if retryCount > s.maxMissingRetries {
					candidate.ExistingRetryCount = s.maxMissingRetries
				} else {
					candidate.ExistingRetryCount = retryCount
				}
				if markErr := s.repository.MarkFailed(ctx, candidate, s.promptVersion, s.modelName, inputSnapshot, rawResponse, message); markErr != nil {
					return summary, markErr
				}
				summary.Failed++
				failed = append(failed, item.matchID)
			}
			if len(requeued) > 0 && currentBatchSize > s.minBatchSize {
				currentBatchSize = currentBatchSize / 2
				if currentBatchSize < s.minBatchSize {
					currentBatchSize = s.minBatchSize
				}
			}
			s.logger.Warn("[AI] batch failed", "sent", len(inputs), "match_ids", matchIDs, "requeued_match_ids", requeued, "failed_match_ids", failed, "fallback_batch", currentBatchSize, "request_duration", time.Since(requestStarted), "prompt_bytes", promptBytes, "error", err)
			break
		}
		if response == nil {
			continue
		}
		rawResponse := response.RawResponse
		if len(rawResponse) == 0 {
			rawResponse, _ = json.Marshal(response)
		}
		responseBytes := len(rawResponse)
		returned := map[string]struct{}{}
		for _, prediction := range response.Predictions {
			matchID := prediction.MatchID
			candidate, ok := candidateByMatchID[matchID]
			if !ok {
				s.logger.Warn("prediction explanation ignored unexpected match_id", "match_id", matchID)
				continue
			}
			input := inputByMatchID[matchID]
			riskAlert := riskAlertForLevel(prediction.RiskLevel)
			_, err = s.repository.UpsertExplanation(ctx, models.UpsertExplanationParams{
				MatchID:           candidate.MatchID,
				MatchPredictionID: candidate.MatchPredictionID,
				GoalPredictionID:  candidate.GoalPredictionID,
				HomeTeamID:        candidate.HomeTeamID,
				AwayTeamID:        candidate.AwayTeamID,
				MatchDate:         candidate.MatchDate,
				Summary:           prediction.Explanation,
				MainReasons:       prediction.KeyFactors,
				RiskAlert:         riskAlert,
				BetStyle:          ai.BetStyleForMaxProbability(maxResultProbability(input)),
				UserTip:           "Use como leitura de tendencia do modelo, sem promessa de acerto.",
				ModelName:         s.modelName,
				PromptVersion:     s.promptVersion,
				InputSnapshot:     inputSnapshot,
				RawResponse:       rawResponse,
				Status:            "generated",
				RetryCount:        0,
			})
			if err != nil {
				return summary, err
			}
			returned[matchID] = struct{}{}
			summary.Generated++
		}
		missing := missingKeys(candidateByMatchID, returned)
		for _, matchID := range missing {
			retryCount := retryByMatchID[matchID] + 1
			candidate := candidateByMatchID[matchID]
			if !s.retryMissing || retryCount > s.maxMissingRetries {
				if retryCount > s.maxMissingRetries {
					candidate.ExistingRetryCount = s.maxMissingRetries
				} else {
					candidate.ExistingRetryCount = retryCount
				}
				message := "gemini batch response did not include match_id " + matchID + " after exceeding max missing retries " + strconv.Itoa(s.maxMissingRetries)
				if err := s.repository.MarkFailed(ctx, candidate, s.promptVersion, s.modelName, inputSnapshot, rawResponse, message); err != nil {
					return summary, err
				}
				summary.Failed++
				continue
			}
			candidate.ExistingRetryCount = retryCount
			for _, item := range batch {
				if item.matchID == matchID {
					item.retryCount = retryCount
					items = append(items, item)
					break
				}
			}
		}
		fallbackBatchSize := currentBatchSize
		if len(missing) > 0 && currentBatchSize > s.minBatchSize {
			fallbackBatchSize = currentBatchSize / 2
			if fallbackBatchSize < s.minBatchSize {
				fallbackBatchSize = s.minBatchSize
			}
			currentBatchSize = fallbackBatchSize
		}
		s.logger.Info("[AI] batch finished", "sent", len(inputs), "returned", len(returned), "missing", len(missing), "saved", len(returned), "sent_match_ids", matchIDs, "returned_match_ids", keysFromSet(returned), "missing_match_ids", missing, "fallback_batch", fallbackBatchSize, "request_duration", time.Since(requestStarted), "prompt_bytes", promptBytes, "response_bytes", responseBytes)
	}
	return summary, nil
}

type explanationWorkItem struct {
	candidate  models.ExplanationCandidate
	input      ai.ExplanationPromptInput
	matchID    string
	retryCount int
}

func sleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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

func maxResultProbability(input ai.ExplanationPromptInput) float64 {
	maximum := input.ResultPrediction.HomeWinProbability
	if input.ResultPrediction.DrawProbability > maximum {
		maximum = input.ResultPrediction.DrawProbability
	}
	if input.ResultPrediction.AwayWinProbability > maximum {
		maximum = input.ResultPrediction.AwayWinProbability
	}
	return maximum
}

func riskAlertForLevel(level string) *string {
	switch level {
	case "high":
		value := "Risco alto: a previsao tem incerteza relevante."
		return &value
	case "medium":
		value := "Risco medio: interprete a previsao com cautela."
		return &value
	default:
		return nil
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func keys(values map[string]models.ExplanationCandidate) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return result
}

func keysFromSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return result
}

func missingKeys(sent map[string]models.ExplanationCandidate, returned map[string]struct{}) []string {
	result := []string{}
	for matchID := range sent {
		if _, ok := returned[matchID]; !ok {
			result = append(result, matchID)
		}
	}
	return result
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
