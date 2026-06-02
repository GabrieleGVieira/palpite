package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResendSenderPostsEmailPayload(t *testing.T) {
	var payload resendEmailRequest
	var authHeader string
	var userAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		userAgent = r.Header.Get("User-Agent")
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := ResendSender{
		apiKey:     "re_test",
		from:       resendDefaultFrom,
		httpClient: server.Client(),
		apiURL:     server.URL,
	}

	err := sender.Send(context.Background(), Message{
		To:      []string{"admin@example.com"},
		Subject: "Subject",
		HTML:    "<p>Hello</p>",
		Text:    "Hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if authHeader != "Bearer re_test" {
		t.Fatalf("unexpected authorization header: %q", authHeader)
	}
	if userAgent != resendUserAgent {
		t.Fatalf("unexpected user agent: %q", userAgent)
	}
	if payload.From != resendDefaultFrom ||
		len(payload.To) != 1 || payload.To[0] != "admin@example.com" ||
		payload.Subject != "Subject" ||
		payload.HTML != "<p>Hello</p>" ||
		payload.Text != "Hello" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestResendSenderReturnsErrorWhenAPIRejectsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "invalid api key", http.StatusUnauthorized)
	}))
	defer server.Close()

	sender := ResendSender{
		apiKey:     "re_test",
		from:       resendDefaultFrom,
		httpClient: &http.Client{Timeout: time.Second},
		apiURL:     server.URL,
	}

	err := sender.Send(context.Background(), Message{
		To:      []string{"admin@example.com"},
		Subject: "Subject",
		Text:    "Hello",
	})
	if err == nil {
		t.Fatal("expected resend api error")
	}
}
