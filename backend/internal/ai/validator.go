package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var forbiddenTerms = []string{"garantido", "certeza", "aposta segura", "lucro"}

func ParseAndValidateExplanation(raw []byte, input ExplanationPromptInput) (*ExplanationAIResponse, error) {
	var response ExplanationAIResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	if err := ValidateExplanationResponse(response, input); err != nil {
		return nil, err
	}
	return &response, nil
}

func ParseAndValidateBatchExplanations(raw []byte, inputs []ExplanationPromptInput) (*BatchExplanationAIResponse, error) {
	var response BatchExplanationAIResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	if response.Predictions == nil {
		return nil, errors.New("predictions is required")
	}
	if len(response.Predictions) == 0 {
		return nil, errors.New("predictions cannot be empty")
	}

	sent := map[string]struct{}{}
	for _, input := range inputs {
		if input.Match.MatchID != nil && strings.TrimSpace(*input.Match.MatchID) != "" {
			sent[*input.Match.MatchID] = struct{}{}
		}
	}
	if len(sent) == 0 {
		return nil, errors.New("batch has no match_id values")
	}

	seen := map[string]struct{}{}
	for _, prediction := range response.Predictions {
		matchID := strings.TrimSpace(prediction.MatchID)
		if matchID == "" {
			return nil, errors.New("prediction match_id is required")
		}
		if _, ok := sent[matchID]; !ok {
			return nil, fmt.Errorf("prediction match_id %s was not sent in batch", matchID)
		}
		if _, ok := seen[matchID]; ok {
			return nil, fmt.Errorf("prediction match_id %s is duplicated", matchID)
		}
		seen[matchID] = struct{}{}
		if strings.TrimSpace(prediction.Explanation) == "" {
			return nil, fmt.Errorf("prediction %s explanation is required", matchID)
		}
		if len(prediction.KeyFactors) < 2 || len(prediction.KeyFactors) > 4 {
			return nil, fmt.Errorf("prediction %s key_factors must contain between 2 and 4 items", matchID)
		}
		for _, factor := range prediction.KeyFactors {
			if strings.TrimSpace(factor) == "" {
				return nil, fmt.Errorf("prediction %s key_factors cannot contain empty items", matchID)
			}
		}
		if !validRiskLevel(prediction.RiskLevel) {
			return nil, fmt.Errorf("prediction %s risk_level is invalid", matchID)
		}
		if containsForbiddenBatchTerm(prediction) {
			return nil, fmt.Errorf("prediction %s contains forbidden betting certainty language", matchID)
		}
	}
	return &response, nil
}

func ValidateExplanationResponse(response ExplanationAIResponse, input ExplanationPromptInput) error {
	if strings.TrimSpace(response.Summary) == "" {
		return errors.New("summary is required")
	}
	if len([]rune(response.Summary)) > 240 {
		return errors.New("summary exceeds max length")
	}
	if len(response.MainReasons) < 2 || len(response.MainReasons) > 4 {
		return errors.New("main_reasons must contain between 2 and 4 items")
	}
	for _, reason := range response.MainReasons {
		if strings.TrimSpace(reason) == "" {
			return errors.New("main_reasons cannot contain empty items")
		}
	}
	if !validBetStyle(response.BetStyle) {
		return errors.New("bet_style is invalid")
	}
	expected := BetStyleForMaxProbability(maxResultProbability(input))
	if response.BetStyle != expected {
		return fmt.Errorf("bet_style must be %s for supplied probabilities", expected)
	}
	if strings.TrimSpace(response.UserTip) == "" {
		return errors.New("user_tip is required")
	}
	if input.ResultPrediction.Confidence == "low" && (response.RiskAlert == nil || strings.TrimSpace(*response.RiskAlert) == "") {
		return errors.New("risk_alert is required when confidence is low")
	}
	if containsForbiddenTerm(response) {
		return errors.New("response contains forbidden betting certainty language")
	}
	return nil
}

func BetStyleForMaxProbability(maxProbability float64) string {
	if maxProbability >= 60 {
		return "conservative"
	}
	if maxProbability >= 45 {
		return "moderate"
	}
	return "risky"
}

func maxResultProbability(input ExplanationPromptInput) float64 {
	maximum := input.ResultPrediction.HomeWinProbability
	if input.ResultPrediction.DrawProbability > maximum {
		maximum = input.ResultPrediction.DrawProbability
	}
	if input.ResultPrediction.AwayWinProbability > maximum {
		maximum = input.ResultPrediction.AwayWinProbability
	}
	return maximum
}

func validBetStyle(value string) bool {
	return value == "conservative" || value == "moderate" || value == "risky"
}

func validRiskLevel(value string) bool {
	return value == "low" || value == "medium" || value == "high"
}

func containsForbiddenTerm(response ExplanationAIResponse) bool {
	values := []string{response.Summary, response.UserTip, response.BetStyle}
	if response.RiskAlert != nil {
		values = append(values, *response.RiskAlert)
	}
	values = append(values, response.MainReasons...)
	text := strings.ToLower(strings.Join(values, " "))
	for _, term := range forbiddenTerms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}

func containsForbiddenBatchTerm(response BatchPredictionExplanation) bool {
	values := []string{response.Explanation, response.RiskLevel}
	values = append(values, response.KeyFactors...)
	text := strings.ToLower(strings.Join(values, " "))
	for _, term := range forbiddenTerms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}
