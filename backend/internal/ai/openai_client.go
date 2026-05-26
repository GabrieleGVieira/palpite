package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

type OpenAIClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

func NewOpenAIClient(apiKey string, model string, timeout time.Duration) (*OpenAIClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("OPENAI_API_KEY is required")
	}
	if strings.TrimSpace(model) == "" {
		model = "gpt-4.1-mini"
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &OpenAIClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultOpenAIBaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: 2,
	}, nil
}

func (c *OpenAIClient) GeneratePredictionExplanation(ctx context.Context, input ExplanationPromptInput) (*ExplanationAIResponse, error) {
	prompt, err := BuildPredictionExplanationPrompt(input)
	if err != nil {
		return nil, err
	}
	raw, err := c.createResponse(ctx, prompt.User, prompt.System)
	if err != nil {
		return nil, err
	}
	parsed, validationErr := ParseAndValidateExplanation(raw, input)
	if validationErr == nil {
		parsed.RawResponse = raw
		return parsed, nil
	}

	correctionPrompt := prompt.User + "\n\nA resposta anterior foi invalida pelo seguinte motivo: " + validationErr.Error() + ". Retorne novamente apenas JSON valido seguindo o schema."
	raw, err = c.createResponse(ctx, correctionPrompt, prompt.System)
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

func (c *OpenAIClient) createResponse(ctx context.Context, userPrompt string, systemPrompt string) ([]byte, error) {
	requestBody := map[string]any{
		"model": c.model,
		"input": []map[string]any{
			{
				"role": "system",
				"content": []map[string]string{
					{"type": "input_text", "text": systemPrompt},
				},
			},
			{
				"role": "user",
				"content": []map[string]string{
					{"type": "input_text", "text": userPrompt},
				},
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":        "json_schema",
				"name":        "prediction_explanation",
				"description": "Short Brazilian Portuguese explanation for a precomputed football match prediction.",
				"strict":      true,
				"schema":      ExplanationJSONSchema(),
			},
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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
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
			return extractOutputText(body)
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("openai transient error status=%d body=%s", resp.StatusCode, truncate(string(body), 500))
			continue
		}
		return nil, fmt.Errorf("openai request failed status=%d body=%s", resp.StatusCode, truncate(string(body), 500))
	}
	return nil, lastErr
}

func extractOutputText(body []byte) ([]byte, error) {
	var response struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if strings.TrimSpace(response.OutputText) != "" {
		return []byte(response.OutputText), nil
	}
	for _, item := range response.Output {
		for _, content := range item.Content {
			if strings.TrimSpace(content.Text) != "" {
				return []byte(content.Text), nil
			}
		}
	}
	return nil, errors.New("openai response did not include output text")
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
