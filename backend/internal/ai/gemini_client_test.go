package ai

import "testing"

func TestExtractGeminiOutputText(t *testing.T) {
	body := []byte(`{
		"candidates": [
			{
				"content": {
					"parts": [
						{"text": "{\"summary\":\"Brasil aparece como leve favorito.\"}"}
					]
				}
			}
		]
	}`)

	output, err := extractGeminiOutputText(body)
	if err != nil {
		t.Fatalf("extractGeminiOutputText() error = %v", err)
	}
	if string(output) != `{"summary":"Brasil aparece como leve favorito."}` {
		t.Fatalf("output = %s", output)
	}
}

func TestExtractGeminiOutputTextRejectsEmptyResponse(t *testing.T) {
	if _, err := extractGeminiOutputText([]byte(`{"candidates":[]}`)); err == nil {
		t.Fatalf("expected missing output text error")
	}
}

func TestRetryDelayFromGeminiErrorMessage(t *testing.T) {
	body := []byte(`{"error":{"code":429,"message":"You exceeded your current quota, limit: 5, model: gemini-2.5-flash. Please retry in 14.976511193s.","status":"RESOURCE_EXHAUSTED"}}`)

	delay := retryDelayFromResponse("", body)
	if delay <= 14_000_000_000 || delay >= 15_000_000_000 {
		t.Fatalf("delay = %s", delay)
	}
}
