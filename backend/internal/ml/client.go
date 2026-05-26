package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) Client {
	return Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c Client) Predict(ctx context.Context, request PredictRequest) (PredictResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return PredictResponse{}, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/predict", bytes.NewReader(body))
	if err != nil {
		return PredictResponse{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return PredictResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return PredictResponse{}, fmt.Errorf("ml-service returned status %d", response.StatusCode)
	}

	var prediction PredictResponse
	if err := json.NewDecoder(response.Body).Decode(&prediction); err != nil {
		return PredictResponse{}, err
	}
	return prediction, nil
}
