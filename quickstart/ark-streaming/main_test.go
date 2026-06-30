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

func TestRunStreamsConfiguredArk(t *testing.T) {
	var requests []struct {
		MaxOutputTokens int `json:"max_output_tokens"`
		Thinking        struct {
			Type string `json:"type"`
		} `json:"thinking"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			MaxOutputTokens int `json:"max_output_tokens"`
			Thinking        struct {
				Type string `json:"type"`
			} `json:"thinking"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		requests = append(requests, request)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"one\"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\", two\"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\", three\"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":1,\"output_tokens\":2,\"total_tokens\":3}}}\n\n"))
	}))
	defer server.Close()

	t.Setenv("GOPACT_LLM_BASEURL", server.URL)
	t.Setenv("GOPACT_LLM_TOKEN", "test-token")
	t.Setenv("GOPACT_LLM_MODEL", "ep-test")

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if strings.Count(out.String(), "one, two, three") != 2 {
		t.Fatalf("output = %q, want streamed text", out.String())
	}
	if len(requests) != 2 {
		t.Fatalf("requests = %d, want disabled and enabled calls", len(requests))
	}
	if requests[0].Thinking.Type != "disabled" || requests[1].Thinking.Type != "enabled" {
		t.Fatalf("thinking = %q/%q, want disabled/enabled", requests[0].Thinking.Type, requests[1].Thinking.Type)
	}
	if requests[0].MaxOutputTokens != arkStreamingMaxOutputTokens || requests[1].MaxOutputTokens != arkStreamingMaxOutputTokens {
		t.Fatalf("max_output_tokens = %d/%d, want %d", requests[0].MaxOutputTokens, requests[1].MaxOutputTokens, arkStreamingMaxOutputTokens)
	}
}
