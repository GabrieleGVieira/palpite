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
