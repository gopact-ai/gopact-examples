package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunExecutesToolCall(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			_, _ = w.Write([]byte(`{
				"choices": [{"message": {"role": "assistant", "tool_calls": [{
					"id": "call_1",
					"type": "function",
					"function": {"name": "uppercase", "arguments": "{\"text\":\"gopact\"}"}
				}]}}]
			}`))
			return
		}

		var body struct {
			Messages []struct {
				Role       string `json:"role"`
				ToolCallID string `json:"tool_call_id"`
				Content    string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("Decode() error = %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if len(body.Messages) == 0 || body.Messages[len(body.Messages)-1].Content != "GOPACT" {
			t.Errorf("last tool message = %#v, want GOPACT result", body.Messages)
		}
		_, _ = w.Write([]byte(`{
			"choices": [{"message": {"role": "assistant", "content": "tool result was GOPACT"}}]
		}`))
	}))
	defer server.Close()

	t.Setenv("GOPACT_LLM_BASEURL", server.URL)
	t.Setenv("GOPACT_LLM_TOKEN", "test-token")
	t.Setenv("GOPACT_LLM_MODEL", "test-model")

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(out.String(), "tool result was GOPACT") {
		t.Fatalf("output = %q, want final tool response", out.String())
	}
}
