package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunUsesConfiguredLLM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		_, _ = w.Write([]byte(`{
			"choices": [{"message": {"role": "assistant", "content": "hello from fake llm"}}]
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
	if !strings.Contains(out.String(), "hello from fake llm") {
		t.Fatalf("output = %q, want fake response", out.String())
	}
}
