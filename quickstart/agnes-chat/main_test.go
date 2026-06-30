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

func TestRunUsesConfiguredAgnesLLM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %q, want Agnes chat completions path", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		_, _ = w.Write([]byte(`{
			"choices": [{"message": {"role": "assistant", "content": "hello from fake agnes"}}]
		}`))
	}))
	defer server.Close()

	t.Setenv("GOPACT_LLM_BASEURL", server.URL)
	t.Setenv("GOPACT_LLM_TOKEN", "test-token")
	t.Setenv("GOPACT_LLM_MODEL", "agnes-test")

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(out.String(), "hello from fake agnes") {
		t.Fatalf("output = %q, want fake Agnes response", out.String())
	}
}

func TestRunUsesAgnesSpecificConfigBeforeSharedLLM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer agnes-token" {
			t.Fatalf("Authorization = %q, want Agnes-specific bearer token", got)
		}
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if body.Model != "agnes-model" {
			t.Fatalf("model = %q, want Agnes-specific model", body.Model)
		}
		_, _ = w.Write([]byte(`{
			"choices": [{"message": {"role": "assistant", "content": "hello from agnes-specific config"}}]
		}`))
	}))
	defer server.Close()

	t.Setenv("GOPACT_LLM_BASEURL", "https://ark.example.test/api/v3")
	t.Setenv("GOPACT_LLM_TOKEN", "shared-token")
	t.Setenv("GOPACT_LLM_MODEL", "shared-model")
	t.Setenv("GOPACT_AGNES_BASEURL", server.URL)
	t.Setenv("GOPACT_AGNES_API_KEY", "agnes-token")
	t.Setenv("GOPACT_AGNES_MODEL", "agnes-model")

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(out.String(), "hello from agnes-specific config") {
		t.Fatalf("output = %q, want fake Agnes response", out.String())
	}
}
