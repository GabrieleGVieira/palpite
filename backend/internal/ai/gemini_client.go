package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaultGeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"

type GeminiClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("gemini rate limit: retry after %s: %s", e.RetryAfter, e.Message)
	}
	return "gemini rate limit: " + e.Message
}

func NewGeminiClient(apiKey string, model string, timeout time.Duration) (*GeminiClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("GEMINI_API_KEY is required")
	}
	if strings.TrimSpace(model) == "" {
		model = "gemini-2.5-flash"
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &GeminiClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultGeminiBaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: 2,
	}, nil
}

func (c *GeminiClient) GeneratePredictionExplanation(ctx context.Context, input ExplanationPromptInput) (*ExplanationAIResponse, error) {
	prompt, err := BuildPredictionExplanationPrompt(input)
	if err != nil {
		return nil, err
	}
	raw, err := c.generateContent(ctx, prompt.User, prompt.System, GeminiExplanationSchema())
	if err != nil {
		return nil, err
	}
	parsed, validationErr := ParseAndValidateExplanation(raw, input)
	if validationErr == nil {
		parsed.RawResponse = raw
		return parsed, nil
	}

	correctionPrompt := prompt.User + "\n\nA resposta anterior foi invalida pelo seguinte motivo: " + validationErr.Error() + ". Retorne novamente apenas JSON valido seguindo o schema."
	raw, err = c.generateContent(ctx, correctionPrompt, prompt.System, GeminiExplanationSchema())
	if err != nil {
		return nil, err
	}
	parsed, validationErr = ParseAndValidateExplanation(raw, input)
	if validationErr != nil {
		return nil, InvalidResponseError{RawResponse: raw, Err: validationErr}
	}
	parsed.RawResponse = raw
	return parsed, nil
}

func (c *GeminiClient) GeneratePredictionExplanations(ctx context.Context, inputs []ExplanationPromptInput) (*BatchExplanationAIResponse, error) {
	prompt, err := BuildBatchPredictionExplanationPrompt(inputs)
	if err != nil {
		return nil, err
	}
	raw, err := c.generateContent(ctx, prompt.User, prompt.System, GeminiBatchExplanationSchema())
	if err != nil {
		return nil, err
	}
	parsed, validationErr := ParseAndValidateBatchExplanations(raw, inputs)
	if validationErr == nil {
		parsed.RawResponse = raw
		return parsed, nil
	}

	correctionPrompt := prompt.User + "\n\nA resposta anterior foi invalida pelo seguinte motivo: " + validationErr.Error() + ". Retorne novamente apenas JSON valido seguindo o schema. O array predictions deve ter exatamente " + strconv.Itoa(len(inputs)) + " itens, um para cada match_id enviado, sem omitir partidas."
	raw, err = c.generateContent(ctx, correctionPrompt, prompt.System, GeminiBatchExplanationSchema())
	if err != nil {
		return nil, err
	}
	parsed, validationErr = ParseAndValidateBatchExplanations(raw, inputs)
	if validationErr != nil {
		return nil, InvalidResponseError{RawResponse: raw, Err: validationErr}
	}
	parsed.RawResponse = raw
	return parsed, nil
}

type InvalidResponseError struct {
	RawResponse []byte
	Err         error
}

func (e InvalidResponseError) Error() string {
	return e.Err.Error()
}

func (e InvalidResponseError) Unwrap() error {
	return e.Err
}

func (c *GeminiClient) generateContent(ctx context.Context, userPrompt string, systemPrompt string, responseSchema map[string]any) ([]byte, error) {
	requestBody := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": userPrompt},
				},
			},
		},
		"systemInstruction": map[string]any{
			"parts": []map[string]string{
				{"text": systemPrompt},
			},
		},
		"generationConfig": map[string]any{
			"responseMimeType": "application/json",
			"responseSchema":   responseSchema,
		},
	}
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			timer := time.NewTimer(time.Duration(attempt) * 500 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/models/"+c.model+":generateContent", bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("x-goog-api-key", c.apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return extractGeminiOutputText(body)
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := retryDelayFromResponse(resp.Header.Get("Retry-After"), body)
			lastErr = RateLimitError{RetryAfter: retryAfter, Message: truncate(string(body), 500)}
			if retryAfter > 0 {
				if err := sleepWithContext(ctx, retryAfter); err != nil {
					return nil, err
				}
			}
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("gemini transient error status=%d body=%s", resp.StatusCode, truncate(string(body), 500))
			continue
		}
		return nil, fmt.Errorf("gemini request failed status=%d body=%s", resp.StatusCode, truncate(string(body), 500))
	}
	return nil, lastErr
}

func retryDelayFromResponse(retryAfterHeader string, body []byte) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfterHeader)); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	var response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &response)
	message := response.Error.Message
	if strings.TrimSpace(message) == "" {
		message = string(body)
	}
	match := regexp.MustCompile(`retry in ([0-9]+(?:\.[0-9]+)?)s`).FindStringSubmatch(message)
	if len(match) != 2 {
		return 0
	}
	seconds, err := strconv.ParseFloat(match[1], 64)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds*1000) * time.Millisecond
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func extractGeminiOutputText(body []byte) ([]byte, error) {
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			if strings.TrimSpace(part.Text) != "" {
				return []byte(part.Text), nil
			}
		}
	}
	return nil, errors.New("gemini response did not include output text")
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
