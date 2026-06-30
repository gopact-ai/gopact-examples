package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRunStreamsBothOpenAIAPIs(t *testing.T) {
	var mu sync.Mutex
	seen := map[string]bool{}
	tokenBudgets := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		var body struct {
			MaxTokens       int `json:"max_tokens"`
			MaxOutputTokens int `json:"max_output_tokens"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		mu.Lock()
		seen[r.URL.Path] = true
		if body.MaxOutputTokens != 0 {
			tokenBudgets[r.URL.Path] = body.MaxOutputTokens
		} else {
			tokenBudgets[r.URL.Path] = body.MaxTokens
		}
		mu.Unlock()

		w.Header().Set("Content-Type", "text/event-stream")
		switch r.URL.Path {
		case "/chat/completions":
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello from chat\"}}]}\n\n"))
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
		case "/responses":
			_, _ = w.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"hello from responses\"}\n\n"))
			_, _ = w.Write([]byte("data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":1,\"output_tokens\":2,\"total_tokens\":3}}}\n\n"))
		default:
			t.Fatalf("path = %q, want openai stream path", r.URL.Path)
		}
	}))
	defer server.Close()

	t.Setenv("GOPACT_LLM_BASEURL", server.URL)
	t.Setenv("GOPACT_LLM_TOKEN", "test-token")
	t.Setenv("GOPACT_LLM_MODEL", "test-model")

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !seen["/chat/completions"] || !seen["/responses"] {
		t.Fatalf("seen paths = %#v, want both openai APIs", seen)
	}
	if tokenBudgets["/chat/completions"] != openAIStreamingMaxOutputTokens ||
		tokenBudgets["/responses"] != openAIStreamingMaxOutputTokens {
		t.Fatalf("token budgets = %#v, want %d for both APIs", tokenBudgets, openAIStreamingMaxOutputTokens)
	}
	if got := out.String(); !strings.Contains(got, "chat_completions: hello from chat") || !strings.Contains(got, "responses: hello from responses") {
		t.Fatalf("output = %q, want both stream outputs", got)
	}
}
